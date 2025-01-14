# Custom HTTP Server

A simple, custom-built HTTP server in Go. This HTTP server is a lightweight, extensible HTTP server with support for middlewares, basic route handling, and multi-threading using goroutines.

[Screencast from 13-03-25 10:24:17 AM IST.webm](https://github.com/user-attachments/assets/743a8b4b-93df-4cbc-adaa-75c033e2d727)

## Features

- **Custom HTTP Server**: Built from scratch using Go.
  - The server listens for incoming connections on a specified address and port.
  - It parses incoming HTTP requests, including the request line, headers, and body.
  - It supports HTTP methods such as `GET`, `POST`, etc., and handles routing based on request method and path.
- **Routing**: Support for adding dynamic routes using HTTP methods (e.g., GET, POST).
- **Middleware**: Allows app-level and route-specific middleware for request handling.
- **Concurrency**: Supports handling multiple concurrent connections using goroutines. Uses a custom worker pool to manage the goroutines.

## How to Clone and Use

To clone and run the server locally, follow these steps:

1. **Clone the repository**:
   ```bash
   git clone https://github.com/yourusername/HTTP-Server.git
   cd HTTP-Server
   ```
2. **Initialize Go Modules:**
   ```bash
   go mod init github.com/manish-singh-bisht/HTTP-Server
   go mod tidy
   ```
3. **Build and run the server**
   ```bash
   go build -o http-server
   ./http-server
   ```

## Example

```bash
package main

import (
	"fmt"

	server "github.com/manish-singh-bisht/HTTP-Server-From-Scratch/server"
)

const (
	PORT    = 4222
	ADDRESS = "0.0.0.0"
)

func main() {
	httpServer := server.NewHTTPServer(ADDRESS, PORT)

	// App-level middleware
	httpServer.Use(func(req *server.HTTPRequest, resp *server.HTTPResponse, moveAhead *bool) {
		fmt.Println("App-level middleware: Logging request")
	})

	// Test with: curl http://localhost:4222/health
	httpServer.AddRoute("GET", "/health",
		func(req *server.HTTPRequest, resp *server.HTTPResponse, moveAhead *bool) {
			resp.StatusCode = 200
			resp.Headers["Content-Type"] = "application/json"
			resp.Body = "OK\r\n"
		})

	// Adding a route with route-specific middleware
	// Test with: curl -H "Authorization: Bearer valid-token" http://localhost:4222/secure
	// Without token: curl http://localhost:4222/secure
	httpServer.AddRoute("GET", "/secure",
		// Route-specific middleware: Authentication check
		func(req *server.HTTPRequest, resp *server.HTTPResponse, moveAhead *bool) {
			if req.Headers["Authorization"] != "Bearer valid-token" {
				resp.StatusCode = 401
				resp.Body = "Unauthorized \r\n"
				*moveAhead = false
			}
		},
		// Route handler
		func(req *server.HTTPRequest, resp *server.HTTPResponse, moveAhead *bool) {
			resp.StatusCode = 200
			resp.Body = "Welcome to the secure endpoint!\r\n"
		})

		httpServer.Start()
}


```
