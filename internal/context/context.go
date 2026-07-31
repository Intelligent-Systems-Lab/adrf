package context

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/free5gc/adrf/pkg/factory"
	"github.com/free5gc/openapi/models"
)

var (
	self *ADRFContext
	once sync.Once
)

// ADRFContext holds process-wide runtime metadata.
type ADRFContext struct {
	AdrfName          string
	NfId              string
	NrfUri            string
	UriScheme         models.UriScheme
	RegisterIPv4      string
	BindingIPv4       string
	Port              int
	Locality          string
	HeartbeatInterval int
	PlmnSupportList   []factory.PlmnSupportItem
	ServiceNameList   []models.ServiceName
	StartTime         time.Time
}

type ADRFInfo struct {
	MLModelStorageInd bool `json:"mlModelStorageInd,omitempty"`
	DataStorageInd    bool `json:"dataStorageInd,omitempty"`
}

// NFProfile supplements the pinned free5GC generated model with the Release 18
// adrfInfoList field required by NRF discovery filters.
type NFProfile struct {
	models.NrfNfManagementNfProfile
	AdrfInfoList map[string]ADRFInfo `json:"adrfInfoList"`
}

// Init initializes the singleton runtime context exactly once.
func Init() {
	once.Do(func() {
		self = &ADRFContext{
			NfId:      uuid.New().String(),
			StartTime: time.Now().UTC(),
		}
	})
}

// GetSelf returns the singleton ADRF runtime context.
func GetSelf() *ADRFContext {
	return self
}

func (c *ADRFContext) InitFromConfig(cfg *factory.Config) {
	if cfg == nil || cfg.Configuration == nil {
		return
	}
	c.AdrfName = cfg.GetAdrfName()
	if cfg.Configuration.NfInstanceId != "" {
		c.NfId = cfg.Configuration.NfInstanceId
	}
	c.NrfUri = cfg.Configuration.NrfUri
	c.Locality = cfg.Configuration.Locality
	c.HeartbeatInterval = cfg.Configuration.HeartbeatInterval
	if c.HeartbeatInterval <= 0 {
		c.HeartbeatInterval = 10
	}
	c.PlmnSupportList = cfg.Configuration.PlmnSupportList
	c.ServiceNameList = cfg.Configuration.ServiceNameList

	if cfg.Configuration.Sbi != nil {
		if cfg.Configuration.Sbi.Scheme == "https" {
			c.UriScheme = models.UriScheme_HTTPS
		} else {
			c.UriScheme = models.UriScheme_HTTP
		}
		c.RegisterIPv4 = cfg.Configuration.Sbi.RegisterIPv4
		c.BindingIPv4 = cfg.Configuration.Sbi.BindingIPv4
		c.Port = cfg.Configuration.Sbi.Port
	}
	if c.RegisterIPv4 == "" {
		c.RegisterIPv4 = "127.0.0.1"
	}
	if c.Port == 0 {
		c.Port = 9888
	}
}

func (c *ADRFContext) GetIPv4Uri() string {
	return fmt.Sprintf("%s://%s:%d", c.UriScheme, c.RegisterIPv4, c.Port)
}

// BuildNfProfile constructs the 3GPP models.NrfNfManagementNfProfile for NRF Registration.
func (c *ADRFContext) BuildNfProfile() *NFProfile {
	profile := &NFProfile{
		NrfNfManagementNfProfile: models.NrfNfManagementNfProfile{
			NfInstanceId:  c.NfId,
			NfType:        models.NrfNfManagementNfType_ADRF,
			NfStatus:      models.NrfNfManagementNfStatus_REGISTERED,
			Locality:      c.Locality,
			Ipv4Addresses: []string{c.RegisterIPv4},
		},
		AdrfInfoList: map[string]ADRFInfo{
			"default": {
				MLModelStorageInd: true,
				DataStorageInd:    true,
			},
		},
	}

	var plmns []models.PlmnId
	var snssais []models.ExtSnssai
	for _, item := range c.PlmnSupportList {
		if item.PlmnId != nil {
			plmns = append(plmns, *item.PlmnId)
		}
		for _, sn := range item.SnssaiList {
			snssais = append(snssais, models.ExtSnssai{
				Sst: sn.Sst,
				Sd:  sn.Sd,
			})
		}
	}
	if len(plmns) > 0 {
		profile.NrfNfManagementNfProfile.PlmnList = plmns
	}
	if len(snssais) > 0 {
		profile.NrfNfManagementNfProfile.SNssais = snssais
	}

	// Define NF Services
	var services []models.NrfNfManagementNfService
	for idx, sName := range c.ServiceNameList {
		serv := models.NrfNfManagementNfService{
			ServiceInstanceId: fmt.Sprintf("%s-%d", sName, idx+1),
			ServiceName:       sName,
			Versions: []models.NfServiceVersion{
				{
					ApiVersionInUri: "v1",
					ApiFullVersion:  "1.0.0",
				},
			},
			Scheme:          c.UriScheme,
			NfServiceStatus: models.NfServiceStatus_REGISTERED,
			IpEndPoints: []models.IpEndPoint{
				{
					Ipv4Address: c.RegisterIPv4,
					Port:        int32(c.Port),
				},
			},
			ApiPrefix: c.GetIPv4Uri(),
		}
		services = append(services, serv)
	}
	profile.NrfNfManagementNfProfile.NfServices = services
	return profile
}
