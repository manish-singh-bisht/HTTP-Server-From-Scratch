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
