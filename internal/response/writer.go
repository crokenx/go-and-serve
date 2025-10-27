package response

import (
	"fmt"
	"io"

	"boot.httpserver/internal/headers"
)

type WriterStatus int

const (
	WRITINGSTATUSLINE WriterStatus = iota
	WRITINGHEADERS
	WRITINGBODY
)

type Writer struct {
	Wrt         io.Writer
	WriterState WriterStatus
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.WriterState != WRITINGSTATUSLINE {
		return fmt.Errorf("cannot write status line in state %d", w.WriterState)
	}
	defer func() { w.WriterState = WRITINGHEADERS }()
	response, ok := responses[statusCode]
	if !ok {
		response = ""
	}
	_, err := w.Wrt.Write([]byte(response))
	if err != nil {
		return err
	}
	return nil
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	if w.WriterState != WRITINGHEADERS {
		return fmt.Errorf("cannot write headers in state %d", w.WriterState)
	}
	defer func() { w.WriterState = WRITINGBODY }()
	for k, v := range headers {
		_, err := w.Wrt.Write([]byte(k + ": " + v + "\r\n"))
		if err != nil {
			return err
		}
	}
	_, err := w.Wrt.Write([]byte("\r\n"))
	if err != nil {
		return err
	}
	return nil
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	if w.WriterState != WRITINGBODY {
		return 0, fmt.Errorf("cannot write body in state %d", w.WriterState)
	}
	n, err := w.Wrt.Write(p)
	if err != nil {
		return n, err
	}
	return n, nil
}

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	if w.WriterState != WRITINGBODY {
		return 0, fmt.Errorf("cannot write body in state %d", w.WriterState)
	}

	chunkHeader := fmt.Sprintf("%x\r\n", len(p))
	if _, err := io.WriteString(w.Wrt, chunkHeader); err != nil {
		return 0, err
	}

	if _, err := w.Wrt.Write(p); err != nil {
		return 0, err
	}

	if _, err := io.WriteString(w.Wrt, "\r\n"); err != nil {
		return 0, err
	}

	return len(p), nil
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	return w.Wrt.Write([]byte("0\r\n"))
}

func (w *Writer) WriteTrailers(h headers.Headers) error {
	for key, value := range h {
		_, err := io.WriteString(w.Wrt, key+": "+value+"\r\n")
		if err != nil {
			return err
		}
	}
	_, err := io.WriteString(w.Wrt, "\r\n")
	return err
}
