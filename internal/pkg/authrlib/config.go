package authrlib

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

	"gopkg.in/yaml.v2"
)

// Config is the collection of disparate configurations needed to run authorization service
type Config struct {
	App      AppConfig      `yaml:"app"`
	NewRelic NewRelicConfig `yaml:"newrelic"`
	MsSQL    MsSQLConfig    `yaml:"mssql"`
}

// AppConfig is the environment specific definition of this service
type AppConfig struct {
	Name       string `yaml:"name"`
	Env        string `yaml:"env"`
	Port       int    `yaml:"port"`
	ID         string `yaml:"id"`
	KeysServer string `yaml:"keys-server"`
}

// PortToStr Converts port to string
func (c AppConfig) PortToStr() string {
	return strconv.Itoa(c.Port)
}

// NewRelicConfig is our application performance monitor
// newrelic.com/<insert app id here once deployed>
type NewRelicConfig struct {
	Enabled         bool   `yaml:"enabled"`
	APIKey          string `yaml:"apikey"`
	IgnoreHTTPCodes []int  `yaml:"ignorehttpcodes"`
}

// MsSQLConfig keeps the db connection information
type MsSQLConfig struct {
	Connection Connection `yaml:"connection"`
}

type Connection struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Database string `yaml:"database"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

// ApplicationConfigService represent configuration service for authorization service
type ApplicationConfigService interface {
	Config() *Config
	IsProduction() bool
}

// dummy type which implements ApplicationConfigService and holds path to config file
type configService struct {
	configFile *string
	config     *Config
}

// NewApplicationConfigService builds new implementation of ConfigService
func NewApplicationConfigService(configFile *string) ApplicationConfigService {
	appConfigService := &configService{configFile: configFile, config: &Config{}}
	if err := initialize(configFile, appConfigService.config); err != nil {
		panic("unable to initialize configuration service: " + err.Error())
	}
	return appConfigService
}

// initialize configuration service - read data from provided yml file
func initialize(configFile *string, config *Config) error {
	if _, err := os.Stat(*configFile); os.IsNotExist(err) {
		return fmt.Errorf("unable to find application configuration file: %s", *configFile)
	}
	bytes, err := ioutil.ReadFile(*configFile)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(bytes, config)
}

func (s *configService) Config() *Config {
	return s.config
}

func (s *configService) IsProduction() bool {
	return s.config.App.Env == "production"
}
