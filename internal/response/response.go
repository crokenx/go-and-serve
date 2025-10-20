package response

import (
	"io"
	"strconv"

	"boot.httpserver/internal/headers"
)

type StatusCode int

const (
	OK            StatusCode = 200
	BADREQUEST               = 400
	INTERNALERROR            = 500
)

var responses = map[StatusCode]string{
	OK:            "HTTP/1.1 200 OK\r\n",
	BADREQUEST:    "HTTP/1.1 400 Bad Request\r\n",
	INTERNALERROR: "HTTP/1.1 500 Internal Server Error\r\n",
}

func WriteStatus(w io.Writer, status StatusCode) error {
	response, ok := responses[status]
	if !ok {
		response = ""
	}

	_, err := w.Write([]byte(response))

	if err != nil {
		return err
	}

	return nil
}

func GetDefaultHeaders(contentLength int) headers.Headers {
	h := make(map[string]string)

	h["Content-Length"] = strconv.Itoa(contentLength)
	h["Connection"] = "close"
	h["Content-Type"] = "text/plain"

	return h
}

func WriteHeaders(w io.Writer, headers headers.Headers) error {
	for key, value := range headers {
		_, err := w.Write([]byte(key + ": " + value + "\r\n"))

		if err != nil {
			return err
		}
	}

	_, err := w.Write([]byte("\r\n"))

	if err != nil {
		return err
	}

	return nil
}
