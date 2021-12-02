package models

import (
	"net"
	"time"

	"github.com/miekg/dns"
)

type Material struct {
	HostID             int        `json:"hostId,omitempty"`
	Hostname           string     `json:"hostname,omitempty"`
	Hostgroup          string     `json:"hostgroup,omitempty"`
	HostType           string     `json:"hostType,omitempty"`
	HostUsage          string     `json:"hostUsage,omitempty"`
	HostAdditionalInfo string     `json:"hostAdditionalInfo,omitempty"`
	CreationDate       string     `json:"creationDate,omitempty"`
	AllocationDate     string     `json:"allocationDate,omitempty"`
	PfsID              int        `json:"pfsId,omitempty"`
	Pfs                string     `json:"pfs,omitempty"`
	PfsShortName       string     `json:"pfsShortName,omitempty"`
	PfsCriticity       string     `json:"pfsCriticity,omitempty"`
	PfsDomain          string     `json:"pfsDomain,omitempty"`
	MachineID          int        `json:"machineId,omitempty"`
	Machine            string     `json:"machine,omitempty"`
	Equipment          string     `json:"equipment,omitempty"`
	Profile            string     `json:"profile,omitempty"`
	Rampup             int        `json:"rampup,omitempty"`
	ShutdownOrder      int        `json:"shutdownOrder,omitempty"`
	ShutdownLevel      int        `json:"shutdownLevel,omitempty"`
	Env                string     `json:"env,omitempty"`
	SiteID             int        `json:"siteId,omitempty"`
	Site               string     `json:"site,omitempty"`
	Etat               string     `json:"etat,omitempty"`
	RoomID             int        `json:"roomId,omitempty"`
	Room               string     `json:"room,omitempty"`
	Active             int        `json:"active,omitempty"`
	BackIP             string     `json:"backIP,omitempty"`
	IloType            string     `json:"iloType,omitempty"`
	IloIP              string     `json:"iloIP,omitempty"`
	CardProps          []CardProp `json:"-"`
	storedIp           net.IP     `json:"-"`
}

type CardProp struct {
	ArticleType         string `json:"articleType"`
	ArticleLib          string `json:"articleLib"`
	ArticleSerialNumber string `json:"articleSerialNumber"`
	ArticleID           int    `json:"articleId"`
}

func (m *Material) IP() net.IP {
	if len(m.storedIp) != 0 {
		return m.storedIp
	}
	m.storedIp = net.ParseIP(m.BackIP)
	return m.storedIp
}

func (m *Material) IsIPv6() bool {
	return m.IP().To4() == nil
}

func (m *Material) IsIPv4() bool {
	return !m.IsIPv6()
}

func (m *Material) StringRR(queryType uint16) string {
	if queryType == dns.TypeSRV && m.Hostname != "" {
		return m.Hostname + "."
	}
	return m.IP().String()
}

type SnmpInfo struct {
	Description      string
	EnterpriseName   string
	ObjectIdentifier string
	Uptime           time.Duration
	Contact          string
	Name             string
	Location         string
	Services         int64
}
