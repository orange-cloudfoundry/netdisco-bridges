package services

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/miekg/dns"
	"github.com/orange-cloudfoundry/go-netdisco"
	log "github.com/sirupsen/logrus"

	"github.com/orange-cloudfoundry/netdisco-bridges/models"
	"github.com/orange-cloudfoundry/netdisco-bridges/rtmakers"
)

type Resolver struct {
	entries              models.Entries
	nClient              *netdisco.Client
	entriesCacheResolve  *sync.Map
	netdiscoResolveCache *sync.Map
	warmedUp             bool
	warmupChan           chan bool
	tickWorker           time.Duration
	nbWorkers            int
}

type netdiscoResolved struct {
	Devices    []netdisco.Device
	ExpireWhen time.Time
}

func NewResolver(entries models.Entries, nClient *netdisco.Client, nbWorkers int, tickWorker time.Duration) *Resolver {
	return &Resolver{
		entries:              entries,
		nClient:              nClient,
		entriesCacheResolve:  &sync.Map{},
		netdiscoResolveCache: &sync.Map{},
		tickWorker:           tickWorker,
		nbWorkers:            nbWorkers,
		warmupChan:           make(chan bool, 1),
	}
}

func (r *Resolver) GetEntries() models.Entries {
	return r.entries
}

func (r *Resolver) GetEntryRoutes(format string, domain string) (interface{}, error) {
	routes := make([]models.Routing, 0)

	for _, e := range r.entries {
		if e.Routing == nil {
			continue
		}
		if domain != "" && e.Domain != domain {
			continue
		}
		for _, d := range r.DevicesFromEntry(e) {
			route, err := e.Routing.UnTemplate(d)
			if err != nil {
				return nil, err
			}

			routes = append(routes, route)
		}
	}
	return rtmakers.ConvertRoute(format, routes)
}

func (r *Resolver) DevicesFromEntry(entry *models.Entry) []netdisco.Device {
	rawMaterials, ok := r.entriesCacheResolve.Load(entry.Domain)
	if !ok {
		return r.resolveFromNetdisco(entry.Domain)
	}
	return rawMaterials.([]netdisco.Device)
}

func (r *Resolver) Resolve(domain string, queryType uint16) []dns.RR {
	return DevicesToRRS(domain, r.ResolveDevices(domain), queryType)
}

func (r *Resolver) ResolveDevices(domain string) []netdisco.Device {
	if domain == "" {
		return []netdisco.Device{}
	}
	rawMaterials, ok := r.entriesCacheResolve.Load(domain)
	if !ok {
		return r.resolveFromNetdisco(domain)
	}
	return rawMaterials.([]netdisco.Device)
}

func (r *Resolver) resolveFromNetdisco(domain string) []netdisco.Device {
	nrRaw, ok := r.netdiscoResolveCache.Load(domain)
	if ok && nrRaw.(*netdiscoResolved).ExpireWhen.After(time.Now()) {
		return nrRaw.(*netdiscoResolved).Devices
	}
	devices, err := r.nClient.SearchDevice(&netdisco.SearchDeviceQuery{
		DNS:      domain,
		Matchall: false,
	})
	if err != nil {
		log.Errorf("error when searching on netdisco dns entry: %s", err.Error())
		return nil
	}
	r.netdiscoResolveCache.Store(domain, &netdiscoResolved{
		Devices:    devices,
		ExpireWhen: time.Now().Add(r.tickWorker),
	})
	return devices
}

func (r *Resolver) MakeDNSHandler(inUdp bool) dns.Handler {
	return dns.HandlerFunc(func(w dns.ResponseWriter, msg *dns.Msg) {
		m := new(dns.Msg)
		m.SetReply(msg)
		m.Compress = true
		rrs := make([]dns.RR, 0)
		log.Debugf("receive request for with question: \n %s", msg.String())

		for _, question := range msg.Question {
			domain := strings.TrimSuffix(question.Name, ".")
			rrs = append(rrs, r.Resolve(domain, question.Qtype)...)
		}
		m.Answer = append(m.Answer, rrs...)

		// if in udp we check if we truncate to handle big answer and make dns client use tcp instead of udp to retrieve all
		if inUdp {
			m.Truncate(dns.DefaultMsgSize)
		}
		err := w.WriteMsg(m)
		if err != nil {
			log.Errorf("error writing dns response: %s", err.Error())
		}
	})
}

func (r *Resolver) RunWorkers(ctx context.Context) {
	ticker := time.NewTicker(r.tickWorker)
	go func() {
		for {
			// do first tick
			if !r.warmedUp {
				r.dispatchWorker()
				ticker.Reset(r.tickWorker)
			}
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				r.cleanNetdiscoResolved()
				r.dispatchWorker()
				ticker.Reset(r.tickWorker)
			}
		}
	}()
	ticker.Stop()
}

func (r *Resolver) cleanNetdiscoResolved() {
	keys := make([]string, 0)
	now := time.Now()
	r.netdiscoResolveCache.Range(func(key, value interface{}) bool {
		if value.(*netdiscoResolved).ExpireWhen.After(now) {
			return true
		}
		keys = append(keys, key.(string))
		return true
	})
	for _, key := range keys {
		r.netdiscoResolveCache.Delete(key)
	}
}

