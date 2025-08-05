package clamav

// Response represents ClamAV daemon responses to commands over a TCP connection.
type Response []byte

var (
	// RespPing is the expected response to a PING command
	RespPing Response = []byte("PONG")
	// RespReload is the expected response to a RELOAD command
	RespReload Response = []byte("RELOADING")
	// RespScan is the expected response for a clean file scan
	RespScan Response = []byte("stream: OK")
	// RespErrUnknownCommand indicates an unknown command was sent
	RespErrUnknownCommand Response = []byte("UNKNOWN COMMAND")
	// RespErrScanFileSizeLimitExceeded indicates file size limit was exceeded
	RespErrScanFileSizeLimitExceeded Response = []byte("INSTREAM size limit exceeded. ERROR")
)
