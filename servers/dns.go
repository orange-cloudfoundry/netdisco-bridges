package servers

import (
	"context"
	"time"

	"github.com/miekg/dns"
	log "github.com/sirupsen/logrus"

	"github.com/orange-cloudfoundry/netdisco-bridges/models"
	"github.com/orange-cloudfoundry/netdisco-bridges/services"
)

type DNSServer struct {
	resolver *services.Resolver
	config   *models.DNSServerConfig
}

func NewDNSServer(resolver *services.Resolver, config *models.DNSServerConfig) *DNSServer {
	return &DNSServer{
		resolver: resolver,
		config:   config,
	}
}

func runDnsServer(srv *dns.Server) {
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("Failed to set udp listener %s\n", err.Error())
	}
}

func (s *DNSServer) Run(ctx context.Context) {
	entry := log.WithField("server", "dns")
	udpServer := &dns.Server{
		Addr:    s.config.Listen,
		Net:     "udp",
		Handler: s.resolver.MakeDNSHandler(true),
	}
	tcpServer := &dns.Server{
		Addr:    s.config.Listen,
		Net:     "tcp",
		Handler: s.resolver.MakeDNSHandler(false),
	}
	entry.Infof("starting udp and tcp dns server on %s", s.config.Listen)
	go runDnsServer(udpServer)
	go runDnsServer(tcpServer)
	<-ctx.Done()
	log.Info("Graceful shutdown dns server ...")
	ctxTimeout, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFunc()
	err := udpServer.ShutdownContext(ctxTimeout)
	if err != nil {
		log.Errorf("error when shutdown udp dns server: %s", err.Error())
	}

	ctxTimeout, cancelFunc = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFunc()
	err = tcpServer.ShutdownContext(ctxTimeout)
	if err != nil {
		log.Errorf("error when shutdown udp dns server: %s", err.Error())
	}
	log.Info("Finished graceful shutdown dns server ...")
}
