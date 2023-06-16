package khronos

import (
	"context"
	"errors"
	"io"
	"log"
	"net"
)

var ErrQuit = errors.New("quit")

type Server struct {
	Addr string

	BaseContext func(net.Listener) context.Context

	ConnContext func(context.Context, net.Conn) context.Context

	Logger *log.Logger

	Queue *PriorityQueueWithRouting
}

func (srv *Server) ListenAndServe() error {
	addr := srv.Addr
	if addr == "" {
		addr = ":7464"
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	return srv.Serve(ln)
}

func (srv *Server) Serve(listener net.Listener) error {
	ctx := context.Background()
	if srv.BaseContext != nil {
		ctx = srv.BaseContext(listener)
		if ctx == nil {
			panic("BaseContext returned a nil context")
		}
	}
	// set server to context
	ctx = context.WithValue(ctx, ServerContextKey, srv)

	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}
		// copy listener's context and add conn
		connCtx := ctx
		if srv.ConnContext != nil {
			connCtx = srv.ConnContext(connCtx, conn)
			if connCtx == nil {
				panic("ConnContext returned a nil context")
			}
		}

		connCtx = PqWithContext(connCtx, srv.Queue)

		go srv.serveConn(connCtx, conn)
	}
}

func (srv *Server) serveConn(ctx context.Context, conn net.Conn) {
	c := &connContext{conn: conn, ctx: ctx}
	writer := &responseWriter{conn}
	defer func() { _ = conn.Close() }()
	for {
		// FIXME
		if err := c.serve(writer); err != nil {
			if errors.Is(err, ErrQuit) {
				srv.logf("khronos: conn closed: %v", err)
				return
			}
			if err = writer.WriteError(err); err != nil {
				srv.logf("khronos: conn error: %v", err)
			}
		}
	}
}

func (srv *Server) logf(format string, args ...interface{}) {
	if srv.Logger != nil {
		srv.Logger.Printf(format, args...)
	}
}

type connContext struct {
	conn net.Conn
	ctx  context.Context
}

func (c *connContext) serve(writer ResponseWriter) error {
	var parser CommandParser
	for {
		select {
		case <-c.ctx.Done():
			return c.ctx.Err()
		default:
		}
		// read command from connection
		// it will block until read a complete command
		if _, err := io.Copy(&parser, c.conn); err != nil {
			return err
		}
		if err := parser.command.Execute(c.ctx, writer); err != nil {
			return err
		}

		// TODO reset deadline line here
		//if err := c.conn.SetDeadline(time.Now().Add(time.Second * 5)); err != nil {
		//	return err
		//}
	}
}

func ListenAndServe(addr string) error {
	server := &Server{
		Addr:   addr,
		Queue:  NewPriorityQueueWithRouting(),
		Logger: log.Default(),
	}
	return server.ListenAndServe()
}
