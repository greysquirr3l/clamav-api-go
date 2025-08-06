// Package clamav provides a client interface for communicating with ClamAV antivirus daemon.
// It implements the ClamAV TCP protocol for various operations including virus scanning,
// status checks, and daemon management.
package clamav

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"os/exec"
	"strings"
	"time"
)

// Clamaver defines the interface for ClamAV operations.
// All methods accept a context for cancellation and timeout handling.
type Clamaver interface {
	Ping(ctx context.Context) ([]byte, error)
	Version(ctx context.Context) ([]byte, error)
	Reload(ctx context.Context) error
	Stats(ctx context.Context) ([]byte, error)
	VersionCommands(ctx context.Context) ([]byte, error)
	Shutdown(ctx context.Context) error
	InStream(ctx context.Context, r io.Reader, size int64) ([]byte, error)
	FreshClam(ctx context.Context) ([]byte, error)
}

// Client implements the Clamaver interface and provides
// TCP-based communication with a ClamAV daemon.
type Client struct {
	dialer  net.Dialer
	address string
	network string
}

var _ Clamaver = (*Client)(nil)

// NewClamavClient creates a new ClamAV client with the specified network parameters.
// addr is the ClamAV daemon address, netw is the network type (usually "tcp"),
// timeout is the connection timeout, and keepalive is the keep-alive duration.
func NewClamavClient(addr string, netw string, timeout time.Duration, keepalive time.Duration) *Client {
	return &Client{
		dialer: net.Dialer{
			Timeout:   timeout,
			KeepAlive: keepalive,
		},
		address: addr,
		network: netw,
	}
}

// Ping sends a PING command to the ClamAV daemon to test connectivity.
func (c *Client) Ping(ctx context.Context) ([]byte, error) {
	conn, err := c.dialer.DialContext(ctx, c.network, c.address)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to clamav: %w", err)
	}
	defer func() { _ = conn.Close() }()

	resp, err := c.SendCommand(conn, CmdPing)
	if err != nil {
		return nil, fmt.Errorf("error while sending command: %w", err)
	}

	err = c.parseResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("error from clamav: %w", err)
	}
	return resp, nil
}

// Version gets the ClamAV daemon version information.
func (c *Client) Version(ctx context.Context) ([]byte, error) {
	conn, err := c.dialer.DialContext(ctx, c.network, c.address)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to clamav: %w", err)
	}
	defer func() { _ = conn.Close() }()

	resp, err := c.SendCommand(conn, CmdVersion)
	if err != nil {
		return nil, fmt.Errorf("error while sending command: %w", err)
	}

	err = c.parseResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("error from clamav: %w", err)
	}
	return resp, nil
}

// Reload instructs the ClamAV daemon to reload its configuration and virus databases.
func (c *Client) Reload(ctx context.Context) error {
	conn, err := c.dialer.DialContext(ctx, c.network, c.address)
	if err != nil {
		return fmt.Errorf("failed to connect to clamav: %w", err)
	}
	defer func() { _ = conn.Close() }()

	resp, err := c.SendCommand(conn, CmdReload)
	if err != nil {
		return fmt.Errorf("error while sending command: %w", err)
	}

	err = c.parseResponse(resp)
	if err != nil {
		return fmt.Errorf("error from clamav: %w", err)
	}

	if !bytes.Equal(resp, RespReload) {
		return fmt.Errorf("error from clamav: %w. Expected %s but got %s", ErrUnexpectedResponse, RespReload, resp)
	}
	return nil
}

// Stats retrieves statistics from the ClamAV daemon.
func (c *Client) Stats(ctx context.Context) ([]byte, error) {
	conn, err := c.dialer.DialContext(ctx, c.network, c.address)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to clamav: %w", err)
	}
	defer func() { _ = conn.Close() }()

	resp, err := c.SendCommand(conn, CmdStats)
	if err != nil {
		return nil, fmt.Errorf("error while sending command: %w", err)
	}

	err = c.parseResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("error from clamav: %w", err)
	}
	return resp, nil
}

