package router_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	rt "github.com/vadymbarabanov/router"
)

type route struct {
	pattern    string
	handler    http.HandlerFunc
	statusCode int
	headers    map[string]string
	body       string
}

func TestRouter_Middlewares(t *testing.T) {
	testCases := []struct {
		testname    string
		middlewares []func(http.HandlerFunc) http.HandlerFunc
		routes      []route
	}{
		{
			testname: "middleware applies header",
			middlewares: []func(http.HandlerFunc) http.HandlerFunc{
				func(next http.HandlerFunc) http.HandlerFunc {
					return func(w http.ResponseWriter, r *http.Request) {
						w.Header().Set("X-Test-Middleware", "test")
						next(w, r)
					}
				},
			},
			routes: []route{{
				pattern: "/test",
				handler: func(w http.ResponseWriter, r *http.Request) {
					w.Write([]byte("test"))
				},
				statusCode: http.StatusOK,
				headers: map[string]string{
					"X-Test-Middleware": "test",
				},
				body: "test",
			}},
		},
		{
			testname: "middlewares execute in correct order",
			middlewares: []func(http.HandlerFunc) http.HandlerFunc{
				func(next http.HandlerFunc) http.HandlerFunc {
					return func(w http.ResponseWriter, r *http.Request) {
						w.Header().Set("X-Test-Middleware", "test-1")
						next(w, r)
					}
				},
				func(next http.HandlerFunc) http.HandlerFunc {
					return func(w http.ResponseWriter, r *http.Request) {
						w.Header().Set("X-Test-Middleware", "test-2")
						next(w, r)
					}
				},
			},
			routes: []route{{
				pattern: "/test",
				handler: func(w http.ResponseWriter, r *http.Request) {
					w.Write([]byte("test"))
				},
				statusCode: http.StatusOK,
				headers: map[string]string{
					"X-Test-Middleware": "test-2",
				},
				body: "test",
			}},
		},
	}

	for _, test := range testCases {
		t.Run(test.testname, func(t *testing.T) {
			router := rt.NewRouter()

			for _, testroute := range test.routes {
				router.Handle(testroute.pattern, testroute.handler)
			}

			for _, middleware := range test.middlewares {
				router.Use(middleware)
			}

			server := httptest.NewServer(router.Mux())

			for _, testroute := range test.routes {
				assertRoute(t, server, testroute)
			}
		})
	}
}

type group struct {
	middlewares []func(http.HandlerFunc) http.HandlerFunc
	routes      []route
}

func TestRouter_Groups(t *testing.T) {
	testCases := []struct {
		testname string
		routes   []route
		groups   []group
	}{
		{
			testname: "group middleware execute on group routes only",
			routes: []route{{
				pattern: "/test-1",
				handler: func(w http.ResponseWriter, r *http.Request) {
					w.Write([]byte("test-1"))
				},
				statusCode: http.StatusOK,
				body:       "test-1",
				headers: map[string]string{
					"X-Test-1": "",
				},
			}},
			groups: []group{{
				middlewares: []func(http.HandlerFunc) http.HandlerFunc{
					func(http.HandlerFunc) http.HandlerFunc {
						return func(w http.ResponseWriter, r *http.Request) {
							w.Header().Set("X-Test-1", "test-1")
							http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
							return
						}
					},
				},
				routes: []route{{
					pattern: "/test-1/header",
					handler: func(w http.ResponseWriter, r *http.Request) {
						w.Write([]byte("test-1 header"))
					},
					statusCode: http.StatusUnauthorized,
					headers: map[string]string{
						"X-Test-1": "test-1",
					},
					body: fmt.Sprintf("%s\n", http.StatusText(http.StatusUnauthorized)),
				}},
			}},
		},
	}

	for _, test := range testCases {
		t.Run(test.testname, func(t *testing.T) {
			router := rt.NewRouter()

			for _, route := range test.routes {
				router.Handle(route.pattern, route.handler)
			}

			for _, group := range test.groups {
				router.Group(func(g *rt.Router) {
					for _, middleware := range group.middlewares {
						g.Use(middleware)
					}

					for _, route := range group.routes {
						g.Handle(route.pattern, route.handler)
					}
				})
			}

			server := httptest.NewServer(router.Mux())

			for _, testroute := range test.routes {
				assertRoute(t, server, testroute)
			}

			for _, group := range test.groups {
				for _, testroute := range group.routes {
					assertRoute(t, server, testroute)
				}
			}
		})
	}
}

func assertRoute(t *testing.T, server *httptest.Server, testroute route) {
	t.Helper()

	resp, err := http.Get(fmt.Sprintf("%s%s", server.URL, testroute.pattern))
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != testroute.statusCode {
		t.Fatalf("response.StatusCode: want (%d) got (%d)", testroute.statusCode, resp.StatusCode)
	}

	if testroute.headers != nil {
		for key, value := range testroute.headers {
			if resp.Header.Get(key) != value {
				t.Fatalf("response.Header.Get(\"%s\"): want (%s) got (%s)\n", key, value, resp.Header.Get(key))
			}
		}
	}

	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	if testroute.body != string(bytes) {
		t.Fatalf("repsonse.Body: want (%s) got (%s)", testroute.body, string(bytes))
	}
}
