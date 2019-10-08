package authrlib

import (
	"io"
	"net/http"
	"os"

	mw "bitbucket.org/teachingstrategies/go-svc-bootstrap/middlewares"
	bootstraputils "bitbucket.org/teachingstrategies/go-svc-bootstrap/utils"

	"github.com/go-chi/chi/middleware"
	newrelic "github.com/newrelic/go-agent"
	"github.com/rs/zerolog"
)

type AppLogger struct {
	zerolog.Logger
}

func (appLogger *AppLogger) HandlerLogger(r *http.Request) *AppLogger {
	rid := middleware.GetReqID(r.Context())

	requestLogger := appLogger.With().Str("rid", rid).Logger()

	hl := requestLogger.Hook(mw.NoticeErrorHook{
		Txn: r.Context().Value(bootstraputils.ContextKey("txn")).(newrelic.Transaction),
	})
	return &AppLogger{hl}
}

// NewLogger Creates a new instance of logger
func NewLogger(isProduction bool) *AppLogger {
	var logWriter io.Writer = os.Stdout
	if !isProduction {
		logWriter = zerolog.ConsoleWriter{Out: logWriter}
	}
	return &AppLogger{zerolog.New(logWriter).With().Timestamp().Logger()}
}
