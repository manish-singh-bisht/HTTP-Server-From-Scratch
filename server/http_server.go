package server

import (
	"fmt"
	"log"
	"net"
	"os"
)

const (
	MAX_WORKERS = 100
)

type HTTPServer struct {
	address       string
	port          int
	routes        map[string][]func(*HTTPRequest, *HTTPResponse, *bool)
	appMiddleware []func(*HTTPRequest, *HTTPResponse, *bool)
	workerPool    *Pool
}

type HTTPRequest struct {
	Method  string
	Path    string
	Headers map[string]string
	Body    string
}

type HTTPResponse struct {
	StatusCode int
	Headers    map[string]string
	Body       string
}

func NewHTTPServer(address string, port int) *HTTPServer {
	return &HTTPServer{
		address:       address,
		port:          port,
		routes:        make(map[string][]func(*HTTPRequest, *HTTPResponse, *bool)),
		appMiddleware: []func(*HTTPRequest, *HTTPResponse, *bool){},
		workerPool:    NewWorkerPool(MAX_WORKERS),
	}
}

func (s *HTTPServer) Start() {
	// Create a new listener
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.address, s.port))
	if err != nil {
		fmt.Println("Failed to bind to port", s.port, ":", err)
		os.Exit(1)
	}
	defer listener.Close()

	fmt.Printf("Server is listening on %s:%d\n", s.address, s.port)
	// Infinite loop to continue listening after finishing one connection.
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}

		s.workerPool.Submit(func() {
			s.handleConnection(conn)
		})

	}
}

/**
 * handleConnection is used to handle the connection from the client
 * @param conn net.Conn - the connection to handle
 */
func (s *HTTPServer) handleConnection(conn net.Conn) {
	defer conn.Close()

	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		fmt.Println("Error reading from connection:", err)
		return
	}

	// Parse the requestLine, headers, body
	// GET /index.html HTTP/1.1\r\nHost: localhost:4221\r\nUser-Agent: curl/7.64.1\r\nAccept: */*\r\n\r\n

	/*
		// Request line
		GET                          // HTTP method
		/index.html                  // Request target
		HTTP/1.1                     // HTTP version
		\r\n                         // CRLF that marks the end of the request line

		// Headers
		Host: localhost:4221\r\n     // Header that specifies the server's host and port
		User-Agent: curl/7.64.1\r\n  // Header that describes the client's user agent
		Accept: \r\n              // Header that specifies which media types the client can accept
		\r\n                         // CRLF that marks the end of the headers

		// Request body (empty) in this example
	*/
	req := s.parseRequest(string(buffer[:n]))
	if req == nil {
		resp := &HTTPResponse{StatusCode: 400, Body: "Bad Request"}
		s.writeResponse(conn, resp)
		return
	}

	resp := &HTTPResponse{Headers: make(map[string]string)}
	s.routeRequest(req, resp)

	if resp.StatusCode == 0 {
		resp.StatusCode = 404
		resp.Body = "Not Found"
	}
	s.writeResponse(conn, resp)
}

/**
 * routeRequest is used to route the request to the appropriate handler
 * @param req *HTTPRequest - the request to route
 * @param resp *HTTPResponse - the response to route
 */
func (s *HTTPServer) routeRequest(req *HTTPRequest, resp *HTTPResponse) {
	pathKey := fmt.Sprintf("%s %s", req.Method, req.Path)
	// The moveAhead is for controlling the flow, when false flow of execution of handlers stop and vice-versa.
	moveAhead := true

	for _, middleware := range s.appMiddleware {
		middleware(req, resp, &moveAhead)
		if !moveAhead {
			return
		}
	}

	routeHandlers, exists := s.routes[pathKey]
	if !exists {
		return
	}

	for _, handler := range routeHandlers {
		handler(req, resp, &moveAhead)
		if !moveAhead {
			return
		}
	}
}

/**
 * AddRoute is used to add a route to the server
 * @param method string - the HTTP method
 * @param path string - the path to add the route to
 * @param handlers ...func(*HTTPRequest, *HTTPResponse, *bool) - the handlers to add to the route
 */
func (s *HTTPServer) AddRoute(method, path string, handlers ...func(*HTTPRequest, *HTTPResponse, *bool)) {
	// Don't add to same pathKey again
	pathKey := fmt.Sprintf("%s %s", method, path)
	if _, exists := s.routes[pathKey]; exists {
		log.Fatalf("route already exists: %s", pathKey)
	}
	if len(handlers) == 0 {
		log.Fatalf("no handlers provided for route: %s", pathKey)
	}
	s.routes[pathKey] = handlers

}

/**
 * Use is used to add middleware to the server
 * @param middleware func(*HTTPRequest, *HTTPResponse, *bool) - the middleware to add
 */
func (s *HTTPServer) Use(middleware func(*HTTPRequest, *HTTPResponse, *bool)) {
	s.appMiddleware = append(s.appMiddleware, middleware)
}
