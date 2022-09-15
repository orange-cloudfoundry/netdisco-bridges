package models

import (
	"fmt"
	"strconv"

	"github.com/orange-cloudfoundry/go-netdisco"
)

type DeviceGrpc struct {
	Name              string           `json:"name"`
	IP                string           `json:"ip"`
	MacAddress        string           `json:"macAddress"`
	DNS               string           `json:"dns"`
	Ports             DeviceGrpcPorts  `json:"ports"`
	SupportedServices DeviceGrpcSupSvc `json:"supportedServices"`
	Mfg               DeviceGrpcMfg    `json:"mfg"`
	Os                DeviceGrpcOs     `json:"os"`
	ChassisID         string           `json:"chassisId"`
	Serial            string           `json:"serial"`
	Uptime            string           `json:"uptime"`
	Description       string           `json:"description"`
	Contact           string           `json:"contact"`
	Location          string           `json:"location"`
	Layers            int              `json:"layers"`
}

type DeviceGrpcMfg struct {
	Name  string `json:"name"`
	Model string `json:"model"`
}

type DeviceGrpcOs struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type DeviceGrpcPorts struct {
	Ssh   int32 `json:"ssh"`
	Http  int32 `json:"http"`
	Https int32 `json:"https"`
}

type DeviceGrpcSupSvc struct {
	Ssh   bool `json:"ssh"`
	Http  bool `json:"http"`
	Https bool `json:"https"`
}

func DeviceGrpcFromNetdisco(device netdisco.Device) DeviceGrpc {
	name := device.Name
	if name == "" {
		name = device.DNS
	}
	if name == "" {
		name = device.IP
	}
	vendor := device.Vendor
	if vendor == "" {
		vendor = "Unknown"
	}
	model := device.Model
	if model == "" {
		model = "Unknown"
	}

	osName := device.Os
	if osName == "" {
		osName = "Unknown"
	}
	osVersion := device.OsVer
	if osVersion == "" {
		osVersion = "Unknown"
	}
	layerInt, _ := strconv.ParseInt(device.Layers, 2, 64)
	return DeviceGrpc{
		Name:       name,
		IP:         device.IP,
		MacAddress: device.Mac,
		DNS:        device.DNS,
		Mfg: DeviceGrpcMfg{
			Name:  vendor,
			Model: model,
		},
		Os: DeviceGrpcOs{
			Name:    osName,
			Version: osVersion,
		},
		Ports: DeviceGrpcPorts{
			Ssh:   22,
			Http:  80,
			Https: 443,
		},
		SupportedServices: DeviceGrpcSupSvc{
			Ssh:   true,
			Http:  false,
			Https: false,
		},
		ChassisID:   device.ChassisID,
		Serial:      device.Serial,
		Uptime:      fmt.Sprintf("%ds", device.Uptime/100),
		Description: device.Description,
		Contact:     device.Contact,
		Location:    device.Location,
		Layers:      int(layerInt),
	}
}
