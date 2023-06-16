package khronos

import (
	"context"
	"strconv"
)

// Command is an interface that represents a command that can be executed.
type Command interface {
	// Name returns the name of the command.
	// For example, the command "ping" has the name "ping".
	// The name is used to look up the command in the command library, and it is the identifier used in the protocol.
	Name() string

	// Args returns the arguments of the command.
	Args() []string

	// Execute executes the command and writes the response to the writer.
	Execute(ctx context.Context, write ResponseWriter) error
}

// CommandConstructor is a function that constructs a command.
type CommandConstructor func(args []string) (Command, error)

// RegisterCommand registers a command constructor.
var commandLibraries = make(map[string]CommandConstructor)

// ArgsCommand is a command that has arguments.
type ArgsCommand struct {
	args []string
}

// Args returns the arguments of the command.
func (a ArgsCommand) Args() []string {
	return a.args
}

type CommandCommand struct {
	ArgsCommand
}

func (c *CommandCommand) Name() string {
	return "command"
}

func (c *CommandCommand) Execute(_ context.Context, writer ResponseWriter) error {
	args := c.Args()
	if len(args) != 0 {
		return writer.WriteError(&wrongNumberOfArgsError{c.Name()})
	}
	commands := make([]string, 0, len(commandLibraries))
	for name := range commandLibraries {
		commands = append(commands, name)
	}
	return writer.WriteArray(commands)
}

// PingCommand is the command "ping".
// It is used to test if the server is alive.
// This command has one or zero arguments.
// If the command has no arguments, the server replies with "PONG".
// Otherwise, the server replies with the argument with string type.
type PingCommand struct {
	ArgsCommand
}

func (c *PingCommand) Name() string {
	return "ping"
}

func (c *PingCommand) Execute(_ context.Context, writer ResponseWriter) error {
	args := c.Args()
	if len(args) == 0 {
		return writer.WriteStatus(Pong)
	}
	if len(args) == 1 {
		return writer.WriteString(args[0])
	}
	return writer.WriteError(&wrongNumberOfArgsError{c.Name()})
}

func NewPingCommand(args []string) (Command, error) {
	if len(args) > 1 {
		return nil, &wrongNumberOfArgsError{"ping"}
	}
	cmd := &PingCommand{}
	cmd.args = args
	return cmd, nil
}

// EchoCommand is the command "echo".
// It is used to test if the server is alive.
// This command has one argument.
// The server replies with the argument with string type.
type EchoCommand struct {
	ArgsCommand
}

// Name returns the name of the command.
func (c *EchoCommand) Name() string {
	return "echo"
}

// Execute executes the command and writes the response to the writer.
func (c *EchoCommand) Execute(_ context.Context, writer ResponseWriter) error {
	args := c.Args()
	if len(args) != 1 {
		return writer.WriteError(&wrongNumberOfArgsError{c.Name()})
	}
	return writer.WriteString(args[0])
}

func NewEchoCommand(args []string) (Command, error) {
	if len(args) != 1 {
		return nil, &wrongNumberOfArgsError{"echo"}
	}
	cmd := &EchoCommand{}
	cmd.args = args
	return cmd, nil
}

type PushCommand struct {
	ArgsCommand
}

func (c *PushCommand) Name() string {
	return "push"
}

func (c *PushCommand) Execute(ctx context.Context, writer ResponseWriter) error {
	args := c.Args()
	if len(args) != 3 {
		return &wrongNumberOfArgsError{"push"}
	}
	key, value, score := args[0], args[1], args[2]
	priority, err := strconv.ParseInt(score, 10, 64)
	if err != nil {
		return err
	}
	pq := PqFromContext(ctx)
	item := &Item{value: value, priority: priority}
	pq.Enqueue(key, item)
	return writer.WriteStatus(OK)
}

func NewPushCommand(args []string) (Command, error) {
	if len(args) != 3 {
		return nil, &wrongNumberOfArgsError{"push"}
	}
	cmd := &PushCommand{}
	cmd.args = args
	return cmd, nil
}

type PopCommand struct {
	ArgsCommand
}

func (c *PopCommand) Name() string {
	return "pop"
}

func (c *PopCommand) Execute(ctx context.Context, writer ResponseWriter) error {
	args := c.Args()
	if len(args) != 1 {
		return &wrongNumberOfArgsError{"pop"}
	}
	key := args[0]
	pq := PqFromContext(ctx)
	item := pq.Dequeue(key)
	return writer.WriteString(item.value)
}

func NewPopCommand(args []string) (Command, error) {
	if len(args) != 1 {
		return nil, &wrongNumberOfArgsError{"pop"}
	}
	cmd := &PopCommand{}
	cmd.args = args
	return cmd, nil
}

type LengthCommand struct {
	ArgsCommand
}

func (c *LengthCommand) Name() string {
	return "length"
}

func (c *LengthCommand) Execute(ctx context.Context, writer ResponseWriter) error {
	args := c.Args()
	if len(args) != 1 {
		return &wrongNumberOfArgsError{"length"}
	}
	key := args[0]
	pq := PqFromContext(ctx)
	length := pq.Length(key)
	return writer.WriteInt64(int64(length))
}

func NewLengthCommand(args []string) (Command, error) {
	if len(args) != 1 {
		return nil, &wrongNumberOfArgsError{"length"}
	}
	cmd := &LengthCommand{}
	cmd.args = args
	return cmd, nil
}

type QuitCommand struct{}

func (c *QuitCommand) Args() []string {
	return nil
}

func (c *QuitCommand) Name() string {
	return "quit"
}

func (c *QuitCommand) Execute(_ context.Context, writer ResponseWriter) error {
	args := c.Args()
	if len(args) != 0 {
		return &wrongNumberOfArgsError{"quit"}
	}
	_, _ = writer.Write([]byte("\r\n"))
	return ErrQuit
}

func NewQuitCommand(args []string) (Command, error) {
	if len(args) != 0 {
		return nil, &wrongNumberOfArgsError{"quit"}
	}
	cmd := &QuitCommand{}
	return cmd, nil
}

func init() {
	commandLibraries["ping"] = NewPingCommand
	commandLibraries["echo"] = NewEchoCommand
	commandLibraries["push"] = NewPushCommand
	commandLibraries["pop"] = NewPopCommand
	commandLibraries["length"] = NewLengthCommand
	commandLibraries["quit"] = NewQuitCommand
}
