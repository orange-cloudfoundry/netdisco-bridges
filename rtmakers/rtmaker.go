package rtmakers

import "github.com/orange-cloudfoundry/netdisco-bridges/models"

type RTMaker interface {
	Convert(routes []models.Routing) ([]byte, error)
}

func ConvertRoute(format string, routes []models.Routing) (interface{}, error) {
	switch format {
	case "traefik":
		return NewTraefik().Convert(routes)
	}
	return routes, nil
}
