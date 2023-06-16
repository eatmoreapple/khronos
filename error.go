package khronos

import "strings"

type wrongNumberOfArgsError struct {
	command string
}

func (e *wrongNumberOfArgsError) Error() string {
	return "ERR wrong number of arguments for '" + e.command + "' command"
}

type wrongCommandError struct {
	command string
	args    []string
}

func (e *wrongCommandError) Error() string {
	if len(e.args) > 0 {
		return "ERR wrong command '" + e.command + "' args = " + strings.Join(e.args, " ")
	}
	return "ERR unknown command '" + e.command + "'"
}
