package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	uuid "github.com/google/uuid"
	zerolog "github.com/rs/zerolog"
	rp "github.com/zitadel/oidc/v2/pkg/client/rp"
	httphelper "github.com/zitadel/oidc/v2/pkg/http"
	z_oidc "github.com/zitadel/oidc/v2/pkg/oidc"
)

var (
	callbackPath = "/auth/callback"
	key          = []byte("test1234test1234")
)

var scopes = []string{
	"openid",
	"profile",
	"email",
}

func main() {
	ctx := context.Background()
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	// create a logger and add it to the context
	logz := zerolog.New(os.Stdout).With().Caller().Timestamp().Logger()

	ctx = logz.WithContext(ctx)
	log := zerolog.Ctx(ctx)

	issuer := "https://accounts.google.com"
	issuer = "http://localhost:9044"
	clientID := "1096301616546-edbl612881t7rkpljp3qa3juminskulo.apps.googleusercontent.com"
	clientSecret := "gOKwmN181CgsnQQDWqTSZjFs"

	//scopes = append(scopes, fmt.Sprintf(orgid_scope_template, orgID))
	// NOTE: we don't need orgid scope because the primary domain will do the propery redirect.
	port := "5556"
	redirectURI := fmt.Sprintf("http://localhost:%v%v", port, callbackPath)

	envData := struct {
		ClientID    string   `json:"client_id"`
		Issuer      string   `json:"issuer"`
		Scopes      []string `json:"scopes"`
		RedirectURI string   `json:"redirect_uri"`
	}{clientID, issuer, scopes, redirectURI}

	log.Info().
		Interface("envData", envData).
		Msg("zitadel environment")

	keyPath := ""

	//scopes := strings.Split("openid profile email urn:zitadel:iam:org:project:id:zitadel:aud", " ")
	//scopes := strings.Split("openid profile email urn:zitadel:iam:org:project:id:zitadel:aud", " ")

	cookieHandler := httphelper.NewCookieHandler(key, key, httphelper.WithUnsecure())

	options := []rp.Option{
		rp.WithCookieHandler(cookieHandler),
		rp.WithVerifierOpts(rp.WithIssuedAtOffset(5 * time.Second)),
	}
	if clientSecret == "" {
		options = append(options, rp.WithPKCE(cookieHandler))
	}
	if keyPath != "" {
		options = append(options, rp.WithJWTProfile(rp.SignerFromKeyPath(keyPath)))
	}

	provider, err := rp.NewRelyingPartyOIDC(issuer, clientID, clientSecret, redirectURI, scopes, options...)
	if err != nil {
		log.Fatal().Err(err).Msg("error creating provider")
	}

	// generate some state (representing the state of the user in your application,
	// e.g. the page where he was before sending him to login
	state := func() string {
		return uuid.New().String()
	}

	// register the AuthURLHandler at your preferred path.
	// the AuthURLHandler creates the auth request and redirects the user to the auth server.
	// including state handling with secure cookie and the possibility to use PKCE.
	// Prompts can optionally be set to inform the server of
	// any messages that need to be prompted back to the user.
	http.Handle("/login", rp.AuthURLHandler(state, provider, rp.WithPromptURLParam("login")))

	type UserResponse struct {
		UserInfo          *z_oidc.UserInfo                      `json:"userinfo"`
		Tokens            *z_oidc.Tokens[*z_oidc.IDTokenClaims] `json:"tokens"`
		AccessTokenParsed *jwt.Token                            `json:"access_token_parsed"`
	}
	jwtParser := jwt.NewParser()
	// for demonstration purposes the returned userinfo response is written as JSON object onto response
	marshalUserinfo := func(w http.ResponseWriter, r *http.Request, tokens *z_oidc.Tokens[*z_oidc.IDTokenClaims], state string, rp rp.RelyingParty, info *z_oidc.UserInfo) {
		userResponse := UserResponse{
			UserInfo: info,
			Tokens:   tokens,
		}
		accessToken, _, err := jwtParser.ParseUnverified(tokens.AccessToken, jwt.MapClaims{})
		if err != nil {
			log.Error().Err(err).Msg("Error parsing JWT - accessToken")
		}
		userResponse.AccessTokenParsed = accessToken

		data, err := json.MarshalIndent(userResponse, "", "    ")

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(data)
	}

	// you could also just take the access_token and id_token without calling the userinfo endpoint:
	//
	// marshalToken := func(w http.ResponseWriter, r *http.Request, tokens *oidc.Tokens, state string, rp rp.RelyingParty) {
	//	data, err := json.Marshal(tokens)
	//	if err != nil {
	//		http.Error(w, err.Error(), http.StatusInternalServerError)
	//		return
	//	}
	//	w.Write(data)
	//}

	// you can also try token exchange flow
	//
	// requestTokenExchange := func(w http.ResponseWriter, r *http.Request, tokens *oidc.Tokens, state string, rp rp.RelyingParty, info oidc.UserInfo) {
	// 	data := make(url.Values)
	// 	data.Set("grant_type", string(oidc.GrantTypeTokenExchange))
	// 	data.Set("requested_token_type", string(oidc.IDTokenType))
	// 	data.Set("subject_token", tokens.RefreshToken)
	// 	data.Set("subject_token_type", string(oidc.RefreshTokenType))
	// 	data.Add("scope", "profile custom_scope:impersonate:id2")

	// 	client := &http.Client{}
	// 	r2, _ := http.NewRequest(http.MethodPost, issuer+"/oauth/token", strings.NewReader(data.Encode()))
	// 	// r2.Header.Add("Authorization", "Basic "+"d2ViOnNlY3JldA==")
	// 	r2.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	// 	r2.SetBasicAuth("web", "secret")

	// 	resp, _ := client.Do(r2)
	// 	fmt.Println(resp.Status)

	// 	b, _ := io.ReadAll(resp.Body)
	// 	resp.Body.Close()

	// 	w.Write(b)
	// }

	// register the CodeExchangeHandler at the callbackPath
	// the CodeExchangeHandler handles the auth response, creates the token request and calls the callback function
	// with the returned tokens from the token endpoint
	// in this example the callback function itself is wrapped by the UserinfoCallback which
	// will call the Userinfo endpoint, check the sub and pass the info into the callback function
	http.Handle(callbackPath, rp.CodeExchangeHandler(rp.UserinfoCallback(marshalUserinfo), provider))

	// if you would use the callback without calling the userinfo endpoint, simply switch the callback handler for:
	//
	// http.Handle(callbackPath, rp.CodeExchangeHandler(marshalToken, provider))

	lis := fmt.Sprintf("127.0.0.1:%s", port)
	log.Info().Msgf("listening on http://%s/", lis)
	log.Info().Msg("press ctrl+c to stop")
	err = http.ListenAndServe(lis, nil)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to listen and serve")
	}

}
