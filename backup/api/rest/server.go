package rest

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/aaron70/decoy/internal/services"
	"github.com/aaron70/goaty/errors"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/gorillamux"
)

var ErrSpec = errors.New("SpecificationError")

type RestServer struct {
	Logger     *log.Logger
	Mux        *http.ServeMux
	Decoy      *services.Decoy
	specRouter routers.Router
}

func NewRestServer(w io.Writer, decoy *services.Decoy, spec io.ReadCloser) (*RestServer, error) {
	logger := log.New(w, log.Prefix(), log.Flags())
	var (
		router routers.Router
	)
	if spec != nil {
		loader := openapi3.NewLoader()
		doc, err := loader.LoadFromIoReader(spec)
		defer spec.Close()
		if err != nil {
			return nil, errors.NewError(ErrSpec, err, "Failed to load the spec")
		}
		if err = doc.Validate(loader.Context); err != nil {
			return nil, errors.NewError(ErrSpec, err, "Invalid spec")
		}

		router, err = gorillamux.NewRouter(doc)
		if err != nil {
			return nil, errors.NewError(ErrSpec, err, "Couldn't create the router for the given spec")
		}
	} else {
		logger.Printf("Server started without OpenAPI Spec, then endpoint /mock will be disabled...")
	}

	mux := http.NewServeMux()
	server := &RestServer{
		Logger:     logger,
		Mux:        mux,
		Decoy:      decoy,
		specRouter: router,
	}
	server.RegisterHandlers()
	return server, nil
}

func (s *RestServer) Listen(addr string) error {
	s.Logger.Printf("Decoy server listening at %s", addr)
	return http.ListenAndServe(addr, loggerMiddleware(s.Mux))
}

func (s *RestServer) RegisterHandlers() {
	s.Mux.HandleFunc("/mock", s.mockHandler)
	s.Mux.HandleFunc("/mock/", s.mockHandler)
}

func Start(w io.Writer, decoy *services.Decoy, port int, spec io.ReadCloser) error {
	server, err := NewRestServer(w, decoy, spec)
	if err != nil {
		return errors.NewError(nil, err, "Coudln't create the REST server")
	}

	addr := fmt.Sprintf(":%d", port)
	return server.Listen(addr)
}

func loggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		wrapped := &responseInterceptor{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		log.Printf("%s %s → %d (%s)", r.Method, r.URL.Path, wrapped.statusCode, time.Since(start))
	})
}

type responseInterceptor struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseInterceptor) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
