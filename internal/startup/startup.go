package startup

import (
	di "github.com/fluffy-bunny/fluffy-dozm-di"
	fluffycore_contracts_runtime "github.com/fluffy-bunny/fluffycore/contracts/runtime"
	contracts_startup "github.com/fluffy-bunny/fluffycore/echo/contracts/startup"
	services_startup "github.com/fluffy-bunny/fluffycore/echo/services/startup"
	contracts_config "github.com/fluffy-bunny/oidc-orchestrator/internal/contracts/config"
	services_downstream "github.com/fluffy-bunny/oidc-orchestrator/internal/services/downstream"
	services_handlers_authorize "github.com/fluffy-bunny/oidc-orchestrator/internal/services/handlers/authorize"
	services_handlers_discovery "github.com/fluffy-bunny/oidc-orchestrator/internal/services/handlers/discovery"
	services_handlers_healthz "github.com/fluffy-bunny/oidc-orchestrator/internal/services/handlers/healthz"
	services_handlers_home "github.com/fluffy-bunny/oidc-orchestrator/internal/services/handlers/home"
	services_handlers_jwks "github.com/fluffy-bunny/oidc-orchestrator/internal/services/handlers/jwks"
	services_handlers_signingoogle "github.com/fluffy-bunny/oidc-orchestrator/internal/services/handlers/signingoogle"
	services_handlers_swagger "github.com/fluffy-bunny/oidc-orchestrator/internal/services/handlers/swagger"
	services_handlers_token "github.com/fluffy-bunny/oidc-orchestrator/internal/services/handlers/token"
	services_probe_database "github.com/fluffy-bunny/oidc-orchestrator/internal/services/probes/database"
	echo "github.com/labstack/echo/v4"
	log "github.com/rs/zerolog/log"
)

type (
	startup struct {
		services_startup.StartupBase
		config *contracts_config.Config
	}
)

func init() {
	var _ contracts_startup.IStartup = (*startup)(nil)
}

// GetConfigOptions ...
func (s *startup) GetConfigOptions() *fluffycore_contracts_runtime.ConfigOptions {
	return &fluffycore_contracts_runtime.ConfigOptions{
		RootConfig:  []byte(contracts_config.ConfigDefaultJSON),
		Destination: s.config,
	}
}
func NewStartup() contracts_startup.IStartup {
	myStartup := &startup{
		config: &contracts_config.Config{},
	}
	hooks := &contracts_startup.Hooks{
		PostBuildHook:   myStartup.PostBuildHook,
		PreStartHook:    myStartup.PreStartHook,
		PreShutdownHook: myStartup.PreShutdownHook,
	}
	myStartup.AddHooks(hooks)
	return myStartup
}

// ConfigureServices ...
func (s *startup) ConfigureServices(builder di.ContainerBuilder) error {
	s.SetOptions(&contracts_startup.Options{
		Port: s.config.Port,
	})
	services_probe_database.AddSingletonIProbe(builder)
	services_downstream.AddSingletonIDownstreamOIDCService(builder)
	s.addAppHandlers(builder)
	di.AddInstance[*contracts_config.Config](builder, s.config)
	return nil
}

func (s *startup) PreStartHook(echo *echo.Echo) error {
	log.Info().Msg("PreStartHook")
	return nil
}
func (s *startup) PostBuildHook(container di.Container) error {
	log.Info().Msg("PostBuildHook")
	return nil
}
func (s *startup) PreShutdownHook(echo *echo.Echo) error {
	log.Info().Msg("PreShutdownHook")
	return nil
}
func (s *startup) addAppHandlers(builder di.ContainerBuilder) {
	// add your handlers here
	services_handlers_home.AddScopedIHandler(builder)
	services_handlers_healthz.AddScopedIHandler(builder)
	services_handlers_swagger.AddScopedIHandler(builder)
	services_handlers_discovery.AddScopedIHandler(builder)
	services_handlers_jwks.AddScopedIHandler(builder)
	services_handlers_authorize.AddScopedIHandler(builder)
	services_handlers_token.AddScopedIHandler(builder)
	services_handlers_signingoogle.AddScopedIHandler(builder)
}
