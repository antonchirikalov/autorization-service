package authr

import (
	"net/http"
	"time"

	"bitbucket.org/teachingstrategies/authorization-service/internal/pkg/authrlib"
	"bitbucket.org/teachingstrategies/go-svc-bootstrap/health"
	mw "bitbucket.org/teachingstrategies/go-svc-bootstrap/middlewares"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
)

// CreateRouter defines the exposure of the Authorization Service
// Logging and Telemetry wraps all requests
// Health Checks are used to convey the status of the gateway
func CreateRouter(ctx *authrlib.AppContext) http.Handler {
	appConfig := ctx.ConfigService.Config().App

	r := chi.NewRouter()

	// This is our universal Middleware stack
	// This set of middlewares will ALWAYS run,
	// regardless of subrouted path

	// This recovers from panics, ensuring that logging/metric collection is not lost
	r.Use(mw.Recoverer)
	// This standardizes request paths
	r.Use(middleware.StripSlashes)
	// This gives us a base timeout for requests
	r.Use(middleware.Timeout(time.Second * time.Duration(600)))

	r.Use(mw.AddContext(map[string]interface{}{"env": appConfig.Env}))
	// Metrics via NewRelic Integration
	r.Use(mw.AddProfiling(ctx.NewRelicService.Application()))
	// Adds a request id to our context so that we
	// can piece together requests
	r.Use(middleware.RequestID)
	// Log the access
	r.Use(mw.AddLogging(&(ctx.Logger.Logger), false))
	// This is a JSON API, thus set that content type for everything
	r.Use(render.SetContentType(render.ContentTypeJSON))

	// access handler with middleware
	r.Route("/access", func(r chi.Router) {
		r.With(ctx.AccessValidationMiddlewares...).Route("/{userID:^[0-9]+$}", func(r chi.Router) {
			r.Get("/", ctx.AccessHandler)
		})
	})

	// /health aggregates the status of a collection of health checks,
	// and reports back to the nagging ELB.
	r.Get("/health", health.GetServiceHealth(ctx.Healthchecks, appConfig.Name))

	return r
}
