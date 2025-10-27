package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"boot.httpserver/internal/headers"
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
	body := "<html>\n  <head>\n    <title>200 OK</title>\n  </head>\n  <body>\n    <h1>Success!</h1>\n    <p>Your request was an absolute banger.</p>\n  </body>\n</html>\r\n"
	if req.RequestLine.RequestTarget == "/yourproblem" {
		statusCode = response.BADREQUEST
		body = "<html>\n  <head>\n    <title>400 Bad Request</title>\n  </head>\n  <body>\n    <h1>Bad Request</h1>\n    <p>Your request honestly kinda sucked.</p>\n  </body>\n</html>\r\n"
	}
	if req.RequestLine.RequestTarget == "/myproblem" {
		statusCode = response.INTERNALERROR
		body = "<html>\n  <head>\n    <title>500 Internal Server Error</title>\n  </head>\n  <body>\n    <h1>Internal Server Error</h1>\n    <p>Okay, you know what? This one is on me.</p>\n  </body>\n</html>\r\n"
	}

	err := w.WriteStatusLine(statusCode)
	if err != nil {
		log.Printf("error writing status line: %v", err)
	}

	h := response.GetDefaultHeaders(len(body))

	if req.RequestLine.RequestTarget == "/video" {
		videoHandler(w, h, req.RequestLine.RequestTarget)
		return
	}

	if proxying, target := shouldProxy(req); proxying {
		proxyRequest(w, h, target)
		return
	}

	h.Set("Content-Type", "text/html")
	err = w.WriteHeaders(h)

	w.WriteBody([]byte(body))
}

func shouldProxy(req *request.Request) (bool, string) {
	hasPrefix := strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin")
	if !hasPrefix {
		return false, ""
	}
	target := req.RequestLine.RequestTarget
	target = strings.TrimPrefix(target, "/httpbin")
	return true, target
}

func proxyRequest(w *response.Writer, headers headers.Headers, target string) {
	delete(headers, "Content-Length")
	headers.Set("Transfer-Encoding", "chunked")
	headers.Set("Trailer", "X-Content-Sha256, X-Content-Length")

	w.WriteHeaders(headers)
	res, err := http.Get("https://httpbin.org" + target)
	log.Printf("proxying %s to https://httpbin.org%s\n", target, target)
	if err != nil {
		log.Printf("error proxying request: %v", err)
	}
	buff := make([]byte, 1024)
	totalBuff := make([]byte, 0, 1024)

	for {
		n, e := res.Body.Read(buff)
		totalBuff = append(totalBuff, buff[:n]...)

		log.Printf("chunking %d bytes\n", n)
		if e == io.EOF {
			log.Printf("EOF | ending proxying: %v", e)
			break
		}
		if e != nil {
			log.Printf("error proxying response: %v", e)
			break
		}
		_, e = w.WriteChunkedBody(buff[:n])
		if e != nil {
			log.Printf("error proxying response: %v", e)
		}
	}
	hash := sha256.Sum256(totalBuff)
	trailers := map[string]string{
		"X-Content-Sha256": fmt.Sprintf("%x", hash),
		"X-Content-Length": fmt.Sprintf("%d", len(totalBuff)),
	}
	w.WriteChunkedBodyDone()
	w.WriteTrailers(trailers)
}

func videoHandler(w *response.Writer, headers headers.Headers, target string) {
	delete(headers, "Content-Length")
	headers.Set("Content-Type", "video/mp4")
	w.WriteHeaders(headers)
	data, err := os.ReadFile("assets/vim.mp4")
	if err != nil {
		log.Printf("error reading video")
	}
	w.WriteBody(data)
}
