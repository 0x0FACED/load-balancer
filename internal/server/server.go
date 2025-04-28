package server

import (
	"context"
	"net/http"
)

// один backend сервер
// настраивается через конфиг backedns

type Server struct {
	srv *http.Server
}

func New(srv *http.Server) *Server {
	return &Server{
		srv: srv,
	}
}

func (s *Server) Start() error {
	return s.srv.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}

func (s *Server) Address() string {
	return s.srv.Addr
}
