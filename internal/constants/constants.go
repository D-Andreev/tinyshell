package constants

var SHELL_TRAILING_SYMBOL = "$"
var NOOP = func() {}
var EXECUTE_PERMISSION_BIT = "x"

type ErrMsg string

const (
	ERROR_READ_STDIN               = "Error reading stdin %s\n"
	INVALID_EXIT_CODE              = "Coud not parse exit code %s\n"
	COMMAND_NOT_FOUND              = "%s command not found\n"
	ERROR_READING_EXTERNAL_PROGRAM = "Could not read external program %s\n"
	PERMISSION_DENIED              = "%s: You shall not pass!\n"
	ERROR_FROM_EXT_PROGRAM         = "Error from %s: %s\n"
	INVALID_EXIT_STATUS_CODE       = "Coud not parse exit status code %s\n"
)
