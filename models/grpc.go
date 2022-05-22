package models

import (
	"fmt"
	"strconv"

	"github.com/orange-cloudfoundry/go-netdisco"
)

type DeviceGrpc struct {
	Name        string        `json:"name"`
	IP          string        `json:"ip"`
	MacAddress  string        `json:"macAddress"`
	DNS         string        `json:"dns"`
	Mfg         DeviceGrpcMfg `json:"mfg"`
	Os          DeviceGrpcOs  `json:"os"`
	ChassisID   string        `json:"chassisId"`
	Serial      string        `json:"serial"`
	Uptime      string        `json:"uptime"`
	Description string        `json:"description"`
	Contact     string        `json:"contact"`
	Location    string        `json:"location"`
	Layers      int           `json:"layers"`
}

type DeviceGrpcMfg struct {
	Name  string `json:"name"`
	Model string `json:"model"`
}

type DeviceGrpcOs struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

func DeviceGrpcFromNetdisco(device netdisco.Device) DeviceGrpc {
	layerInt, _ := strconv.ParseInt(device.Layers, 2, 64)
	return DeviceGrpc{
		Name:       device.Name,
		IP:         device.IP,
		MacAddress: device.Mac,
		DNS:        device.DNS,
		Mfg: DeviceGrpcMfg{
			Name:  device.Vendor,
			Model: device.Model,
		},
		Os: DeviceGrpcOs{
			Name:    device.Os,
			Version: device.OsVer,
		},
		ChassisID:   device.ChassisID,
		Serial:      device.Serial,
		Uptime:      fmt.Sprintf("%ds", int64(device.Uptime/100)),
		Description: device.Description,
		Contact:     device.Contact,
		Location:    device.Location,
		Layers:      int(layerInt),
	}
}
