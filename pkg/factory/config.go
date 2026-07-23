package factory

import (
	"os"

	"gopkg.in/yaml.v3"

	"github.com/free5gc/adrf/internal/logger"
	"github.com/free5gc/openapi/models"
)

const (
	AdrfDefaultConfigPath = "./config/adrfcfg.yaml"
	AdrfDefaultName       = "ADRF"

	AdrfSbiDefaultScheme = "http"
	AdrfSbiDefaultIPv4   = "127.0.0.1"
	AdrfSbiDefaultPort   = 9888

	AdrfDataManagementResUriPrefix     = "/nadrf-datamanagement/v1"
	AdrfDataStoreRecordsPath           = "/data-store-records"
	AdrfDataRetrievalSubscriptionsPath = "/data-retrieval-subscriptions"

	AdrfMLModelManagementResUriPrefix = "/nadrf-mlmodelmanagement/v1"
	AdrfMLModelStoreRecordsPath       = "/mlmodel-store-records"
)


type Config struct {
	Info          *Info          `yaml:"info"`
	Configuration *Configuration `yaml:"configuration"`
	Logger        *Logger        `yaml:"logger,omitempty"`
}

type Info struct {
	Version     string `yaml:"version,omitempty"`
	Description string `yaml:"description,omitempty"`
}

type Configuration struct {
	AdrfName          string             `yaml:"adrfName,omitempty"`
	NrfUri            string             `yaml:"nrfUri,omitempty"`
	NfInstanceId      string             `yaml:"nfInstanceId,omitempty"`
	Sbi               *Sbi               `yaml:"sbi,omitempty"`
	ServiceNameList   []models.ServiceName `yaml:"serviceNameList,omitempty"`
	PlmnSupportList   []PlmnSupportItem  `yaml:"plmnSupportList,omitempty"`
	Locality          string             `yaml:"locality,omitempty"`
	HeartbeatInterval int                `yaml:"heartbeatInterval,omitempty"`
	Mongodb           *Mongodb           `yaml:"mongodb,omitempty"`
	Retrieval         *Retrieval         `yaml:"retrieval,omitempty"`
	MLModelStorage    *MLModelStorage    `yaml:"mlModelStorage,omitempty"`
}

type PlmnSupportItem struct {
	PlmnId     *models.PlmnId  `yaml:"plmnId"`
	SnssaiList []models.Snssai `yaml:"snssaiList,omitempty"`
}

type MLModelStorage struct {
	LocalDirectory string `yaml:"localDirectory,omitempty"`
}

type Sbi struct {
	Scheme       string `yaml:"scheme"`
	RegisterIPv4 string `yaml:"registerIPv4,omitempty"`
	BindingIPv4  string `yaml:"bindingIPv4,omitempty"`
	Port         int    `yaml:"port,omitempty"`
	OAuth        bool   `yaml:"oauth,omitempty"`
	TLS          *TLS   `yaml:"tls,omitempty"`
}

type TLS struct {
	Pem string `yaml:"pem,omitempty"`
	Key string `yaml:"key,omitempty"`
}

type Mongodb struct {
	Name string `yaml:"name"`
	Url  string `yaml:"url"`
}

type Retrieval struct {
	CorrIDBatchSize int       `yaml:"corrIdBatchSize,omitempty"`
	OneIDPerFetch   bool      `yaml:"oneIdPerFetch,omitempty"`
	Snapshot        *Snapshot `yaml:"snapshot,omitempty"`
}

type Snapshot struct {
	Enabled bool `yaml:"enabled,omitempty"`
}

type Logger struct {
	Enable       bool   `yaml:"enable"`
	Level        string `yaml:"level"`
	ReportCaller bool   `yaml:"reportCaller"`
}

func ReadConfig(cfgPath string) (*Config, error) {
	if cfgPath == "" {
		cfgPath = AdrfDefaultConfigPath
	}

	cfg := &Config{}
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		logger.CfgLog.Errorf("failed to read config file: %v", err)
		return nil, err
	}
	if err = yaml.Unmarshal(data, cfg); err != nil {
		logger.CfgLog.Errorf("failed to parse config file: %v", err)
		return nil, err
	}

	cfg.setDefaults()
	logger.CfgLog.Infof("config loaded: %s", cfgPath)
	return cfg, nil
}

func (c *Config) setDefaults() {
	if c.Configuration == nil {
		c.Configuration = &Configuration{}
	}
	if c.Configuration.AdrfName == "" {
		c.Configuration.AdrfName = AdrfDefaultName
	}
	if c.Configuration.Sbi == nil {
		c.Configuration.Sbi = &Sbi{}
	}
	if c.Configuration.Sbi.Scheme == "" {
		c.Configuration.Sbi.Scheme = AdrfSbiDefaultScheme
	}
	if c.Configuration.Sbi.BindingIPv4 == "" {
		c.Configuration.Sbi.BindingIPv4 = AdrfSbiDefaultIPv4
	}
	if c.Configuration.Sbi.Port == 0 {
		c.Configuration.Sbi.Port = AdrfSbiDefaultPort
	}
	if c.Configuration.Retrieval == nil {
		c.Configuration.Retrieval = &Retrieval{}
	}
	if c.Configuration.Retrieval.CorrIDBatchSize <= 0 {
		c.Configuration.Retrieval.CorrIDBatchSize = 100
	}
	if c.Configuration.Retrieval.Snapshot == nil {
		c.Configuration.Retrieval.Snapshot = &Snapshot{Enabled: true}
	}
	if c.Configuration.MLModelStorage == nil {
		c.Configuration.MLModelStorage = &MLModelStorage{}
	}
	if c.Configuration.MLModelStorage.LocalDirectory == "" {
		c.Configuration.MLModelStorage.LocalDirectory = "./storage/models"
	}
}

func (c *Config) GetSbiScheme() string {
	if c == nil || c.Configuration == nil || c.Configuration.Sbi == nil || c.Configuration.Sbi.Scheme == "" {
		return AdrfSbiDefaultScheme
	}
	return c.Configuration.Sbi.Scheme
}

func (c *Config) GetAdrfName() string {
	if c == nil || c.Configuration == nil || c.Configuration.AdrfName == "" {
		return AdrfDefaultName
	}
	return c.Configuration.AdrfName
}
