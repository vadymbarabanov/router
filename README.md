# `router` 

Simple golang http router with middlewares

## Features

- [`router.Use(middleware)`](#middleware)
- [`router.Group(func(subrouter))`](#group)
- [`router.Mount(subrouter)`](#sub-router)

## Usage

### Middleware

```go
package main

import (
	"net/http"
	rt "github.com/vadymbarabanov/router"
)

func main() {
	router := rt.NewRouter()

	router.Use(func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			log.Printf("%s %s %s", r.Method, r.Proto, r.URL.Path)
			next(w, r)
		}
	})

	router.HandleFunc("GET /hello", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("world!"))
	})

	http.ListenAndServe(":4000", router.Mux())
}
```

### Group

```go
package main

import (
	"net/http"
	rt "github.com/vadymbarabanov/router"
)

func main() {
	router := rt.NewRouter()

	router.Group(func(g *rt.Router) {
		g.Use(customGroupMiddleware) // group scoped middleware

		g.HandleFunc("GET /group-route", func(w http.ResponseWriter, r *http.Request) { /* ... */ })

		g.Group(func(g *rt.Router) { /* ... */ }) // nested groups
	})

	http.ListenAndServe(":4000", router.Mux())
}
```


### Sub-router

```go
package main

import (
	"net/http"
	rt "github.com/vadymbarabanov/router"
)

func main() {
	router := rt.NewRouter()

	router.HandleFunc("GET /hello", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("world!"))
	})

	protectedRouter := rt.NewRouter()
	protectedRouter.Use(authMiddleware)

	protectedRouter.HandleFunc("GET /profile", func(w http.ResponseWriter, r *http.Request) {
		// ...
	})

	router.Mount(protectedRouter)

	http.ListenAndServe(":4000", router.Mux())
}
```

