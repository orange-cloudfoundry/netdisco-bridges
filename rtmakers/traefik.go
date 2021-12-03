package rtmakers

import (
	"fmt"
	"strings"

	"github.com/orange-cloudfoundry/netdisco-bridges/models"
)

// traefikHTTPConfiguration contains all the HTTP configuration parameters.
type traefikHTTPConfiguration struct {
	Routers  map[string]*traefikRouter  `json:"routers,omitempty" toml:"routers,omitempty" yaml:"routers,omitempty" export:"true"`
	Services map[string]*traefikService `json:"services,omitempty" toml:"services,omitempty" yaml:"services,omitempty" export:"true"`
}

// traefikService holds a service configuration (can only be of one type at the same time).
type traefikService struct {
	LoadBalancer *traefikServersLoadBalancer `json:"loadBalancer,omitempty" toml:"loadBalancer,omitempty" yaml:"loadBalancer,omitempty" export:"true"`
}

// traefikRouter holds the router configuration.
type traefikRouter struct {
	EntryPoints []string `json:"entryPoints,omitempty" toml:"entryPoints,omitempty" yaml:"entryPoints,omitempty" export:"true"`
	Service     string   `json:"service,omitempty" toml:"service,omitempty" yaml:"service,omitempty" export:"true"`
	Rule        string   `json:"rule,omitempty" toml:"rule,omitempty" yaml:"rule,omitempty"`
}

// traefikServer holds the server configuration.
type traefikServer struct {
	URL string `json:"url,omitempty" toml:"url,omitempty" yaml:"url,omitempty" label:"-"`
}

// traefikServersLoadBalancer holds the traefikServersLoadBalancer configuration.
type traefikServersLoadBalancer struct {
	Servers []traefikServer `json:"servers,omitempty" toml:"servers,omitempty" yaml:"servers,omitempty" label-slice-as-struct:"server" export:"true"`
}

type Traefik struct {
}

func NewTraefik() *Traefik {
	return &Traefik{}
}

func (t *Traefik) Convert(routes []models.Routing) (interface{}, error) {
	routers := make(map[string]*traefikRouter)
	services := make(map[string]*traefikService)

	for _, r := range routes {
		serviceName := t.serviceName(r)
		router, service := t.convertRoute(r)
		routers[serviceName] = router
		services[serviceName] = service
	}

	return struct {
		Http traefikHTTPConfiguration `json:"http"`
	}{
		Http: traefikHTTPConfiguration{
			Routers:  routers,
			Services: services,
		},
	}, nil
}

func (t *Traefik) convertRoute(route models.Routing) (*traefikRouter, *traefikService) {

	port := ""
	if route.Port > 0 {
		port = fmt.Sprintf(":%d", route.Port)
	}
	url := fmt.Sprintf("%s://%s%s", route.Scheme, route.IP, port)

	entrypoints := []string{"http"}
	if intEntryPts, ok := route.Metadata["entryPoints"]; ok {
		if newEntryPoints, ok := intEntryPts.([]string); ok {
			entrypoints = newEntryPoints
		}
	}
	return &traefikRouter{
			EntryPoints: entrypoints,
			Service:     t.serviceName(route),
			Rule:        fmt.Sprintf("Host(`%s`)", route.Host),
		},
		&traefikService{
			LoadBalancer: &traefikServersLoadBalancer{Servers: []traefikServer{
				{
					URL: url,
				},
			}},
		}
}

func (t *Traefik) serviceName(route models.Routing) string {
	return strings.Replace(route.Host, ".", "-", -1)
}
