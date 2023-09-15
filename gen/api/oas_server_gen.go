// Code generated by ogen, DO NOT EDIT.

package api

import (
	"context"
)

// Handler handles operations described by OpenAPI v3 specification.
type Handler interface {
	// Health implements health operation.
	//
	// Health Check.
	//
	// GET /health
	Health(ctx context.Context) (HealthRes, error)
	// ListCluster implements listCluster operation.
	//
	// List Clusters.
	//
	// GET /clusters
	ListCluster(ctx context.Context) (ListClusterRes, error)
}

// Server implements http server based on OpenAPI v3 specification and
// calls Handler to handle requests.
type Server struct {
	h Handler
	baseServer
}

// NewServer creates new Server.
func NewServer(h Handler, opts ...ServerOption) (*Server, error) {
	s, err := newServerConfig(opts...).baseServer()
	if err != nil {
		return nil, err
	}
	return &Server{
		h:          h,
		baseServer: s,
	}, nil
}