// VersionCommands retrieves the list of available commands from the ClamAV daemon.
func (c *Client) VersionCommands(ctx context.Context) ([]byte, error) {
	conn, err := c.dialer.DialContext(ctx, c.network, c.address)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to clamav: %w", err)
	}
	defer func() { _ = conn.Close() }()

	resp, err := c.SendCommand(conn, CmdVersionCommands)
	if err != nil {
		return nil, fmt.Errorf("error while sending command: %w", err)
	}

	err = c.parseResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("error from clamav: %w", err)
	}
	return resp, nil
}

// Shutdown instructs the ClamAV daemon to shutdown gracefully.
func (c *Client) Shutdown(ctx context.Context) error {
	conn, err := c.dialer.DialContext(ctx, c.network, c.address)
	if err != nil {
		return fmt.Errorf("failed to connect to clamav: %w", err)
	}
	defer func() { _ = conn.Close() }()

	_, err = c.SendCommand(conn, CmdShutdown)
	if err != nil {
		return fmt.Errorf("error while sending command: %w", err)
	}
	return nil
}

// InStream will attempt to connect to Clamd, send the command over the network ("INSTREAM")
// and stream the given io.Reader to let Clamd scan it.
//
// The stream is sent to Clamd in chunks, after INSTREAM, on the same socket on which the command was sent.
//
// It will read the response and return it as a byte slice as well as any error
// encountered.
//
// See https://linux.die.net/man/8/clamd for a detailed explanation of the INSTREAM command.
func (c *Client) InStream(ctx context.Context, r io.Reader, size int64) ([]byte, error) {
	conn, err := c.dialer.DialContext(ctx, c.network, c.address)
	if err != nil {
		return nil, fmt.Errorf("error while dialing %s/%s: %w", c.network, c.address, err)
	}
	defer func() { _ = conn.Close() }()

	// The format of the chunk is: '<length><data>' where <length> is the size of the following data in bytes
	// expressed as a 4 byte unsigned integer in network byte order and <data> is the actual chunk.
	// Streaming is terminated by sending a zero-length chunk.

	reader := bufio.NewReaderSize(r, 2048)
	writer := bufio.NewWriter(conn)

	// Start scan command.
	_, err = writer.Write(CmdInstream)
	if err != nil {
		return nil, fmt.Errorf("error while writing command to %s/%s: %w", c.network, c.address, err)
	}
	if err = writer.Flush(); err != nil {
		return nil, fmt.Errorf("error while flushing command to %s/%s: %w", c.network, c.address, err)
	}

	// The size (referred previously as '<length>') must be a byte[] of length 4 - representing a
	// uint32 in a big-endian format (network byte order, tcp standard).
	b := make([]byte, 4)
	if size > 0 && size <= 4294967295 { // Check for valid uint32 range
		binary.BigEndian.PutUint32(b, uint32(size))
	} else {
		return nil, fmt.Errorf("file size %d exceeds maximum allowed size", size)
	}
	_, err = writer.Write(b)
	if err != nil {
		return nil, fmt.Errorf("error while writing data length to %s/%s: %w", c.network, c.address, err)
	}
	if err = writer.Flush(); err != nil {
		return nil, fmt.Errorf("error while flushing data length to %s/%s: %w", c.network, c.address, err)
	}

	// Streaming the data
	_, err = reader.WriteTo(writer)
	if err != nil {
		resp, e := c.readResponse(conn)
		if e != nil {
			return nil, fmt.Errorf("error while streaming content to %s/%s: %w", c.network, c.address, err)
		}
		err = c.parseResponse(resp)
		if err != nil {
			if errors.Is(err, ErrScanFileSizeLimitExceeded) {
				return nil, err
			}
			return nil, fmt.Errorf("error from clamav: %w", err)
		}
		return resp, fmt.Errorf("error while streaming content to %s/%s: %w", c.network, c.address, err)
	}

	// Sending 4 bytes to signal the end of the transfer.
	_, err = writer.Write([]byte{'\000', '\000', '\000', '\000'})
	if err != nil {
		return nil, fmt.Errorf("error while writing end of transfer signal to %s/%s: %w", c.network, c.address, err)
	}
	if err = writer.Flush(); err != nil {
		return nil, fmt.Errorf("error while flushing end of transfer signal to %s/%s: %w", c.network, c.address, err)
	}

	resp, err := c.readResponse(conn)
	if err != nil {
		return nil, err
	}

	err = c.parseResponse(resp)
	if err != nil {
		if errors.Is(err, ErrVirusFound) {
			return resp, err
		}
		return nil, fmt.Errorf("error from clamav: %w", err)
	}

	return resp, nil
}

