package clamav

// Command represents ClamAV daemon commands over a TCP connection.
//
// It's recommended to prefix clamd commands with the letter z (eg. zSCAN)
// to indicate that the command will be delimited by a NULL character and
// that clamd should continue reading command data until a NULL character is read.
// The null delimiter assures that the complete command and its entire argument
// will be processed as a single command. Alternatively commands may be prefixed
// with the letter n (e.g. nSCAN) to use a newline character as the delimiter.
// Clamd replies will honour the requested terminator in turn. If clamd doesn't
// recognize the command, or the command doesn't follow the requirements specified below,
// it will reply with an error message, and close the connection.
//
// More information on clamd(8)
type Command []byte

var (
	// CmdPing sends a ping command to test connectivity
	CmdPing Command = []byte("zPING\000")
	// CmdVersion requests the ClamAV version information
	CmdVersion Command = []byte("zVERSION\000")
	// CmdReload instructs the daemon to reload its configuration
	CmdReload Command = []byte("zRELOAD\000")
	// CmdInstream begins an INSTREAM scan session
	CmdInstream Command = []byte("zINSTREAM\000")
	// CmdStats requests daemon statistics
	CmdStats Command = []byte("zSTATS\000")
	// CmdVersionCommands requests the list of available commands
	CmdVersionCommands Command = []byte("nVERSIONCOMMANDS\n") // From https://linux.die.net/man/8/clamd, it is recommended to use nVERSIONCOMMANDS.
	// CmdShutdown instructs the daemon to shutdown gracefully
	CmdShutdown Command = []byte("zSHUTDOWN\000")
)
