package main

import (
	"flag"
	"net/http"

	"bitbucket.org/teachingstrategies/authorization-service/internal/app/authr"
	"bitbucket.org/teachingstrategies/authorization-service/internal/app/authr/access"
	"bitbucket.org/teachingstrategies/authorization-service/internal/pkg/authrlib"
	"bitbucket.org/teachingstrategies/go-svc-bootstrap/health"

	_ "github.com/denisenkom/go-mssqldb"
)

func main() {
	configFile := flag.String("config", "./config.yml", "config file path")
	flag.Parse()

	ctx := new(authrlib.AppContext)

	ctx.ConfigService = authrlib.NewApplicationConfigService(configFile)
	ctx.Logger = authrlib.NewLogger(ctx.ConfigService.IsProduction())

	ctx.Logger.Info().Msg("initializing authorization service")

	// initialize new relic
	newRelic, err := authrlib.CreateNewRelicService(ctx)
	if err != nil {
		ctx.Logger.Fatal().Err(err).Msg("unable to initialize new relic")
	}
	ctx.NewRelicService = newRelic

	// Init healthchecks
	ctx.Healthchecks = health.NewHealthCheckCollection()

	// initialize db
	ctx.DbManager, err = authrlib.NewDbManager(ctx)
	if err != nil {
		ctx.Logger.Fatal().Err(err).Msg("unable to connect mssql")
	}
	defer ctx.DbManager.Release()

	// handler for /access router ( should be moved to router ?)
	ctx.AccessHandler = access.NewAccessHandler(ctx)
	ctx.AccessValidationMiddlewares = access.NewAccessValidationMiddlewares(ctx)

	// Initialize router
	router := authr.CreateRouter(ctx)

	port := ctx.ConfigService.Config().App.PortToStr()
	ctx.Logger.Info().Str("port", port).Msg("starting server")

	if err = http.ListenAndServe(":"+port, router); err != nil {
		ctx.Logger.Fatal().Err(err).Msg("unable to start a server")
	}
}
