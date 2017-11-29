package cmd

const (
	// Success is the same as EXIT_SUCCESS in C
	Success = iota

	// BadArgs passed to cli; not our fault.
	BadArgs

	// BadPassword passed to prompt or switch; not our fault.
	BadPassword

	// DaemonNotResponding means the daemon does not respond in timely fashion.
	// Probably our fault.
	DaemonNotResponding

	// UnknownError is an uncategorized error, probably our fault.
	UnknownError
)
