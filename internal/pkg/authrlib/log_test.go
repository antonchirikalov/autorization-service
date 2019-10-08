package authrlib

import (
	"bytes"
	"context"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"
	"unsafe"

	bootstraputils "bitbucket.org/teachingstrategies/go-svc-bootstrap/utils"
	"github.com/go-chi/chi"
	newrelic "github.com/newrelic/go-agent"

	"github.com/rs/zerolog"

	"github.com/stretchr/testify/assert"
)

func TestHandlerLog(t *testing.T) {
	//verify hook added
	appLogger := NewLogger(true)
	rs := reflect.ValueOf(&appLogger.Logger).Elem()
	rf := rs.FieldByName("hooks")
	rf = reflect.NewAt(rf.Type(), unsafe.Pointer(rf.UnsafeAddr())).Elem()
	hooks, _ := rf.Interface().([]zerolog.Hook)
	assert.Equal(t, 1, len(hooks))

	// check out
	var buf bytes.Buffer
	logger := &AppLogger{Logger: zerolog.New(&buf)}
	r := httptest.NewRequest("GET", "/test", nil)
	rctx := chi.NewRouteContext()
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
	txn := (*newrelicApp()).StartTransaction("/test", httptest.NewRecorder(), r)
	r = r.WithContext(context.WithValue(r.Context(), bootstraputils.ContextKey("txn"), txn))

	logger.HandlerLogger(r).Info().Str("test", "testme").Msg("req_test")
	assert.Equal(t, `{"level":"info","rid":"","test":"testme","message":"req_test"}`, strings.TrimSpace(buf.String()))
}

func newrelicApp() *newrelic.Application {
	newrelicConfig := newrelic.NewConfig("testapp - test", "1234567890123456789012345678901234567890")
	newrelicConfig.Enabled = false
	application, _ := newrelic.NewApplication(newrelicConfig)
	return &application
}

func TestNewLogger(t *testing.T) {
	// prod
	appLogger := NewLogger(true)
	w := extractWriter(appLogger)
	_, ok := w.(*os.File)
	assert.True(t, ok)
	// non prod
	appLogger = NewLogger(false)
	w = extractWriter(appLogger)
	_, ok = w.(zerolog.ConsoleWriter)
	assert.True(t, ok)
}

func extractWriter(appLogger *AppLogger) interface{} {
	rs := reflect.ValueOf(&appLogger.Logger).Elem()
	rf := rs.FieldByName("w")
	rf = reflect.NewAt(rf.Type(), unsafe.Pointer(rf.UnsafeAddr())).Elem()
	lw := rf.Interface()

	rs = reflect.ValueOf(&lw).Elem().Elem()
	rf = rs.Field(0)
	return rf.Interface()
}
