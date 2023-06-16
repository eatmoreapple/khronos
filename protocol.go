package khronos

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"strconv"
	"strings"
	"sync"
)

var ErrInvalidSyntax = errors.New("invalid syntax")

const (
	ErrorReply  = '-'
	StatusReply = '+'
	IntReply    = ':'
	StringReply = '$'
	ArrayReply  = '*'
)

func parseInt(b []byte) (int, error) {
	if len(b) == 0 {
		return 0, ErrInvalidSyntax
	}
	return strconv.Atoi(string(b))
}

type RespProtocolParser struct{ *bufio.Reader }

func (p *RespProtocolParser) readLine() ([]byte, error) {
	line, _, err := p.Reader.ReadLine()
	if err != nil {
		return nil, err
	}
	return line, nil
}

// readArrayLength reads the array length from the reader.
// each protocol message starts with a byte indicating the type of the message.
// for array, it is '*'.
func (p *RespProtocolParser) readArrayLength() (int, error) {
	line, err := p.readLine()
	if err != nil {
		return 0, err
	}
	if len(line) == 0 {
		return 0, ErrInvalidSyntax
	}
	if line[0] != ArrayReply {
		return 0, ErrInvalidSyntax
	}
	return parseInt(line[1:])
}

// readString reads an argument from the reader.
func (p *RespProtocolParser) readString() (string, error) {
	line, err := p.readLine()
	if err != nil {
		return "", err
	}
	if len(line) == 0 {
		return "", ErrInvalidSyntax
	}
	if line[0] != StringReply {
		return "", ErrInvalidSyntax
	}
	length, err := parseInt(line[1:])
	if err != nil {
		return "", err
	}
	var buf = make([]byte, length)
	if _, err = io.ReadFull(p, buf); err != nil {
		return "", err
	}
	// discard the trailing crlf
	if _, err = p.Discard(2); err != nil {
		return "", err
	}
	return string(buf), nil
}

// readCommandName reads the command name from the reader.
func (p *RespProtocolParser) readCommandName() (string, error) {
	return p.readString()
}

// readCommandArgs reads the command arguments from the reader.
func (p *RespProtocolParser) readCommandArgs(length int) ([]string, error) {
	var args = make([]string, length)
	for i := 0; i < length; i++ {
		arg, err := p.readString()
		if err != nil {
			return nil, err
		}
		args[i] = arg
	}
	return args, nil
}

// Parse reads a command from the reader.
func (p *RespProtocolParser) Parse() (string, []string, error) {
	length, err := p.readArrayLength()
	if err != nil {
		return "", nil, err
	}
	if length == 0 {
		return "", nil, ErrInvalidSyntax
	}
	name, err := p.readCommandName()
	if err != nil {
		return "", nil, err
	}
	args, err := p.readCommandArgs(length - 1)
	if err != nil {
		return "", nil, err
	}
	return name, args, nil
}

func NewRespProtocolParser(r io.Reader) *RespProtocolParser {
	return &RespProtocolParser{bufio.NewReader(r)}
}

type CommandParser struct {
	command Command
}

// Write do nothing just to implement io.Writer.
// this is used to hook io.Copy.
func (p *CommandParser) Write(b []byte) (int, error) {
	return len(b), nil
}

// ReadFrom hooks io.Copy
func (p *CommandParser) ReadFrom(r io.Reader) (int64, error) {
	parser := NewRespProtocolParser(r)
	cmd, args, err := parser.Parse()
	if err != nil {
		return 0, err
	}
	cmd = strings.ToLower(cmd)
	constructor, ok := commandLibraries[cmd]
	if !ok {
		return 0, &wrongCommandError{command: cmd, args: args}
	}
	command, err := constructor(args)
	if err != nil {
		return 0, err
	}
	p.command = command
	return 0, err
}

type protocolBuilder struct {
	*bytes.Buffer
}

func (w *protocolBuilder) WriteString(s string) {
	w.Write([]byte("$" + strconv.Itoa(len(s)) + "\r\n" + s + "\r\n"))
}

func (w *protocolBuilder) WriteInt64(i int64) {
	w.Write([]byte(":" + strconv.FormatInt(i, 10) + "\r\n"))
}

func (w *protocolBuilder) WriteError(err error) {
	w.Write([]byte("-" + err.Error() + "\r\n"))
}

func (w *protocolBuilder) WriteStatus(s string) {
	w.Write([]byte("+" + s + "\r\n"))
}

func (w *protocolBuilder) WriteArray(a []string) {
	w.WriteInt64(int64(len(a)))
	for _, s := range a {
		w.WriteString(s)
	}
}

var protocolWriterPool = sync.Pool{
	New: func() interface{} {
		return &protocolBuilder{bytes.NewBuffer(nil)}
	},
}

func getprotocolBuilder() *protocolBuilder {
	return protocolWriterPool.Get().(*protocolBuilder)
}

func putProtocolBuilder(w *protocolBuilder) {
	w.Reset()
	protocolWriterPool.Put(w)
}
