package server

import (
	"bytes"
	"io"
	"log"
	"net"
	"strconv"
	"sync/atomic"

	"boot.httpserver/internal/request"
	"boot.httpserver/internal/response"
)

type HandlerError struct {
	StatusCode int
	Message    string
}

func (e *HandlerError) writeError(w io.Writer) {
	response.WriteStatus(w, response.StatusCode(e.StatusCode))
	messageBytes := []byte(e.Message)
	headers := response.GetDefaultHeaders(len(messageBytes))
	response.WriteHeaders(w, headers)
	w.Write(messageBytes)
}

type Handler func(w *response.Writer, req *request.Request)

type Server struct {
	handler   Handler
	listening atomic.Bool
	listener  net.Listener
}

func Serve(port int, handler Handler) (*Server, error) {
	server := &Server{}
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		log.Fatal(err)
	}
	server.listener = listener
	server.handler = handler
	go server.listen()
	server.listening.Store(true)
	return server, nil
}

func (s *Server) listen() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if !s.listening.Load() {
				return
			}
			log.Printf("Error accepting connection: %v\n", err)
			continue
		}
		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()
	log.Printf("Got connection from %s\n", conn.RemoteAddr())

	rq, err := request.RequestFromReader(conn)

	if err != nil {
		errorHandler := &HandlerError{
			StatusCode: 500,
			Message:    err.Error(),
		}
		errorHandler.writeError(conn)
		log.Printf("Error creating error: %v\n", err)
		return
	}

	buf := bytes.NewBuffer([]byte{})
	writer := &response.Writer{}
	writer.Wrt = buf

	s.handler(writer, rq)

	b := buf.Bytes()
	conn.Write(b)
}

func (s *Server) Close() error {
	s.listening.Store(false)
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}
