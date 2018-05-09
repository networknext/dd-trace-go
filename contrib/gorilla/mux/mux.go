// Package mux provides tracing functions for tracing the gorilla/mux package (https://github.com/gorilla/mux).
package mux

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/networknext/dd-trace-go/contrib/internal"
	"github.com/networknext/dd-trace-go/tracer/ext"
)

// Router registers routes to be matched and dispatches a handler.
type Router struct {
	*mux.Router
	config *routerConfig
}

// NewRouterWithTracer returns a new router instance traced with the global tracer.
func NewRouter(opts ...RouterOption) *Router {
	cfg := new(routerConfig)
	defaults(cfg)
	for _, fn := range opts {
		fn(cfg)
	}
	cfg.tracer.SetServiceInfo(cfg.serviceName, "gorilla/mux", ext.AppTypeWeb)
	return &Router{
		Router: mux.NewRouter(),
		config: cfg,
	}
}

// ServeHTTP dispatches the request to the handler
// whose pattern most closely matches the request URL.
// We only need to rewrite this function to be able to trace
// all the incoming requests to the underlying multiplexer
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var (
		match mux.RouteMatch
		route = "unknown"
	)
	// get the resource associated to this request
	if r.Match(req, &match) && match.Route != nil {
		if r, err := match.Route.GetPathTemplate(); err == nil {
			route = r
		}
	}
	resource := req.Method + " " + route
	internal.TraceAndServe(r.Router, w, req, r.config.serviceName, resource, r.config.tracer)
}
