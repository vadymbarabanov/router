package router

import (
	"net/http"
	"slices"
)

type route struct {
	pattern string
	handler http.HandlerFunc
}

type Router struct {
	routes      []*route
	middlewares []func(http.HandlerFunc) http.HandlerFunc
	subrouters  []*Router
}

func NewRouter() *Router {
	return &Router{
		routes:      make([]*route, 0),
		middlewares: make([]func(http.HandlerFunc) http.HandlerFunc, 0),
		subrouters:  make([]*Router, 0),
	}
}

func (r *Router) Use(middleware func(http.HandlerFunc) http.HandlerFunc) {
	r.middlewares = append(r.middlewares, middleware)
}

func (r *Router) Handle(pattern string, handler http.Handler) {
	r.routes = append(r.routes, &route{
		pattern: pattern,
		handler: handler.ServeHTTP,
	})
}

func (r *Router) HandleFunc(pattern string, handler http.HandlerFunc) {
	route := &route{
		pattern: pattern,
		handler: handler,
	}
	r.routes = append(r.routes, route)
}

func (r *Router) Group(callback func(g *Router)) {
	subrouter := NewRouter()
	callback(subrouter)
	r.Mount(subrouter)
}

func (r *Router) Mount(subrouter *Router) {
	r.subrouters = append(r.subrouters, subrouter)
}

func (r *Router) Mux() *http.ServeMux {
	mux := http.NewServeMux()

	r.applyHandlers(mux, make([]func(http.HandlerFunc) http.HandlerFunc, 0))

	return mux
}

func (r *Router) applyHandlers(mux *http.ServeMux, middlewares []func(http.HandlerFunc) http.HandlerFunc) {
	middlewares = slices.Concat(middlewares, r.middlewares)

	for _, route := range r.routes {
		for i := len(middlewares) - 1; i >= 0; i -= 1 {
			route.handler = middlewares[i](route.handler)
		}

		mux.Handle(route.pattern, route.handler)
	}

	for _, subrouter := range r.subrouters {
		subrouter.applyHandlers(mux, middlewares)
	}
}
