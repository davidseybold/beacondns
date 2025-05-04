package server

import (
	"context"
	"net/http"
)

type HTTPServer struct {
	server *http.Server
}

func NewHTTPServer(s *http.Server) *HTTPServer {
	return &HTTPServer{server: s}
}

func (rs *HTTPServer) Start() error {

	if err := rs.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (rs *HTTPServer) Stop() error {

	return rs.server.Shutdown(context.Background())
}
