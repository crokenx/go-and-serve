package main

import (
	"fmt"
	"log"
	"net"

	"boot.httpserver/internal/request"
)

func main() {

	conn, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	for {
		conn, err := conn.Accept()
		if err != nil {
			break
		}
		fmt.Println("A connection has been accepted")

		rq, err := request.RequestFromReader(conn)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Request line: \n - Method: %s\n - Target: %s\n - Version: %s\n", rq.RequestLine.Method, rq.RequestLine.RequestTarget, rq.RequestLine.HttpVersion)
		fmt.Println("Headers:")

		for key, value := range rq.Headers {
			fmt.Printf("  - %s: %s\n", key, value)
		}

		fmt.Printf("Body:\n%s\n", string(rq.Body))

		fmt.Printf("The connection has been closed\n")
	}
}
