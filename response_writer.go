package khronos

import (
	"io"
)

// ResponseWriter write RESP (REdis Serialization Protocol) is the protocol used in Redis.
type ResponseWriter interface {
	WriteError(err error) error
	WriteStatus(status Status) error
	WriteInt64(i int64) error
	WriteArray(a []string) error
	WriteString(s string) error
	Write(b []byte) (int, error)
}

type responseWriter struct {
	io.Writer
}

func (w *responseWriter) WriteFrom(reader io.Reader) error {
	_, err := io.Copy(w.Writer, reader)
	return err
}

func (w *responseWriter) WriteError(err error) error {
	builder := getprotocolBuilder()
	defer putProtocolBuilder(builder)
	builder.WriteError(err)
	return w.WriteFrom(builder)
}

func (w *responseWriter) WriteStatus(status Status) error {
	builder := getprotocolBuilder()
	defer putProtocolBuilder(builder)
	builder.WriteStatus(status.String())
	return w.WriteFrom(builder)
}

func (w *responseWriter) WriteInt64(i int64) error {
	builder := getprotocolBuilder()
	defer putProtocolBuilder(builder)
	builder.WriteInt64(i)
	return w.WriteFrom(builder)
}

func (w *responseWriter) WriteString(s string) error {
	builder := getprotocolBuilder()
	defer putProtocolBuilder(builder)
	builder.WriteString(s)
	return w.WriteFrom(builder)
}

func (w *responseWriter) WriteArray(a []string) error {
	builder := getprotocolBuilder()
	defer putProtocolBuilder(builder)
	builder.WriteArray(a)
	return w.WriteFrom(builder)
}
