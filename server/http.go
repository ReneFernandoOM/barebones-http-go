package server

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
)

type Request struct {
	Method  string
	URI     string
	Version string
	Headers map[string]string
	Params  map[string]string
	Body    []byte
}

type Response struct {
	StatusCode    int
	StatusMsg     string
	ContentType   string
	ContentLength int
	Headers       map[string]string
	Body          []byte
}

func NewResponse(statusCode int, statusMsg string, contentType string, body []byte) Response {
	return Response{
		StatusCode:    statusCode,
		StatusMsg:     statusMsg,
		ContentType:   contentType,
		ContentLength: len(body),
		Headers:       make(map[string]string),
		Body:          body,
	}
}

func TextResponse(status int, body string) Response {
	return NewResponse(status, statusText(status), "text/plain", []byte(body))
}

func FileResponse(status int, data []byte) Response {
	return NewResponse(status, statusText(status), "application/octet-stream", data)
}

func (r Response) String() string {
	// implement this using strings.builder (just because)
	var buf bytes.Buffer

	// status line
	fmt.Fprintf(&buf, "HTTP/1.1 %d %s\r\n", r.StatusCode, r.StatusMsg)

	if r.ContentType != "" {
		fmt.Fprintf(&buf, "Content-Type: %s\r\n", r.ContentType)
	}

	fmt.Fprintf(&buf, "Content-Length: %d\r\n", r.ContentLength)

	// custom headers
	for key, value := range r.Headers {
		fmt.Fprintf(&buf, "%s: %s\r\n", key, value)
	}

	buf.WriteString("\r\n")

	if r.Body != nil {
		buf.Write(r.Body)
	}
	return buf.String()
}

func parseRequest(conn net.Conn) (*Request, error) {
	reader := bufio.NewReader(conn)

	// parse request line
	requestLine, err := reader.ReadString('\n')

	// client closed connection suddenly
	if err == io.EOF {
		return nil, err
	}

	if err != nil {
		return nil, fmt.Errorf("reading request line: %v", err)
	}

	parts := strings.Split(strings.TrimSpace(requestLine), " ")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid request line: %s", requestLine)
	}

	method := parts[0]
	uri := parts[1]
	version := parts[2]

	if !strings.HasSuffix(uri, "/") && !strings.Contains(uri, "?") {
		uri = uri + "/"
	}

	headers := make(map[string]string)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("reading header: %v", err)
		}

		line = strings.TrimSpace(line)
		// end of headers
		if line == "" {
			break
		}

		headerAndValue := strings.Split(line, ":")

		key := strings.ToLower(strings.TrimSpace(headerAndValue[0]))
		value := strings.TrimSpace(headerAndValue[1])

		headers[key] = value
	}

	// parse body if `content-length` is present
	var body []byte
	if contentLenStr, ok := headers["content-length"]; ok {
		contentLen, err := strconv.Atoi(contentLenStr)
		if err != nil {
			return nil, fmt.Errorf("invalid content length: %v", err)
		}

		if contentLen > 0 {
			body = make([]byte, contentLen)
			if _, err := io.ReadFull(reader, body); err != nil {
				return nil, fmt.Errorf("reading body: %v", err)
			}
		}
	}

	return &Request{
		Method:  method,
		URI:     uri,
		Version: version,
		Headers: headers,
		Params:  make(map[string]string),
		Body:    body,
	}, nil
}

func statusText(code int) string {
	switch code {
	case 200:
		return "OK"
	case 201:
		return "Created"
	case 204:
		return "No Content"
	case 400:
		return "Bad Request"
	case 401:
		return "Unauthorized"
	case 403:
		return "Forbidden"
	case 404:
		return "Not Found"
	case 405:
		return "Method Not Allowed"
	case 500:
		return "Internal Server Error"
	default:
		return "Unknown"
	}
}