// SendCommand will attempt send the given command to Clamd
// over the network.
// It will read the response and return it as a byte slice as well as any error
// encountered.
//
// See https://linux.die.net/man/8/clamd for a list of supported commands.
func (c *Client) SendCommand(conn net.Conn, cmd Command) ([]byte, error) {
	writer := bufio.NewWriter(conn)

	_, err := writer.Write(cmd)
	if err != nil {
		return nil, fmt.Errorf("error while writing command to %s/%s: %w", c.network, c.address, err)
	}
	if err = writer.Flush(); err != nil {
		return nil, fmt.Errorf("error while flushing command to %s/%s: %w", c.network, c.address, err)
	}

	resp, err := c.readResponse(conn)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// readResponse will read from the given io.Reader until a null character is found
// and returns the read bytes before the null character or any error encountered.
func (c *Client) readResponse(r io.Reader) ([]byte, error) {
	reader := bufio.NewReader(r)

	resp, err := reader.ReadBytes('\000')
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("error while reading response from %s/%s: %w", c.network, c.address, err)
	}

	// Clamd terminate the response with a NULL character (\000)
	// which can safely be trimed
	return bytes.TrimSuffix(resp, []byte("\000")), nil
}

// parseResponse will attempt to parse the Clamav response to the command
// and determine whether or not Clamav answered with an error.
// See clamav/errors.go for a list of known errors.
func (c *Client) parseResponse(msg []byte) error {
	if bytes.EqualFold(msg, RespErrScanFileSizeLimitExceeded) {
		return ErrScanFileSizeLimitExceeded
	}

	if bytes.HasPrefix(msg, []byte("stream: ")) && bytes.HasSuffix(msg, []byte("FOUND")) {
		return ErrVirusFound
	}

	if bytes.Equal(msg, RespErrUnknownCommand) {
		return ErrUnknownCommand
	}

	return nil
}

// FreshClam executes the freshclam command to update virus definitions.
// This method runs freshclam as a separate process since it's not available
// through the ClamAV TCP protocol. It captures both stdout and stderr
// and returns the combined output.
func (c *Client) FreshClam(ctx context.Context) ([]byte, error) {
	// Create the freshclam command with context for cancellation
	cmd := exec.CommandContext(ctx, "freshclam", "--verbose", "--stdout")

	// Capture both stdout and stderr for comprehensive output
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Check if it's a context cancellation
		if ctx.Err() != nil {
			return nil, fmt.Errorf("freshclam command cancelled: %w", ctx.Err())
		}

		// FreshClam may return non-zero exit codes for warnings (e.g., already up to date)
		// Check if the output contains success indicators
		outputStr := string(output)
		if strings.Contains(outputStr, "Database updated") ||
			strings.Contains(outputStr, "up to date") ||
			strings.Contains(outputStr, "Your ClamAV installation is OUTDATED") {
			// These are considered successful outcomes
			return output, nil
		}

		return output, fmt.Errorf("freshclam command failed: %w", err)
	}

	return output, nil
}
