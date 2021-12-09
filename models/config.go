package models

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"time"

	pmodel "github.com/prometheus/common/model"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type DNSServerConfig struct {
	Disable bool   `yaml:"disable"`
	Listen  string `yaml:"listen"`
}

func (c *DNSServerConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type plain DNSServerConfig
	err := unmarshal((*plain)(c))
	if err != nil {
		return err
	}
	if c.Disable {
		return nil
	}
	if c.Listen == "" {
		c.Listen = "0.0.0.0:53"
	}
	return nil
}

type Config struct {
	DNSServer             *DNSServerConfig  `yaml:"dns_server"`
	HTTPServer            *HTTPServerConfig `yaml:"http_server"`
	Entries               Entries           `yaml:"entries"`
	Netdisco              *NetdiscoConfig   `yaml:"netdisco"`
	Workers               *WorkersConfig    `yaml:"workers"`
	Log                   *Log              `yaml:"log"`
	DisableReportsMetrics bool              `yaml:"disable_reports_metrics"`
}

func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type plain Config
	err := unmarshal((*plain)(c))
	if err != nil {
		return err
	}
	if len(c.Entries) == 0 {
		return fmt.Errorf("you must set at least one entry")
	}
	if c.Netdisco == nil {
		return fmt.Errorf("netdisco config must be set")
	}
	if c.DNSServer == nil {
		c.DNSServer = &DNSServerConfig{
			Listen: "0.0.0.0:53",
		}
	}
	if c.HTTPServer == nil {
		c.HTTPServer = &HTTPServerConfig{
			Listen: "0.0.0.0:8080",
		}
	}
	if c.Workers == nil {
		c.Workers = &WorkersConfig{
			NbWorkers:       5,
			RefreshInterval: pmodel.Duration(25 * time.Minute),
		}
	}

	return nil
}

type HTTPServerConfig struct {
	Disable   bool   `yaml:"disable"`
	Listen    string `yaml:"listen"`
	EnableSSL bool   `yaml:"enable_ssl"`
	TLSPem    TLSPem `yaml:"tls_pem"`
}

func (c *HTTPServerConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type plain HTTPServerConfig
	err := unmarshal((*plain)(c))
	if err != nil {
		return err
	}
	if c.Disable {
		return nil
	}
	if c.Listen == "" && !c.EnableSSL {
		c.Listen = "0.0.0.0:8080"
	}
	if c.Listen == "" && c.EnableSSL {
		c.Listen = "0.0.0.0:8443"
	}
	return nil
}

type WorkersConfig struct {
	NbWorkers       int             `yaml:"nb_workers"`
	RefreshInterval pmodel.Duration `yaml:"refresh_interval"`
}

func (c *WorkersConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type plain WorkersConfig
	err := unmarshal((*plain)(c))
	if err != nil {
		return err
	}
	if c.NbWorkers <= 0 {
		c.NbWorkers = 5
	}
	if c.RefreshInterval <= 0 {
		c.RefreshInterval = pmodel.Duration(25 * time.Minute)
	}
	return nil
}

type NetdiscoConfig struct {
	Endpoint           string `yaml:"endpoint"`
	Username           string `yaml:"username"`
	Password           string `yaml:"password"`
	ApiKey             string `yaml:"api_key"`
	InsecureSkipVerify bool   `yaml:"insecure_skip_verify"`
}

func (c *NetdiscoConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type plain NetdiscoConfig
	err := unmarshal((*plain)(c))
	if err != nil {
		return err
	}
	if c.Endpoint == "" {
		return fmt.Errorf("endpoint to netdisco must be set")
	}

	return nil
}

type TLSPem struct {
	CertChain  string `yaml:"cert_chain"`
	PrivateKey string `yaml:"private_key"`
}

func (t TLSPem) BuildCertif() (tls.Certificate, error) {
	if t.PrivateKey == "" || t.CertChain == "" {
		return tls.Certificate{}, fmt.Errorf("Error parsing PEM blocks of router.tls_pem, missing cert or key.")
	}

	certificate, err := tls.X509KeyPair([]byte(t.CertChain), []byte(t.PrivateKey))
	if err != nil {
		errMsg := fmt.Sprintf("Error loading key pair: %s", err.Error())
		return tls.Certificate{}, fmt.Errorf(errMsg)
	}
	return certificate, nil
}

type Log struct {
	Level   string `yaml:"level"`
	NoColor bool   `yaml:"no_color"`
	InJson  bool   `yaml:"in_json"`
}

func (c *Log) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type plain Log
	err := unmarshal((*plain)(c))
	if err != nil {
		return err
	}
	log.SetFormatter(&log.TextFormatter{
		DisableColors: c.NoColor,
	})
	if c.Level != "" {
		lvl, err := log.ParseLevel(c.Level)
		if err != nil {
			return err
		}
		log.SetLevel(lvl)
	}
	if c.InJson {
		log.SetFormatter(&log.JSONFormatter{})
	}

	return nil
}

func LoadConfig(path string) (Config, error) {
	var cnf Config
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	err = yaml.Unmarshal(b, &cnf)
	if err != nil {
		return Config{}, err
	}
	if len(cnf.Entries) == 0 {
		return Config{}, fmt.Errorf("you must set at least one entry")
	}
	return cnf, nil
}
