package clamav

import "errors"

var (
	// ErrUnknownCommand indicates an unrecognized command was sent to ClamAV
	ErrUnknownCommand = errors.New("unknown command")
	// ErrUnknownResponse indicates ClamAV returned an unrecognized response
	ErrUnknownResponse = errors.New("unknown response from clamav")
	// ErrUnexpectedResponse indicates ClamAV returned an unexpected response format
	ErrUnexpectedResponse = errors.New("unexpected response from clamav")
	// ErrScanFileSizeLimitExceeded indicates the file size exceeds ClamAV's limit
	ErrScanFileSizeLimitExceeded = errors.New("size limit exceeded")
	// ErrVirusFound indicates a virus was detected in the scanned content
	ErrVirusFound = errors.New("file contains potential virus")
)
