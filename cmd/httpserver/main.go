package main

import (
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

func handler(w *response.Writer, req *request.Request) {
	statusCode := response.OK
	body := "<html>\n  <head>\n    <title>200 OK</title>\n  </head>\n  <body>\n    <h1>Success!</h1>\n    <p>Your request was an absolute banger.</p>\n  </body>\n</html>"
	if req.RequestLine.RequestTarget == "/yourproblem" {
		statusCode = response.BADREQUEST
		body = "<html>\n  <head>\n    <title>400 Bad Request</title>\n  </head>\n  <body>\n    <h1>Bad Request</h1>\n    <p>Your request honestly kinda sucked.</p>\n  </body>\n</html>"
	}
	if req.RequestLine.RequestTarget == "/myproblem" {
		statusCode = response.INTERNALERROR
		body = "<html>\n  <head>\n    <title>500 Internal Server Error</title>\n  </head>\n  <body>\n    <h1>Internal Server Error</h1>\n    <p>Okay, you know what? This one is on me.</p>\n  </body>\n</html>"
	}
	err := w.WriteStatusLine(statusCode)
	if err != nil {
		log.Printf("error writing status line: %v", err)
	}

	headers := response.GetDefaultHeaders(len(body))
	headers.Set("Content-Type", "text/html")
	err = w.WriteHeaders(headers)

	w.WriteBody([]byte(body))
}
