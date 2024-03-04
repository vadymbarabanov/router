package router

import (
	"net/http"
	"slices"
)

type route struct {
	pattern string
	handler http.Handler
}

type Router struct {
	routes      []*route
	middlewares []func(http.Handler) http.Handler
	subrouters  []*Router
}

func NewRouter() *Router {
	return &Router{
		routes:      make([]*route, 0),
		middlewares: make([]func(http.Handler) http.Handler, 0),
		subrouters:  make([]*Router, 0),
	}
}

func (r *Router) Use(middleware func(http.Handler) http.Handler) {
	r.middlewares = append(r.middlewares, middleware)
}

func (r *Router) Handle(pattern string, handler http.Handler) {
	r.routes = append(r.routes, &route{
		pattern: pattern,
		handler: handler,
	})
}

func (r *Router) HandleFunc(pattern string, handler func(w http.ResponseWriter, r *http.Request)) {
	route := &route{
		pattern: pattern,
		handler: http.HandlerFunc(handler),
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

	r.applyHandlers(mux, make([]func(http.Handler) http.Handler, 0))

	return mux
}

func (r *Router) applyHandlers(mux *http.ServeMux, middlewares []func(http.Handler) http.Handler) {
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
