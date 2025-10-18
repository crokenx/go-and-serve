package request

import (
	"errors"
	"io"
	"strconv"
	"strings"
	"unicode"

	"boot.httpserver/internal/headers"
)

const bufferSize = 8

type ParserState int

const (
	Initialized ParserState = iota
	ParsingHeaders
	ParsingBody
	Done
)

type Request struct {
	RequestLine RequestLine
	State       ParserState
	Headers     headers.Headers
	Body        []byte
}

func (r *Request) parse(data []byte) (int, error) {
	totalBytesParsed := 0

	for r.State != Done {
		n, err := r.parseSingle(data[totalBytesParsed:])

		if err != nil {
			return 0, err
		}

		if n == 0 {
			return totalBytesParsed, nil
		}

		totalBytesParsed += n

	}

	return totalBytesParsed, nil
}

func (r *Request) parseSingle(data []byte) (int, error) {
	switch r.State {
	case Initialized:
		requestLine, n, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		}
		if n == 0 {
			return 0, nil
		}
		r.RequestLine = *requestLine
		r.State = ParsingHeaders

		return n, nil
	case ParsingHeaders:
		n, done, err := r.Headers.Parse(data)

		if err != nil {
			return 0, err
		}

		if done {
			r.State = ParsingBody
		}

		return n, nil
	case ParsingBody:
		value, exists := r.Headers.Get("Content-Length")

		if !exists {
			r.State = Done
			return 0, nil
		}

		num, err := strconv.Atoi(value)

		if err != nil {
			return 0, err
		}

		if len(data) > num {
			return 0, errors.New("invalid length")
		}

		if len(data) == num {
			r.State = Done
			r.Body = append([]byte(nil), data...)

			return len(data), nil
		}

		return 0, nil
	case Done:
		return 0, errors.New("request already done")
	default:
		return 0, errors.New("unknown parser state")
	}
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	buf := make([]byte, bufferSize)
	readToIndex := 0

	request := &Request{
		State:   Initialized,
		Headers: make(headers.Headers),
	}

	for request.State != Done {
		// if the buffer is full double it
		if readToIndex >= len(buf) {
			aux := make([]byte, cap(buf)*2)
			copy(aux, buf)
			buf = aux
		}

		n, err := reader.Read(buf[readToIndex:])

		if err != nil {
			if err == io.EOF {
				request.State = Done
				break
			}
			return nil, err
		}

		readToIndex += n

		n, err = request.parse(buf[:readToIndex])

		if err != nil {
			return nil, err
		}

		copy(buf, buf[n:])

		readToIndex -= n
	}

	if value, exists := request.Headers.Get("Content-Length"); exists {
		num, err := strconv.Atoi(value)

		if err != nil {
			return nil, err
		}

		if len(request.Body) < num {
			return nil, errors.New("invalid length")
		}
	}

	if !isAllUpperCase(request.RequestLine.Method) {
		return nil, errors.New("method not allowed")
	}

	if request.RequestLine.HttpVersion != "1.1" {
		return nil, errors.New("http version not supported")
	}

	return request, nil
}

func parseRequestLine(line []byte) (*RequestLine, int, error) {
	str := string(line)
	idx := strings.Index(str, "\r\n")

	if idx == -1 {
		return nil, 0, nil
	}

	str = str[:idx]

	parts := strings.Split(str, " ")

	if len(parts) < 3 {
		return nil, 0, errors.New("line is invalid")
	}

	version := strings.Split(parts[2], "/")[1]

	requestLine := RequestLine{
		HttpVersion:   version,
		RequestTarget: parts[1],
		Method:        parts[0],
	}

	consumed := idx + 2

	return &requestLine, consumed, nil
}

func isAllUpperCase(s string) bool {
	for _, r := range s {
		if !unicode.IsLetter(r) || !unicode.IsUpper(r) {
			return false
		}
	}
	return true
}
