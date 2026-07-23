package consumer

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	adrf_context "github.com/free5gc/adrf/internal/context"
	"github.com/free5gc/adrf/internal/logger"
	"github.com/free5gc/openapi/models"
	Nnrf_NFManagement "github.com/free5gc/openapi/nrf/NFManagement"
)

type NrfService struct {
	clientMu sync.RWMutex
	clients  map[string]*Nnrf_NFManagement.APIClient
}

func NewNrfService() *NrfService {
	return &NrfService{
		clients: make(map[string]*Nnrf_NFManagement.APIClient),
	}
}

func (s *NrfService) getNFManagementClient(nrfUri string) *Nnrf_NFManagement.APIClient {
	if nrfUri == "" {
		return nil
	}
	s.clientMu.RLock()
	client, ok := s.clients[nrfUri]
	s.clientMu.RUnlock()
	if ok {
		return client
	}

	cfg := Nnrf_NFManagement.NewConfiguration()
	cfg.SetBasePath(nrfUri)
	client = Nnrf_NFManagement.NewAPIClient(cfg)

	s.clientMu.Lock()
	defer s.clientMu.Unlock()
	s.clients[nrfUri] = client
	return client
}

// SendRegisterNFInstance registers ADRF's NFProfile (with AdrfInfo) to NRF via PUT /nnrf-nfm/v1/nf-instances/{nfInstanceId}.
func (s *NrfService) SendRegisterNFInstance(ctx context.Context, nrfUri string, nfInstanceId string, profile *models.NrfNfManagementNfProfile) (string, *models.ProblemDetails, error) {
	client := s.getNFManagementClient(nrfUri)
	if client == nil {
		return "", nil, fmt.Errorf("empty NRF URI")
	}

	req := &Nnrf_NFManagement.RegisterNFInstanceRequest{
		NfInstanceID:             &nfInstanceId,
		NrfNfManagementNfProfile: profile,
	}

	resp, err := client.NFInstanceIDDocumentApi.RegisterNFInstance(ctx, req)
	if err != nil {
		logger.ConsLog.Errorf("RegisterNFInstance to NRF[%s] failed: %v", nrfUri, err)
		return "", nil, err
	}

	location := ""
	if resp != nil && resp.Location != "" {
		location = resp.Location
	}
	logger.ConsLog.Infof("Successfully registered ADRF NFProfile[%s] to NRF[%s]", nfInstanceId, nrfUri)
	return location, nil, nil
}

// SendUpdateNFInstance sends periodic Keep-Alive Heartbeat PATCH to NRF.
func (s *NrfService) SendUpdateNFInstance(ctx context.Context, nrfUri string, nfInstanceId string, patchJSON []byte) (*models.ProblemDetails, error) {
	client := s.getNFManagementClient(nrfUri)
	if client == nil {
		return nil, fmt.Errorf("empty NRF URI")
	}

	// Make HTTP PATCH request to NRF
	url := fmt.Sprintf("%s/nnrf-nfm/v1/nf-instances/%s", nrfUri, nfInstanceId)
	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json-patch+json")

	httpCli := &http.Client{}
	resp, err := httpCli.Do(req)
	if err != nil {
		logger.ConsLog.Warnf("UpdateNFInstance (Heartbeat) to NRF[%s] failed: %v", nrfUri, err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		logger.ConsLog.Warnf("UpdateNFInstance (Heartbeat) NRF returned status %d", resp.StatusCode)
	}
	return nil, nil
}

// SendDeregisterNFInstance sends DELETE /nnrf-nfm/v1/nf-instances/{nfInstanceId} on shutdown.
func (s *NrfService) SendDeregisterNFInstance(ctx context.Context, nrfUri string, nfInstanceId string) (*models.ProblemDetails, error) {
	client := s.getNFManagementClient(nrfUri)
	if client == nil {
		return nil, fmt.Errorf("empty NRF URI")
	}

	req := &Nnrf_NFManagement.DeregisterNFInstanceRequest{
		NfInstanceID: &nfInstanceId,
	}

	_, err := client.NFInstanceIDDocumentApi.DeregisterNFInstance(ctx, req)
	if err != nil {
		logger.ConsLog.Errorf("DeregisterNFInstance from NRF[%s] failed: %v", nrfUri, err)
		return nil, err
	}
	logger.ConsLog.Infof("Successfully deregistered ADRF NFProfile[%s] from NRF[%s]", nfInstanceId, nrfUri)
	return nil, nil
}

// RegisterADRFProcedure initiates the background registration and periodic heartbeat loop.
func RegisterADRFProcedure(ctx context.Context, consumer *Consumer, adrfCtx *adrf_context.ADRFContext) {
	if adrfCtx.NrfUri == "" {
		logger.ConsLog.Warn("NRF URI is empty in configuration; running ADRF in standalone mode without NRF registration")
		return
	}

	profile := adrfCtx.BuildNfProfile()
	nrfService := consumer.NrfService()
	if nrfService == nil {
		return
	}

	_, _, err := nrfService.SendRegisterNFInstance(ctx, adrfCtx.NrfUri, adrfCtx.NfId, profile)
	if err != nil {
		logger.ConsLog.Warnf("Initial NRF registration failed (%s): %v. Will retry during heartbeat loop", adrfCtx.NrfUri, err)
	}

	// Periodic Heartbeat Routine
	go func() {
		ticker := time.NewTicker(time.Duration(adrfCtx.HeartbeatInterval) * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				// Deregister on shutdown
				deregCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				_, _ = nrfService.SendDeregisterNFInstance(deregCtx, adrfCtx.NrfUri, adrfCtx.NfId)
				cancel()
				return
			case <-ticker.C:
				hbCtx, hbCancel := context.WithTimeout(ctx, 5*time.Second)
				patchJSON := []byte(`[{"op":"replace","path":"/nfStatus","value":"REGISTERED"}]`)
				_, _ = nrfService.SendUpdateNFInstance(hbCtx, adrfCtx.NrfUri, adrfCtx.NfId, patchJSON)
				hbCancel()
			}
		}
	}()
}
