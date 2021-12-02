package servers

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"

	"github.com/orange-cloudfoundry/netdisco-bridges/models"
	"github.com/orange-cloudfoundry/netdisco-bridges/services"
)

type HTTPServer struct {
	resolver *services.Resolver
	config   *models.HTTPServerConfig
	mux      *mux.Router
}

func NewHTTPServer(resolver *services.Resolver, config *models.HTTPServerConfig) *HTTPServer {
	return &HTTPServer{
		resolver: resolver,
		config:   config,
		mux:      mux.NewRouter(),
	}
}

func (s *HTTPServer) listRoutes(w http.ResponseWriter, req *http.Request) {

	format := mux.Vars(req)["format"]
	if format == "" {
		format = req.URL.Query().Get("format")
	}
	domain := mux.Vars(req)["domain"]
	w.Header().Set("Content-Type", "application/json")
	result, err := s.resolver.GetEntryRoutes(format, domain)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(result) //nolint
}

func (s *HTTPServer) listEntries(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.resolver.GetEntries()) //nolint
}

func (s *HTTPServer) listDevices(w http.ResponseWriter, req *http.Request) {
	domain := mux.Vars(req)["domain"]
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.resolver.ResolveDevices(domain)) //nolint
}

func (s *HTTPServer) Run(ctx context.Context) {
	subRouter := s.mux.PathPrefix("/api/v1").Subrouter()
	subRouter.HandleFunc("/entries/*/routes", s.listRoutes)
	subRouter.HandleFunc("/entries/*/routes/{format}", s.listRoutes)
	subRouter.HandleFunc("/entries/{domain}/routes", s.listRoutes)
	subRouter.HandleFunc("/entries", s.listEntries)
	subRouter.HandleFunc("/entries/{domain}/devices", s.listDevices)
	listener, err := s.makeListener()
	if err != nil {
		log.Fatal(err.Error())
	}
	srv := &http.Server{Handler: s.mux}
	go func() {
		if err = srv.Serve(listener); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %+s\n", err)
		}
	}()
	<-ctx.Done()

	ctxTimeout, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFunc()
	log.Info("Graceful shutdown http server ...")
	err = srv.Shutdown(ctxTimeout)
	if err != nil {
		log.Errorf("error when shutdown udp dns server: %s", err.Error())
	}
	log.Info("Finished graceful shutdown http server.")
}

func (s *HTTPServer) makeListener() (net.Listener, error) {
	listenAddr := s.config.Listen
	if !s.config.EnableSSL {
		log.Infof("Listen %s without tls ...", listenAddr)
		return net.Listen("tcp", listenAddr)
	}
	log.Infof("Listen %s with tls ...", listenAddr)
	rootCAs, err := x509.SystemCertPool()
	if err != nil {
		rootCAs = nil
	}
	certif, err := s.config.TLSPem.BuildCertif()
	if err != nil {
		return nil, err
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{certif},
		ClientCAs:    rootCAs,
	}

	tlsConfig.BuildNameToCertificate()
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return nil, err
	}
	return tls.NewListener(listener, tlsConfig), nil
}
