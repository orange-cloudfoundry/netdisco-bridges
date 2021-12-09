package metrics

import (
	"strconv"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/orange-cloudfoundry/netdisco-bridges/services"
)

type DeviceCollectors struct {
	resolver *services.Resolver
	domains  []string

	deviceInfo *prometheus.GaugeVec
}

func NewDeviceCollectors(resolver *services.Resolver, domains []string) *DeviceCollectors {
	return &DeviceCollectors{
		resolver: resolver,
		domains:  domains,
		deviceInfo: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   "netdisco",
				Subsystem:   "device",
				Name:        "info",
				Help:        "Netdisco device information for prometheus usage, constant to '1' value.",
				ConstLabels: prometheus.Labels{},
			},
			[]string{
				"domain",
				"name",
				"ip",
				"mac",
				"dns",
				"os",
				"os_ver",
				"layers",
				"uptime_age",
				"model",
				"vendor",
				"serial",
				"num_ports",
				"slots",
				"location",
				"creation",
				"last_discover",
				"contact",
			},
		),
	}
}

func (c *DeviceCollectors) Describe(ch chan<- *prometheus.Desc) {
	c.deviceInfo.Describe(ch)
}

//Collect implements required collect function for all promehteus collectors
func (c *DeviceCollectors) Collect(ch chan<- prometheus.Metric) {
	c.deviceInfo.Reset()
	for _, domain := range c.domains {
		devices := c.resolver.ResolveDevices(domain)
		for _, d := range devices {
			c.deviceInfo.WithLabelValues(domain,
				d.Name,
				d.IP,
				d.Mac,
				d.DNS,
				d.Os,
				d.OsVer,
				d.Layers,
				d.UptimeAge,
				d.Model,
				d.Vendor,
				d.Serial,
				strconv.Itoa(d.NumPorts),
				strconv.Itoa(d.Slots),
				d.Location,
				d.Creation,
				d.LastDiscover,
				d.Contact).Set(1)
		}
	}
	c.deviceInfo.Collect(ch)

}
