package authrlib

import (
	"net/http"

	"bitbucket.org/teachingstrategies/go-svc-bootstrap/health"
)

// AppContext defines application context
type AppContext struct {
	Logger                      *AppLogger
	ConfigService               ApplicationConfigService
	NewRelicService             NewRelicService
	Healthchecks                *health.HealthCheckCollection
	DbManager                   DbManager
	AccessHandler               http.HandlerFunc
	AccessValidationMiddlewares []func(next http.Handler) http.Handler
}
