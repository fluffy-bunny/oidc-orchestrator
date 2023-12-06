package healthz

import (
	"encoding/json"
	"fmt"
	"net/http"

	di "github.com/fluffy-bunny/fluffy-dozm-di"
	contracts_handler "github.com/fluffy-bunny/fluffycore/echo/contracts/handler"
	contracts_config "github.com/fluffy-bunny/oidc-orchestrator/internal/contracts/config"
	wellknown "github.com/fluffy-bunny/oidc-orchestrator/internal/wellknown"
	resty "github.com/go-resty/resty/v2"
	echo "github.com/labstack/echo/v4"
	log "github.com/rs/zerolog/log"
)

type (
	service struct {
		config            *contracts_config.Config
		discoveryDocument *Discovery
	}
	Discovery struct {
		Issuer                            string   `json:"issuer"`
		AuthorizationEndpoint             string   `json:"authorization_endpoint"`
		DeviceAuthorizationEndpoint       string   `json:"device_authorization_endpoint"`
		TokenEndpoint                     string   `json:"token_endpoint"`
		UserinfoEndpoint                  string   `json:"userinfo_endpoint"`
		RevocationEndpoint                string   `json:"revocation_endpoint"`
		JwksURI                           string   `json:"jwks_uri"`
		ResponseTypesSupported            []string `json:"response_types_supported"`
		SubjectTypesSupported             []string `json:"subject_types_supported"`
		IDTokenSigningAlgValuesSupported  []string `json:"id_token_signing_alg_values_supported"`
		ScopesSupported                   []string `json:"scopes_supported"`
		TokenEndpointAuthMethodsSupported []string `json:"token_endpoint_auth_methods_supported"`
		ClaimsSupported                   []string `json:"claims_supported"`
		CodeChallengeMethodsSupported     []string `json:"code_challenge_methods_supported"`
		GrantTypesSupported               []string `json:"grant_types_supported"`
	}
)

func init() {
	var _ contracts_handler.IHandler = (*service)(nil)
}

// AddScopedIHandler registers the *service as a singleton.
func AddScopedIHandler(builder di.ContainerBuilder) {
	contracts_handler.AddScopedIHandleWithMetadata[*service](builder,
		ctor,
		[]contracts_handler.HTTPVERB{
			contracts_handler.GET,
		},
		wellknown.DiscoveryPath,
	)

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

	discoveryDocument := &Discovery{}
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
func (s *service) GetMiddleware() []echo.MiddlewareFunc {
	return []echo.MiddlewareFunc{}
}

// HealthCheck godoc
// @Summary Show the status of server.
// @Description get the status of server.
// @Tags root
// @Accept */*
// @Produce json
// @Success 200 {object} string
// @Router /healthz [get]
func (s *service) Do(c echo.Context) error {
	baseUrl := "http://" + c.Request().Host
	s.discoveryDocument.JwksURI = baseUrl + wellknown.JWKSPath
	s.discoveryDocument.AuthorizationEndpoint = baseUrl + wellknown.AuthorizationPath
	s.discoveryDocument.TokenEndpoint = baseUrl + wellknown.TokenPath

	return c.JSON(http.StatusOK, s.discoveryDocument)
}
