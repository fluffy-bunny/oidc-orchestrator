package downstream

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	di "github.com/fluffy-bunny/fluffy-dozm-di"
	core_utils "github.com/fluffy-bunny/fluffycore/utils"
	contracts_config "github.com/fluffy-bunny/oidc-orchestrator/internal/contracts/config"
	contracts_downstream "github.com/fluffy-bunny/oidc-orchestrator/internal/contracts/downstream"
	req "github.com/imroc/req/v3"
	log "github.com/rs/zerolog/log"
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
	client := req.C()
	resp, err := client.R(). // Use R() to create a request.
					Get(discoveryUrlS)
	if err != nil {
		log.Error().Err(err).Msgf("ctor: %s", discoveryUrlS)
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		log.Error().Err(err).Msgf("ctor: %s", discoveryUrlS)
		return nil, err
	}
	body := resp.Bytes()
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
	// make a copy of the discovery document
	copy := &contracts_downstream.DiscoveryDocument{}
	*copy = *s.discoveryDocument
	return copy, nil
}
func (s *service) GetJWKS() (interface{}, error) {
	jwksUrl := s.discoveryDocument.JwksURI
	client := req.C()
	resp, err := client.R().Get(jwksUrl)
	if err != nil {
		log.Error().Err(err).Msgf("GetJWKS: %s", jwksUrl)
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		log.Error().Err(err).Msgf("GetJWKS: %s", jwksUrl)
		return nil, err
	}
	body := resp.Bytes()
	log.Info().Msgf("GetJWKS: %s", body)
	var jwks interface{}
	err = json.Unmarshal(body, &jwks)
	if err != nil {
		log.Error().Err(err).Msgf("GetJWKS: %s", jwksUrl)
		return nil, err
	}
	return jwks, nil
}

var mockIDToken = `eyJhbGciOiJFUzI1NiIsImtpZCI6IjBiMmNkMmU1NGM5MjRjZTg5ZjAxMGYyNDI4NjIzNjdkIiwidHlwIjoiSldUIn0.eyJhdWQiOiJteWF1ZCIsImVtYWlsIjoiZ2hzdGFobEBnbWFpbC5jb20iLCJleHAiOjE3MDE5Njg5MDgsImZhbWlseV9uYW1lIjoiU3RhaGwiLCJnaXZlbl9uYW1lIjoiSGVyYiIsImlhdCI6MTcwMTk2NzEwOCwiaXNzIjoiaHR0cDovL2xvY2FsaG9zdDo5MDQ0IiwibmFtZSI6IkhlcmIgU3RhaGwiLCJwZXJtaXNzaW9ucyI6WyJwZXJtaXNzaW9uLm9uZSIsInBlcm1pc3Npb24udHdvIiwicGVybWlzc2lvbi50aHJlZSJdLCJzdWIiOiIxMDQ3NTg5MjQ0MjgwMzY2NjM5NTEifQ.NMo1-LmNUZBDf55uRjOgmS7pyZXMvxehfPScReswRVhIDm3ONUU-25cGpn6Vwbha2Jq62x2BMnviW4gH-UkD0A`

func (s *service) RefreshToken(ctx context.Context, authToken string, request *contracts_downstream.RefreshTokenRequest) (*contracts_downstream.RefreshTokenResponse, error) {

	// grant_type: authorization_code
	client := req.C()

	// build a form to post
	form := map[string]string{
		"grant_type":    "refresh_token",
		"refresh_token": request.RefreshToken,
	}
	if !core_utils.IsEmptyOrNil(request.Scope) {
		form["scope"] = request.Scope
	}
	resp, err := client.R().
		SetContext(ctx).
		SetHeader("Authorization", authToken).
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetHeader("Accept", "application/json").
		SetHeader("User-Agent", "Go-http-client/1.1").
		SetFormData(form).
		Post(s.discoveryDocument.TokenEndpoint)
	if err != nil {
		log.Error().Err(err).Msgf("ExchangeCodeForToken: %s", s.discoveryDocument.TokenEndpoint)
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("status code: %d", resp.StatusCode)
		log.Error().Err(err).Msgf("ExchangeCodeForToken: %s", s.discoveryDocument.TokenEndpoint)
		return nil, err
	}
	body := resp.Bytes()
	log.Info().Msgf("ExchangeCodeForToken: %s", body)
	response := &contracts_downstream.RefreshTokenResponse{}
	err = json.Unmarshal(body, response)
	if err != nil {
		log.Error().Err(err).Msgf("ExchangeCodeForToken: %s", s.discoveryDocument.TokenEndpoint)
		return nil, err
	}
	return response, nil
}

func (s *service) ExchangeCodeForToken(ctx context.Context,
	authToken string, code string, redirectURL string) (*contracts_downstream.AuthorizationCodeResponse, error) {
	/*
			return &contracts_downstream.AuthorizationCodeResponse{
			AccessToken: mockIDToken,
			IDToken:     mockIDToken,
		}, nil
	*/

	// grant_type: authorization_code
	client := req.C()

	// build a form to post
	form := map[string]string{
		"grant_type":   "authorization_code",
		"code":         code,
		"redirect_uri": redirectURL,
	}
	resp, err := client.R().
		SetContext(ctx).
		SetHeader("Authorization", authToken).
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetHeader("Accept", "application/json").
		SetHeader("User-Agent", "Go-http-client/1.1").
		SetFormData(form).
		Post(s.discoveryDocument.TokenEndpoint)
	if err != nil {
		log.Error().Err(err).Msgf("ExchangeCodeForToken: %s", s.discoveryDocument.TokenEndpoint)
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("status code: %d", resp.StatusCode)
		log.Error().Err(err).Msgf("ExchangeCodeForToken: %s", s.discoveryDocument.TokenEndpoint)
		return nil, err
	}
	body := resp.Bytes()
	log.Info().Msgf("ExchangeCodeForToken: %s", body)
	response := &contracts_downstream.AuthorizationCodeResponse{}
	err = json.Unmarshal(body, response)
	if err != nil {
		log.Error().Err(err).Msgf("ExchangeCodeForToken: %s", s.discoveryDocument.TokenEndpoint)
		return nil, err
	}
	return response, nil

}
