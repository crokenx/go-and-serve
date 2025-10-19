package server

import (
	"log"
	"net"
	"strconv"
	"sync/atomic"
)

type Server struct {
	listening atomic.Bool
	listener  net.Listener
}

func Serve(port int) (*Server, error) {
	server := &Server{}
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		log.Fatal(err)
	}
	server.listener = listener
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
	response := []byte(
		"HTTP/1.1 200 OK\r\n" +
			"Content-Type: text/plain\r\n" +
			"\r\n" +
			"Hello World!\n")
	conn.Write(response)
}

func (s *Server) Close() error {
	s.listening.Store(false)
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}
