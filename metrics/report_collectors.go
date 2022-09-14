package metrics

import (
	"sync"

	"github.com/orange-cloudfoundry/go-netdisco"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

type ReportsCollectors struct {
	nClient *netdisco.Client

	noDns             *prometheus.GaugeVec
	dnsMismatch       *prometheus.GaugeVec
	portAdminDown     *prometheus.GaugeVec
	portErrorDisabled *prometheus.GaugeVec

	portsInUse    *prometheus.GaugeVec
	portsShutdown *prometheus.GaugeVec
	portsCounts   *prometheus.GaugeVec
	portsFree     *prometheus.GaugeVec
	nodesIpCount  *prometheus.GaugeVec
	vlanMismatch  *prometheus.GaugeVec
}

func NewReportsCollectors(nClient *netdisco.Client) *ReportsCollectors {
	return &ReportsCollectors{
		nClient: nClient,
		noDns: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   "netdisco_reports",
				Subsystem:   "device",
				Name:        "addr_no_dns",
				Help:        "Netdisco device no dns set reports, constant to '1' value if in report.",
				ConstLabels: prometheus.Labels{},
			},
			[]string{
				"name",
				"ip",
				"dns",
			},
		),
		dnsMismatch: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   "netdisco_reports",
				Subsystem:   "device",
				Name:        "dns_mismatch",
				Help:        "Netdisco device dns_mismatch set reports, constant to '1' value if in report.",
				ConstLabels: prometheus.Labels{},
			},
			[]string{
				"name",
				"ip",
				"dns",
			},
		),
		portAdminDown: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   "netdisco_reports",
				Subsystem:   "port",
				Name:        "admin_down",
				Help:        "Netdisco device port admin down reports, constant to '1' value if in report.",
				ConstLabels: prometheus.Labels{},
			},
			[]string{
				"name",
				"ip",
				"dns",
				"port",
				"description",
				"up_admin",
			},
		),
		portErrorDisabled: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   "netdisco_reports",
				Subsystem:   "port",
				Name:        "error_disabled",
				Help:        "Netdisco device port admin down reports, constant to '1' value if in report.",
				ConstLabels: prometheus.Labels{},
			},
			[]string{
				"name",
				"ip",
				"dns",
				"port",
				"reason",
			},
		),
		portsCounts: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   "netdisco_reports",
				Subsystem:   "device",
				Name:        "ports_count",
				Help:        "Number of ports for a netdisco device.",
				ConstLabels: prometheus.Labels{},
			},
			[]string{
				"ip",
				"dns",
			},
		),
		portsInUse: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   "netdisco_reports",
				Subsystem:   "device",
				Name:        "ports_in_use",
				Help:        "Number of ports in use for a netdisco device.",
				ConstLabels: prometheus.Labels{},
			},
			[]string{
				"ip",
				"dns",
			},
		),
		portsShutdown: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   "netdisco_reports",
				Subsystem:   "device",
				Name:        "ports_shutdown",
				Help:        "Number of ports shutdown for a netdisco device.",
				ConstLabels: prometheus.Labels{},
			},
			[]string{
				"ip",
				"dns",
			},
		),
		portsFree: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   "netdisco_reports",
				Subsystem:   "device",
				Name:        "ports_free",
				Help:        "Number of ports free for a netdisco device.",
				ConstLabels: prometheus.Labels{},
			},
			[]string{
				"ip",
				"dns",
			},
		),
		nodesIpCount: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   "netdisco_reports",
				Subsystem:   "node",
				Name:        "ip_count",
				Help:        "Number of ips for a netdisco node.",
				ConstLabels: prometheus.Labels{},
			},
			[]string{
				"name",
				"dns",
				"mac",
				"vendor",
				"port",
				"switch",
			},
		),
		vlanMismatch: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   "netdisco_reports",
				Subsystem:   "vlan",
				Name:        "mismatch",
				Help:        "Netdisco vlan mismatch report, constant to '1' value if in report.",
				ConstLabels: prometheus.Labels{},
			},
			[]string{
				"right_device",
				"right_port",
				"left_device",
				"left_port",
			},
		),
	}
}

func (c *ReportsCollectors) Describe(ch chan<- *prometheus.Desc) {
	c.noDns.Describe(ch)
	c.dnsMismatch.Describe(ch)
	c.portsCounts.Describe(ch)
	c.portsFree.Describe(ch)
	c.portsShutdown.Describe(ch)
	c.portsInUse.Describe(ch)
	c.nodesIpCount.Describe(ch)
	c.portAdminDown.Describe(ch)
	c.portErrorDisabled.Describe(ch)
	c.vlanMismatch.Describe(ch)
}

