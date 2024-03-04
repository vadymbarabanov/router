package main

import (
	"context"
	"log"
	"net/http"

	rt "github.com/vadymbarabanov/router"
)

func main() {
	router := rt.NewRouter()

	// net/http handlers
	router.HandleFunc("GET /hello", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("world!"))
	})

	// middlewares
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Printf("%s %s %s", r.Method, r.Proto, r.URL.Path)
			next.ServeHTTP(w, r)
		})
	})

	router.Group(func(g *rt.Router) {
		// group scope middlewares
		g.Use(attachCustomHeader)

		g.HandleFunc("GET /custom-header", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Check repsonse headers!"))
		})

		// nested groups
		g.Group(func(g *rt.Router) {
			// ...
		})
	})

	protectedRouter := rt.NewRouter()
	protectedRouter.Use(authenticator)

	protectedRouter.HandleFunc("GET /profile", func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value("user_id").(string)

		w.Write([]byte(userID))
	})

	// mount a subrouter
	router.Mount(protectedRouter)

	// create a serve mux
	http.ListenAndServe(":4000", router.Mux())
}

func authenticator(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_id")
		if err != nil {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		_ = cookie // ...getting session with user id from session storage...

		r = r.WithContext(context.WithValue(r.Context(), "user_id", "1234"))

		next.ServeHTTP(w, r)
	})
}

func attachCustomHeader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Custom", "true")
		next.ServeHTTP(w, r)
	})
}
