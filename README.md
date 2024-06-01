# `net/http.ServeMux` wrapper with middlewares utils

`router` is `net/http.ServeMux` wrapper which provides several methods to organize middlewares.

## Example Usage

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

	// middlewares
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// ...
			next.ServeHTTP(w, r)
		})
	})

	router.Group(func(g *rt.Router) {
		// group scope middlewares
		g.Use(customGroupMiddleware)

		g.HandleFunc("GET /group-route", func(w http.ResponseWriter, r *http.Request) {
			// ...
		})

		// nested groups
		g.Group(func(g *rt.Router) {
			// ...
		})
	})

	protectedRouter := rt.NewRouter()
	protectedRouter.Use(authMiddleware)

	protectedRouter.HandleFunc("GET /profile", func(w http.ResponseWriter, r *http.Request) {
		// ...
	})

	// mount a subrouter
	router.Mount(protectedRouter)

	// create a serve mux
	http.ListenAndServe(":4000", router.Mux())
}
```

