package models

import (
	"net/url"
	"time"

	"github.com/cespare/xxhash/v2"
)

const SeparatorByte byte = 255

var separatorByteSlice = []byte{SeparatorByte}

func MustLoadLocation(name string) *time.Location {
	loc, err := time.LoadLocation(name)
	if err != nil {
		panic(err)
	}
	return loc
}

var loc = MustLoadLocation("Europe/Paris")

type SearchResponse struct {
	Metadata Metadata    `json:"metadata"`
	Result   []*Material `json:"result"`
}

type Metadata struct {
	Count  int `json:"count"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

type SearchQuery struct {
	Hostname            string    `json:"hostname" yaml:"hostname"`
	HostId              string    `json:"host_id" yaml:"host_id"`
	PFS                 string    `json:"pfs" yaml:"pfs"`
	PfsShortName        string    `json:"pfs_short_name" yaml:"pfs_short_name"`
	Machine             string    `json:"machine" yaml:"machine"`
	Env                 string    `json:"env" yaml:"env"`
	Equipment           string    `json:"equipment" yaml:"equipment"`
	Profile             string    `json:"profile" yaml:"profile"`
	ProfileID           string    `json:"profile_id" yaml:"profile_id"`
	Active              string    `json:"active" yaml:"active"`
	Etat                string    `json:"etat" yaml:"etat"`
	CreationDate        time.Time `json:"creation_date" yaml:"creation_date"`
	AllocationDate      time.Time `json:"allocation_date" yaml:"allocation_date"`
	OSType              string    `json:"os_type" yaml:"os_type"`
	OSCurrent           string    `json:"os_current" yaml:"os_current"`
	System              string    `json:"system" yaml:"system"`
	Site                string    `json:"site" yaml:"site"`
	LocationHeight      string    `json:"location_height" yaml:"location_height"`
	LocationLabel       string    `json:"location_label" yaml:"location_label"`
	LocationBuilding    string    `json:"location_building" yaml:"location_building"`
	LocationRoom        string    `json:"location_room" yaml:"location_room"`
	LocationStreet      string    `json:"location_street" yaml:"location_street"`
	LocationBay         string    `json:"location_bay" yaml:"location_bay"`
	LocationSlot        string    `json:"location_slot" yaml:"location_slot"`
	LocationFace        string    `json:"location_face" yaml:"location_face"`
	ArticleType         string    `json:"article_type" yaml:"article_type"`
	ArticleLib          string    `json:"article_lib" yaml:"article_lib"`
	ArticleSerialNumber string    `json:"article_serial_number" yaml:"article_serial_number"`
	ArticleId           string    `json:"article_id" yaml:"article_id"`

	queryId uint64 `json:"-" yaml:"-"`
}

func (q *SearchQuery) Serialize() url.Values {
	values := make(url.Values)

	if q.Hostname != "" {
		values["hostname"] = []string{q.Hostname}
	}
	if q.HostId != "" {
		values["hostId"] = []string{q.HostId}
	}
	if q.PFS != "" {
		values["pfs"] = []string{q.PFS}
	}
	if q.PfsShortName != "" {
		values["pfsShortName"] = []string{q.PfsShortName}
	}
	if q.Machine != "" {
		values["machine"] = []string{q.Machine}
	}
	if q.Env != "" {
		values["env"] = []string{q.Env}
	}
	if q.Equipment != "" {
		values["equipment"] = []string{q.Equipment}
	}
	if q.Profile != "" {
		values["profile"] = []string{q.Profile}
	}
	if q.ProfileID != "" {
		values["profileId"] = []string{q.ProfileID}
	}
	if q.Active != "" {
		values["active"] = []string{q.Active}
	}
	if q.Etat != "" {
		values["etat"] = []string{q.Etat}
	}
	if !q.CreationDate.IsZero() {
		values["creationDate"] = []string{q.CreationDate.In(loc).Format("2006-01-02 15:04:05")}
	}
	if !q.AllocationDate.IsZero() {
		values["allocationDate"] = []string{q.AllocationDate.In(loc).Format("2006-01-02 15:04:05")}
	}
	if q.OSType != "" {
		values["properties.osType"] = []string{q.OSType}
	}
	if q.OSCurrent != "" {
		values["properties.osCurrent"] = []string{q.OSCurrent}
	}
	if q.System != "" {
		values["properties.systeme"] = []string{q.System}
	}
	if q.Site != "" {
		values["site"] = []string{q.Site}
	}
	if q.LocationHeight != "" {
		values["location.height"] = []string{q.LocationHeight}
	}
	if q.LocationHeight != "" {
		values["location.label"] = []string{q.LocationLabel}
	}
	if q.LocationBuilding != "" {
		values["location.building"] = []string{q.LocationBuilding}
	}
	if q.LocationRoom != "" {
		values["location.room"] = []string{q.LocationRoom}
	}
	if q.LocationStreet != "" {
		values["location.street"] = []string{q.LocationStreet}
	}
	if q.LocationBay != "" {
		values["location.bay"] = []string{q.LocationBay}
	}
	if q.LocationSlot != "" {
		values["location.slot"] = []string{q.LocationSlot}
	}
	if q.LocationFace != "" {
		values["location.face"] = []string{q.LocationFace}
	}
	if q.ArticleType != "" {
		values["machineCardProperties.articleType"] = []string{q.ArticleType}
	}
	if q.ArticleLib != "" {
		values["machineCardProperties.articleLib"] = []string{q.ArticleLib}
	}
	if q.ArticleSerialNumber != "" {
		values["machineCardProperties.articleSerialNumber"] = []string{q.ArticleSerialNumber}
	}
	if q.ArticleId != "" {
		values["machineCardProperties.articleId"] = []string{q.ArticleId}
	}
	return values
}

func (q *SearchQuery) Id() uint64 {
	if q.queryId != 0 {
		return q.queryId
	}
	xxh := xxhash.New()
	for key, val := range q.Serialize() {
		xxh.WriteString("$" + key + "$" + val[0]) // nolint
		xxh.Write(separatorByteSlice)             // nolint
	}
	q.queryId = xxh.Sum64()
	return q.queryId
}

type SearchRequest struct {
	HostMatch              string
	ManufacturerNameMatch  string
	ManufacturerModelMatch string
	LocationMatch          string
	LayersMatch            string
	SerialMatch            string
	OsName                 string
	OsVersion              string
	MatchAll               bool
}