// Collect implements required collect function for all promehteus collectors
func (c *ReportsCollectors) Collect(ch chan<- prometheus.Metric) {
	c.noDns.Reset()
	c.dnsMismatch.Reset()
	c.portsCounts.Reset()
	c.portsFree.Reset()
	c.portsShutdown.Reset()
	c.portsInUse.Reset()
	c.nodesIpCount.Reset()
	c.portAdminDown.Reset()
	c.portErrorDisabled.Reset()
	c.vlanMismatch.Reset()

	wg := &sync.WaitGroup{}
	wg.Add(8)

	go func() {
		defer wg.Done()
		devices, err := c.nClient.ReportsDeviceAddrNoDns()
		if err != nil {
			log.Errorf("Error report addr no dns: %s", err.Error())
			return
		}
		for _, d := range devices {
			c.noDns.WithLabelValues(d.Name, d.IP, d.DNS).Set(1)
		}
	}()

	go func() {
		defer wg.Done()
		devices, err := c.nClient.ReportsDeviceDnsMismatch()
		if err != nil {
			log.Errorf("Error report dns mismatch: %s", err.Error())
			return
		}
		for _, d := range devices {
			c.dnsMismatch.WithLabelValues(d.Name, d.IP, d.DNS).Set(1)
		}
	}()

	go func() {
		defer wg.Done()
		devices, err := c.nClient.ReportsPortAdminDown()
		if err != nil {
			log.Errorf("Error report port admin down: %s", err.Error())
			return
		}
		for _, d := range devices {
			c.dnsMismatch.WithLabelValues(d.Name, d.IP, d.DNS).Set(1)
		}
	}()

	go func() {
		defer wg.Done()
		ports, err := c.nClient.ReportsDevicePortUtilization(nil)
		if err != nil {
			log.Errorf("Error report port utilization: %s", err.Error())
			return
		}
		for _, p := range ports {
			c.portsCounts.WithLabelValues(p.IP, p.DNS).Set(float64(p.PortCount))
			c.portsFree.WithLabelValues(p.IP, p.DNS).Set(float64(p.PortsFree))
			c.portsShutdown.WithLabelValues(p.IP, p.DNS).Set(float64(p.PortsShutdown))
			c.portsInUse.WithLabelValues(p.IP, p.DNS).Set(float64(p.PortsInUse))
		}
	}()

	go func() {
		defer wg.Done()
		nodes, err := c.nClient.ReportsNodeMultiIps()
		if err != nil {
			log.Errorf("Error report node multi ips: %s", err.Error())
			return
		}
		for _, n := range nodes {
			c.nodesIpCount.WithLabelValues(n.Name, n.DNS, n.Mac, n.Vendor, n.Port, n.Switch).Set(float64(n.IPCount))
		}
	}()

	go func() {
		defer wg.Done()
		ports, err := c.nClient.ReportsPortAdminDown()
		if err != nil {
			log.Errorf("Error report port admin down: %s", err.Error())
			return
		}
		for _, p := range ports {
			c.portAdminDown.WithLabelValues(p.Name, p.IP, p.DNS, p.Port, p.Description, p.UpAdmin).Set(float64(1))
		}
	}()

	go func() {
		defer wg.Done()
		ports, err := c.nClient.ReportsPortErrorDisabled()
		if err != nil {
			log.Errorf("Error report ports error disabled: %s", err.Error())
			return
		}
		for _, p := range ports {
			c.portErrorDisabled.WithLabelValues(p.Name, p.IP, p.DNS, p.Port, p.Reason).Set(float64(1))
		}
	}()

	go func() {
		defer wg.Done()
		portVlans, err := c.nClient.ReportsPortVlanMismatch()
		if err != nil {
			log.Errorf("Error report  vlan mismatch: %s", err.Error())
			return
		}
		for _, p := range portVlans {
			c.vlanMismatch.WithLabelValues(p.RightDevice, p.RightPort, p.LeftDevice, p.LeftPort).Set(float64(1))
		}
	}()

	wg.Wait()

	c.noDns.Collect(ch)
	c.dnsMismatch.Collect(ch)
	c.portsCounts.Collect(ch)
	c.portsFree.Collect(ch)
	c.portsShutdown.Collect(ch)
	c.portsInUse.Collect(ch)
	c.nodesIpCount.Collect(ch)
	c.portAdminDown.Collect(ch)
	c.portErrorDisabled.Collect(ch)
	c.vlanMismatch.Collect(ch)

}
