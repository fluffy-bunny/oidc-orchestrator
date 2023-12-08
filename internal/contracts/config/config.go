package config

type (

	// Config type
	Config struct {
		ApplicationName        string `json:"applicationName" mapstructure:"APPLICATION_NAME"`
		ApplicationEnvironment string `json:"applicationEnvironment" mapstructure:"APPLICATION_ENVIRONMENT"`
		PrettyLog              bool   `json:"prettyLog" mapstructure:"PRETTY_LOG"`
		LogLevel               string `json:"logLevel" mapstructure:"LOG_LEVEL"`
		Port                   int    `json:"port" mapstructure:"PORT"`
		DownStreamAuthority    string `json:"downStreamAuthority" mapstructure:"DOWN_STREAM_AUTHORITY"`
	}
)

var (

	// ConfigDefaultJSON default json
	ConfigDefaultJSON = []byte(`
{
	"APPLICATION_NAME": "in-environment",
	"APPLICATION_ENVIRONMENT": "in-environment",
	"PRETTY_LOG": false,
	"LOG_LEVEL": "info",
	"PORT": 1111,
	"DOWN_STREAM_AUTHORITY": "https://accounts.google.com"
 }
`)
)
