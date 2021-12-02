package services

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"

	"github.com/miekg/dns"
	"github.com/orange-cloudfoundry/go-netdisco"
	log "github.com/sirupsen/logrus"
)

func DevicesToIPs(targets []netdisco.Device) []net.IP {
	ips := make([]net.IP, 0)
	for _, t := range targets {
		if t.IP == "" {
			continue
		}
		ips = append(ips, DeviceIP(t))
	}
	return ips
}

func DevicesToRRS(domain string, targets []netdisco.Device, queryType ...uint16) []dns.RR {
	rrs := make([]dns.RR, 0)
	for _, qtype := range queryType {
		rrs = append(rrs, MaterialsToRRSQueryType(domain, targets, qtype)...)
	}
	return rrs
}

func MaterialsToRRSQueryType(domain string, targets []netdisco.Device, queryType uint16) []dns.RR {
	domain = dns.Fqdn(domain)
	rrs := make([]dns.RR, 0)
	queryTypeStr, ok := dns.TypeToString[queryType]
	if !ok {
		log.Errorf("query type is not supported")
		return rrs
	}

	for _, target := range targets {
		if queryType == dns.TypeA && !DeviceIPIsV4(target) {
			continue
		}
		if queryType == dns.TypeAAAA && !DeviceIPIsV6(target) {
			continue
		}
		if queryType == dns.TypeSRV && target.DNS == "" {
			continue
		}
		var rr dns.RR
		var err error
		switch queryType {
		case dns.TypeSRV:
			rr, err = dns.NewRR(
				fmt.Sprintf("%s 30 IN %s 1 1 22 %s", domain, queryTypeStr, DeviceStringRR(target, queryType)),
			)
		case dns.TypeTXT:
			var b []byte
			// could not happen error here
			b, _ = json.Marshal(target) // nolint
			rr, err = dns.NewRR(
				fmt.Sprintf("%s IN %s %s", domain, queryTypeStr, base64.StdEncoding.EncodeToString(b)),
			)
		default:
			rr, err = dns.NewRR(
				fmt.Sprintf("%s 30 IN %s %s", domain, queryTypeStr, DeviceStringRR(target, queryType)),
			)
		}
		if err != nil {
			log.WithField("target", target.IP).Errorf("could not register: %s", err.Error())
			continue
		}
		rrs = append(rrs, rr)
	}
	return rrs
}

func DeviceIP(device netdisco.Device) net.IP {
	ip := net.ParseIP(device.IP)
	return ip
}

func DeviceIPIsV6(device netdisco.Device) bool {
	return DeviceIP(device).To4() == nil
}

func DeviceIPIsV4(device netdisco.Device) bool {
	return !DeviceIPIsV6(device)
}

func DeviceStringRR(device netdisco.Device, queryType uint16) string {
	if queryType == dns.TypeSRV && device.DNS != "" {
		return device.DNS + "."
	}
	return DeviceIP(device).String()
}
