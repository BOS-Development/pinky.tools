package web

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	log "github.com/annymsMthd/industry-tool/internal/logging"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

type Router struct {
	port       int
	router     *mux.Router
	pathAccess map[string]AuthAccess
	backendKey string
}

type HttpError struct {
	StatusCode int
	Error      error
}

type HandlerArgs struct {
	Request *http.Request
	Params  map[string]string
	User    *int64
}

type AuthAccess int

const (
	AuthAccessBackend AuthAccess = iota
	AuthAccessUser
)

func NewRouter(port int, backendKey string) *Router {
	router := mux.NewRouter().StrictSlash(true)
	r := &Router{
		router:     router,
		port:       port,
		pathAccess: map[string]AuthAccess{},
		backendKey: backendKey,
	}

	return r
}

func (r *Router) RegisterRestAPIRoute(path string, access AuthAccess, handler func(*HandlerArgs) (any, *HttpError), methods ...string) {
	f := func(w http.ResponseWriter, req *http.Request) {
		var user *int64
		var ok bool
		switch access {
		case AuthAccessBackend:
			if !r.hasBackendAccess(req) {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
		case AuthAccessUser:
			user, ok = r.hasUserAccess(req)
			if !ok {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
		default:
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		vars := mux.Vars(req)
		args := &HandlerArgs{
			Request: req,
			Params:  vars,
			User:    user,
		}
		o, httpErr := handler(args)
		if httpErr != nil {
			log.Error("http error occurred", "status_code", httpErr.StatusCode, "error", httpErr.Error.Error())
			w.WriteHeader((httpErr.StatusCode))
			w.Write([]byte(httpErr.Error.Error()))
			return
		}

		bytes, err := json.Marshal(o)
		if err != nil {
			log.Error("json format issue occurred", "error", err.Error())
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
			return
		}

		_, err = w.Write(bytes)
		if err != nil {
			log.Error("Error writing bytes for http return", "Error", err.Error())
		}
	}

	route := r.router.HandleFunc(path, f)
	if len(methods) > 0 {
		route.Methods(methods...)
	}
}

func (r *Router) RegisterMiddleware(middleware mux.MiddlewareFunc) {
	r.router.Use(middleware)
}

func (r *Router) Run(ctx context.Context) func() error {
	svr := &http.Server{
		Addr:    fmt.Sprintf(":%d", r.port),
		Handler: r.router,
	}
	cancel := make(chan error)

	go func() {
		err := svr.ListenAndServe()
		if err != nil {
			cancel <- errors.Wrap(err, "failed to serve http")
		}
	}()

	return func() error {
		select {
		case <-ctx.Done():
			sCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			svr.Shutdown(sCtx)
			return nil
		case msg := <-cancel:
			return msg
		}
	}
}

func (r *Router) hasBackendAccess(req *http.Request) bool {
	if req.Header.Get("BACKEND-KEY") != r.backendKey {
		log.Info("unauthorized keys dont match")
		return false
	}

	return true
}

func (r *Router) hasUserAccess(req *http.Request) (*int64, bool) {
	if req.Header.Get("BACKEND-KEY") != r.backendKey {
		log.Info("unauthorized keys dont match")
		return nil, false
	}
	user := req.Header.Get("USER-ID")
	if user == "" {
		return nil, false
	}

	var id *int64
	i, err := strconv.Atoi(user)
	if err != nil {
		return nil, false
	}
	i64 := int64(i)
	id = &i64

	return id, true
}
