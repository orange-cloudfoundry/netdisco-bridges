package models

import (
	"bytes"
	"html/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/orange-cloudfoundry/go-netdisco"
	"gopkg.in/yaml.v2"
)

type Routing struct {
	Scheme   string                 `yaml:"scheme" json:"scheme"`
	Port     int                    `yaml:"port" json:"port"`
	Host     string                 `yaml:"host" json:"host"`
	Metadata map[string]interface{} `yaml:"metadata" json:"metadata"`
	IP       string                 `yaml:"ip" json:"ip"`
}

func (r Routing) UnTemplate(device netdisco.Device) (Routing, error) {
	txt, _ := yaml.Marshal(r)
	tpl, err := template.New("").Funcs(sprig.FuncMap()).Parse(string(txt))
	if err != nil {
		return r, err
	}
	buf := &bytes.Buffer{}
	err = tpl.Execute(buf, device)
	if err != nil {
		return r, err
	}
	var finalRouting Routing
	err = yaml.Unmarshal(buf.Bytes(), &finalRouting)
	if err != nil {
		return r, err
	}
	if finalRouting.Scheme == "" {
		finalRouting.Scheme = "https"
	}
	if finalRouting.Host == "" {
		finalRouting.Host = device.DNS
	}
	finalRouting.IP = device.IP
	return finalRouting, nil
}
