package cursor

import (
	"bufio"
	"io"
	"strings"
)

type sseStreamReader struct {
	reader  *bufio.Reader
	body    io.ReadCloser
	closed  bool
}

func newSSEStreamReader(body io.ReadCloser) *sseStreamReader {
	return &sseStreamReader{
		reader: bufio.NewReaderSize(body, 4096),
		body:   body,
	}
}

func (s *sseStreamReader) Next() (SSEEvent, error) {
	if s.closed {
		return SSEEvent{}, io.EOF
	}

	var event SSEEvent
	hasData := false

	for {
		line, err := s.reader.ReadString('\n')
		line = strings.TrimRight(line, "\r\n")

		if err != nil {
			if err == io.EOF {
				if hasData {
					return event, nil
				}
				s.closed = true
				return SSEEvent{}, io.EOF
			}
			return SSEEvent{}, err
		}

		if line == "" {
			if hasData {
				return event, nil
			}
			continue
		}

		if strings.HasPrefix(line, ":") {
			continue
		}

		field, value, _ := strings.Cut(line, ":")
		value = strings.TrimPrefix(value, " ")

		switch field {
		case "event":
			event.Event = value
		case "data":
			if hasData {
				event.Data += "\n" + value
			} else {
				event.Data = value
				hasData = true
			}
		case "id":
			event.ID = value
		}
	}
}

func (s *sseStreamReader) Close() error {
	s.closed = true
	return s.body.Close()
}
