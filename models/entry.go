package models

import (
	"fmt"

	"github.com/orange-cloudfoundry/go-netdisco"
)

type Entries []*Entry

type Entry struct {
	Domain  string                        `yaml:"domain" json:"domain"`
	Routing *Routing                      `yaml:"routing" json:"-"`
	Targets []*netdisco.SearchDeviceQuery `yaml:"targets" json:"targets"`
}

func (e *Entry) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type plain Entry
	err := unmarshal((*plain)(e))
	if err != nil {
		return err
	}
	if e.Domain == "" {
		return fmt.Errorf("domain must be set")
	}
	if len(e.Targets) == 0 {
		return fmt.Errorf("at least one target must be set")
	}
	return nil
}
