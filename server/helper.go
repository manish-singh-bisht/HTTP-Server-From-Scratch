package server

import (
	"fmt"
	"net"
	"strings"
)

/**
 * parseRequest is used to parse the request from the client
 * @param data string - the data to parse
 * @return *HTTPRequest - the parsed request
 */
 func (s *HTTPServer) parseRequest(data string) *HTTPRequest {
	parts := strings.Split(data, "\r\n\r\n")
	// The firstSection consist of requestLine+headers and secondSection will be "body".
	firstSection := parts[0]

	lines := strings.Split(firstSection, "\r\n")
	if len(lines) == 0 {
		return nil
	}

	requestLine := strings.Fields(lines[0])
	if len(requestLine) != 3 {
		return nil
	}

	headers := make(map[string]string)
	for _, line := range lines[1:] {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, ": ", 2)
		if len(parts) == 2 {
			headers[parts[0]] = parts[1]
		}
	}

	var body string
	if len(parts) > 1 {
		body = parts[1]
	}

	return &HTTPRequest{
		Method:  requestLine[0],
		Path:    requestLine[1],
		Headers: headers,
		Body:    body,
	}
}

/**
 * writeResponse is used to write the response to the client
 * @param conn net.Conn - the connection to write the response to
 * @param resp *HTTPResponse - the response to write
 */
func (s *HTTPServer) writeResponse(conn net.Conn, resp *HTTPResponse) {
	// Join the headers
	headers := ""
	for k, v := range resp.Headers {
		headers += fmt.Sprintf("%s: %s\r\n", k, v)
	}

	response := fmt.Sprintf("HTTP/1.1 %d\r\n%s\r\n%s", resp.StatusCode, headers, resp.Body)
	conn.Write([]byte(response))
}