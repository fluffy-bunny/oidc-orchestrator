package downstream

import "context"

type (
	DiscoveryDocument struct {
		Issuer                      string `json:"issuer"`
		AuthorizationEndpoint       string `json:"authorization_endpoint"`
		DeviceAuthorizationEndpoint string `json:"device_authorization_endpoint"`
		TokenEndpoint               string `json:"token_endpoint"`
		UserinfoEndpoint            string `json:"userinfo_endpoint"`
		//RevocationEndpoint                string   `json:"revocation_endpoint"`
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
	AuthorizationCodeResponse struct {
		AccessToken  string `json:"access_token"`
		ExpiresIn    int    `json:"expires_in"`
		IDToken      string `json:"id_token,omitempty"`
		Scope        string `json:"scope"`
		TokenType    string `json:"token_type"`
		RefreshToken string `json:"refresh_token,omitempty"`
	}
	RefreshTokenResponse struct {
		AccessToken  string `json:"access_token"`
		ExpiresIn    int    `json:"expires_in"`
		TokenType    string `json:"token_type"`
		RefreshToken string `json:"refresh_token"`
	}
	RefreshTokenRequest struct {
		RefreshToken string `json:"refresh_token"`
		Scope        string `json:"scope"`
	}
	IDownstreamOIDCService interface {
		// GetDiscoveryDocument ...
		GetDiscoveryDocument() (*DiscoveryDocument, error)
		GetJWKS() (interface{}, error)
		ExchangeCodeForToken(ctx context.Context, basicAuth string, code string, redirectURL string) (*AuthorizationCodeResponse, error)
		RefreshToken(ctx context.Context, basicAuth string, request *RefreshTokenRequest) (*RefreshTokenResponse, error)
	}
)
