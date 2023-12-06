package downstream

import (
	"encoding/json"
	"fmt"
	"net/http"

	di "github.com/fluffy-bunny/fluffy-dozm-di"
	contracts_config "github.com/fluffy-bunny/oidc-orchestrator/internal/contracts/config"
	contracts_downstream "github.com/fluffy-bunny/oidc-orchestrator/internal/contracts/downstream"
	resty "github.com/go-resty/resty/v2"
	"github.com/rs/zerolog/log"
)

type (
	service struct {
		config            *contracts_config.Config
		discoveryDocument *contracts_downstream.DiscoveryDocument
	}
)

var (
	stemService = (*service)(nil)
)

func init() {
	var _ contracts_downstream.IDownstreamOIDCService = (*service)(nil)
}
func ctor(config *contracts_config.Config) (*service, error) {
	discoveryUrlS := fmt.Sprintf("%s/.well-known/openid-configuration", config.DownStreamAuthority)

	// pull the discovery from the authority
	client := resty.New()
	resp, err := client.R().Get(discoveryUrlS)
	if err != nil {
		log.Error().Err(err).Msgf("ctor: %s", discoveryUrlS)
		return nil, err
	}
	if resp.StatusCode() != http.StatusOK {
		log.Error().Err(err).Msgf("ctor: %s", discoveryUrlS)
		return nil, err
	}
	body := resp.Body()
	log.Info().Msgf("ctor: %s", body)

	discoveryDocument := &contracts_downstream.DiscoveryDocument{}
	err = json.Unmarshal(body, discoveryDocument)
	if err != nil {
		log.Error().Err(err).Msgf("ctor: %s", discoveryUrlS)
		return nil, err
	}
	if err != nil {
		log.Error().Err(err).Msgf("ctor: %s", discoveryUrlS)
		return nil, err
	}
	return &service{
		config:            config,
		discoveryDocument: discoveryDocument,
	}, nil
}

// AddSingletonIDownstreamOIDCService registers the *service as a singleton.
func AddSingletonIDownstreamOIDCService(builder di.ContainerBuilder) {
	di.AddSingleton[contracts_downstream.IDownstreamOIDCService](builder, ctor)
}
func (s *service) GetDiscoveryDocument() (*contracts_downstream.DiscoveryDocument, error) {
	return s.discoveryDocument, nil
}
func (s *service) GetJWKS() (interface{}, error) {
	jwksUrl := s.discoveryDocument.JwksURI
	client := resty.New()
	resp, err := client.R().Get(jwksUrl)
	if err != nil {
		log.Error().Err(err).Msgf("GetJWKS: %s", jwksUrl)
		return nil, err
	}
	if resp.StatusCode() != http.StatusOK {
		log.Error().Err(err).Msgf("GetJWKS: %s", jwksUrl)
		return nil, err
	}
	body := resp.Body()
	log.Info().Msgf("GetJWKS: %s", body)
	var jwks interface{}
	err = json.Unmarshal(body, &jwks)
	if err != nil {
		log.Error().Err(err).Msgf("GetJWKS: %s", jwksUrl)
		return nil, err
	}
	return jwks, nil
}
