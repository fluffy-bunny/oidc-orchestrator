package main

import (
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/fluffy-bunny/fluffycore-starterkit-echo/cmd/server/docs" // docs is generated by Swag CLI, you have to import it.
	"github.com/fluffy-bunny/fluffycore-starterkit-echo/internal"
	"github.com/fluffy-bunny/fluffycore-starterkit-echo/internal/startup"
	"github.com/fluffy-bunny/fluffycore/echo/runtime"
	"github.com/rs/zerolog/log"
)

var version = "Development"

// https://github.com/swaggo/echo-swagger

// @title Swagger Example API
// @version 1.0
// @description This is a sample server Petstore server.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:9044
// @BasePath /
func main() {
	processDirectory, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		panic(err)
	}
	internal.RootFolder = processDirectory
	fmt.Println("Version:" + version)
	DumpPath("./")
	r := runtime.New(startup.NewStartup())
	err = r.Run()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to run the application")
	}
}
