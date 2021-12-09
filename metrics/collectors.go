package metrics

import (
	"github.com/orange-cloudfoundry/go-netdisco"
	"github.com/prometheus/client_golang/prometheus"
)

type Collectors struct {
	nClient *netdisco.Client

	devicesDesc *prometheus.Desc
}

func NewCollectors() *Collectors {
	return &Collectors{}
}

func (c *Collectors) SetNClient(nClient *netdisco.Client) {
	c.nClient = nClient
}

func (c *Collectors) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.devicesDesc
}

//Collect implements required collect function for all promehteus collectors
func (c *Collectors) Collect(ch chan<- prometheus.Metric) {

	//Write latest value for each metric in the prometheus metric channel.
	//Note that you can pass CounterValue, GaugeValue, or UntypedValue types here.
	ch <- prometheus.MustNewConstMetric(c.devicesDesc, prometheus.CounterValue, 1)

}