func (r *Resolver) dispatchWorker() {
	if !r.warmedUp {
		log.WithField("nb_entries", len(r.entries)).Info("Warming up entries from netdisco ...")
	}
	wg := &sync.WaitGroup{}
	wg.Add(r.nbWorkers)
	entryJobs := make(chan *models.Entry, len(r.entries))

	for w := 0; w < r.nbWorkers; w++ {
		go r.loadEntryWorker(entryJobs, wg)
	}

	for _, entry := range r.entries {
		entryJobs <- entry
	}
	close(entryJobs)
	wg.Wait()
	if !r.warmedUp {
		r.warmedUp = true
		r.warmupChan <- true
		log.WithField("nb_entries", len(r.entries)).Info("Finished warming up entries from netdisco.")
	}
}

func (r *Resolver) WaitWarmup() {
	if r.warmedUp {
		return
	}
	<-r.warmupChan
}

func (r *Resolver) loadEntryWorker(entries <-chan *models.Entry, wg *sync.WaitGroup) {
	defer wg.Done()

	for entry := range entries {
		entryLog := log.WithField("entry_domain", entry.Domain)
		entryLog.Debug("Loading entry from netdisco ...")
		devices, err := r.searchDevicesByEntry(entry)
		if err != nil {
			entryLog.Errorf("devices could not be retrieved: %s", err.Error())
			continue
		}
		r.entriesCacheResolve.Store(entry.Domain, devices)
		entryLog.Debug("Finished loading entry from netdisco.")
	}
}

func (r *Resolver) searchDevicesByEntry(entry *models.Entry) ([]netdisco.Device, error) {
	devices := make([]netdisco.Device, 0)
	for _, target := range entry.Targets {
		newDevices, err := r.nClient.SearchDevice(target)
		if err != nil {
			return nil, err
		}
		devices = append(devices, newDevices...)
	}
	return r.filterDuplicateDevices(devices), nil
}

func (r *Resolver) filterDuplicateDevices(devices []netdisco.Device) []netdisco.Device {
	finalDevices := make([]netdisco.Device, 0)
	for _, device := range devices {
		hasDevice := false
		for _, finalDevice := range finalDevices {
			if device.IP == finalDevice.IP {
				hasDevice = true
				break
			}
		}
		if !hasDevice {
			finalDevices = append(finalDevices, device)
		}
	}
	return finalDevices
}

func (r *Resolver) SearchDeviceByRequest(req *models.SearchRequest) ([]netdisco.Device, error) {
	query := &netdisco.SearchDeviceQuery{
		SeeAllColumns: true,
	}
	filterAfter := false
	if req.HostMatch != "" {
		query.Q = req.HostMatch + "%"
		filterAfter = true
	}
	if req.SerialMatch != "" {
		if req.HostMatch == "" {
			query.Q = "%"
		}
		filterAfter = true
	}
	// if q is not empty and you don't need to match all, we do not need to filter after
	if !req.MatchAll && req.HostMatch != "" {
		filterAfter = false
	}
	// if q is empty and you need to match all, we do not need to filter after
	if req.Empty() {
		query.Q = "%"
		filterAfter = false
	}

	if !filterAfter && req.ManufacturerModelMatch != "" {
		query.Model = req.ManufacturerModelMatch
	}
	if !filterAfter && req.ManufacturerNameMatch != "" {
		query.Vendor = req.ManufacturerNameMatch
	}
	if !filterAfter && req.LocationMatch != "" {
		query.Location = req.LocationMatch
	}
	if !filterAfter && req.LayersMatch != "" {
		query.Layers = req.LayersMatch
	}
	if !filterAfter && req.OsName != "" {
		query.OS = req.OsName
	}
	if !filterAfter && req.OsVersion != "" {
		query.OSVer = req.OsVersion
	}
	devices, err := r.nClient.SearchDevice(query)
	if err != nil {
		return nil, err
	}
	if !filterAfter {
		return devices, nil
	}
	finalDevices := make([]netdisco.Device, 0)
	for _, device := range devices {
		if r.deviceMatchRequest(device, req) {
			finalDevices = append(finalDevices, device)
		}
	}
	return finalDevices, nil
}

func (r *Resolver) deviceMatchRequest(device netdisco.Device, req *models.SearchRequest) bool {
	// not a big fan of what I've done here, refactor if you have time
	if req.ManufacturerNameMatch != "" {
		if strings.Contains(strings.ToLower(device.Vendor), strings.ToLower(req.ManufacturerNameMatch)) {
			if !req.MatchAll {
				return true
			}
		} else {
			return false
		}
	}
	if req.ManufacturerModelMatch != "" {
		if strings.Contains(strings.ToLower(device.Model), strings.ToLower(req.ManufacturerModelMatch)) {
			if !req.MatchAll {
				return true
			}
		} else {
			return false
		}
	}
	if req.LocationMatch != "" {
		if strings.Contains(strings.ToLower(device.Location), strings.ToLower(req.LocationMatch)) {
			if !req.MatchAll {
				return true
			}
		} else {
			return false
		}
	}
	if req.LayersMatch != "" {
		if strings.Contains(strings.ToLower(device.Layers), strings.ToLower(req.LayersMatch)) {
			if !req.MatchAll {
				return true
			}
		} else {
			return false
		}
	}
	if req.SerialMatch != "" {
		if strings.Contains(strings.ToLower(device.Serial), strings.ToLower(req.SerialMatch)) {
			if !req.MatchAll {
				return true
			}
		} else {
			return false
		}
	}
	if req.OsName != "" {
		if strings.Contains(strings.ToLower(device.Os), strings.ToLower(req.OsName)) {
			if !req.MatchAll {
				return true
			}
		} else {
			return false
		}
	}
	if req.OsVersion != "" {
		if strings.Contains(strings.ToLower(device.OsVer), strings.ToLower(req.OsVersion)) {
			if !req.MatchAll {
				return true
			}
		} else {
			return false
		}
	}
	return true
}
