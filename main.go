package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/orange-cloudfoundry/go-netdisco"
	"github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/orange-cloudfoundry/netdisco-bridges/models"
	"github.com/orange-cloudfoundry/netdisco-bridges/servers"
	"github.com/orange-cloudfoundry/netdisco-bridges/services"
)

var (
	configFile = kingpin.Flag("config", "Configuration File").Default("config.yml").Short('c').String()
)

func main() {
	kingpin.Version("bgp-blacklister")
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	cnf, err := models.LoadConfig(*configFile)
	if err != nil {
		logrus.Fatal(err.Error())
		return
	}

	nClient := netdisco.NewClient(
		cnf.Netdisco.Endpoint,
		cnf.Netdisco.Username,
		cnf.Netdisco.Password,
		cnf.Netdisco.InsecureSkipVerify,
	)

	resolver := services.NewResolver(
		cnf.Entries,
		nClient,
		cnf.Workers.NbWorkers,
		time.Duration(cnf.Workers.RefreshInterval),
	)

	ctx, cancelResolver := context.WithCancel(context.Background())

	go func(ctx context.Context) {
		logrus.Info("resolver service started")
		resolver.RunWorkers(ctx)
	}(ctx)

	resolver.WaitWarmup()

	toWait := 2
	if cnf.HTTPServer.Disable {
		toWait--
	}
	if cnf.DNSServer.Disable {
		toWait--
	}
	wg := &sync.WaitGroup{}
	if toWait > 0 {
		wg.Add(toWait)
	}

	if !cnf.DNSServer.Disable {
		dnsServer := servers.NewDNSServer(resolver, cnf.DNSServer)
		go func(ctx context.Context, wg *sync.WaitGroup) {
			defer wg.Done()
			logrus.Info("dns server started")
			dnsServer.Run(ctx)
		}(ctx, wg)
	}

	if !cnf.HTTPServer.Disable {
		httpServer := servers.NewHTTPServer(resolver, cnf.HTTPServer)
		go func(ctx context.Context, wg *sync.WaitGroup) {
			defer wg.Done()
			logrus.Info("http server started")
			httpServer.Run(ctx)
		}(ctx, wg)
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	s := <-sig
	cancelResolver()
	logrus.Infof("Signal (%v) received, stopping\n", s)
	if toWait > 0 {
		wg.Wait()
	}
}
