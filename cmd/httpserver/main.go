package main

import (
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	"boot.httpserver/internal/request"
	"boot.httpserver/internal/response"
	"boot.httpserver/internal/server"
)

const port = 42069

func main() {
	srv, err := server.Serve(port, handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer srv.Close()
	log.Printf("Server started on port: %d\n", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Printf("Server gracefully stopped\n")
}

func handler(w io.Writer, req *request.Request) *server.HandlerError {
	if req.RequestLine.RequestTarget == "/yourproblem" {
		return &server.HandlerError{
			StatusCode: response.BADREQUEST,
			Message:    "Your problem is not my problem\n",
		}
	}
	if req.RequestLine.RequestTarget == "/myproblem" {
		return &server.HandlerError{
			StatusCode: response.INTERNALERROR,
			Message:    "Woopsie, my bad\n",
		}
	}
	w.Write([]byte("All good, frfr\n"))
	return nil
}
