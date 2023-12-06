package downstream

type (
	DiscoveryDocument struct {
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
	IDownstreamOIDCService interface {
		// GetDiscoveryDocument ...
		GetDiscoveryDocument() (*DiscoveryDocument, error)
		GetJWKS() (interface{}, error)
	}
)
