package access

import (
	"bytes"
	"context"
	"database/sql"

	newrelic "github.com/newrelic/go-agent"

	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"strings"

	"github.com/go-chi/chi"
	"github.com/rs/zerolog"

	"github.com/stretchr/testify/assert"

	"bitbucket.org/teachingstrategies/authorization-service/internal/pkg/authrlib"
	"bitbucket.org/teachingstrategies/go-svc-bootstrap/authorization"
	bootstraputils "bitbucket.org/teachingstrategies/go-svc-bootstrap/utils"
	"github.com/stretchr/testify/mock"
)

// dbManagerMock dbmanager mock
type dbManagerMock struct{ mock.Mock }

func (*dbManagerMock) Db() *sql.DB { return nil }
func (*dbManagerMock) Release()    {}

// responserWriterMock
type responserWriterMock struct{ mock.Mock }

func (*responserWriterMock) Header() http.Header        { return http.Header(map[string][]string{}) }
func (*responserWriterMock) Write([]byte) (int, error)  { return -1, writeError }
func (*responserWriterMock) WriteHeader(statusCode int) {}

var writeError = errors.New("write response error")

func TestNewAccessHandler(t *testing.T) {
	dbManager := &dbManagerMock{}
	ctx := &authrlib.AppContext{DbManager: dbManager}
	handler := NewAccessHandler(ctx)
	assert.NotNil(t, handler)
}

func TestReplyAccessError(t *testing.T) {
	const userID = 42

	testCases := []struct {
		name       string
		err        error
		respWriter http.ResponseWriter
		logs       []string
		status     int
	}{
		{
			name: "user not found",
			err:  errNotFound, respWriter: httptest.NewRecorder(), status: http.StatusNotFound,
			logs: []string{`{"level":"warn","error":"user not found","user_id":42,"message":"not found"}`}},
		{
			name: "user not found (unable to reply)",
			err:  errNotFound, respWriter: &responserWriterMock{}, status: 0,
			logs: []string{`{"level":"warn","error":"user not found","user_id":42,"message":"not found"}`,
				`{"level":"warn","error":"write response error","message":"unable to reply: user not found"}`}},
		{
			name: "user not allowed",
			err:  errNotAllowed, respWriter: httptest.NewRecorder(), status: http.StatusForbidden,
			logs: []string{`{"error":"user not allowed", "level":"warn", "message":"forbidden", "user_id":42}`}},
		{
			name: "user not allowed (unable to reply)",
			err:  errNotAllowed, respWriter: &responserWriterMock{}, status: 0,
			logs: []string{`{"error":"user not allowed", "level":"warn", "message":"forbidden", "user_id":42}`,
				`{"level":"warn","error":"write response error","message":"unable to reply: forbidden"}`}},
		{
			name: "unable to fetch access data",
			err:  errors.New("other error"), respWriter: httptest.NewRecorder(), status: http.StatusInternalServerError,
			logs: []string{`{"level":"warn","error":"unable to fetch access data for userID 42; err=other error","user_id":42,"message":"unable to fetch access data"}`}},
		{
			name: "unable to fetch access data (unable to reply)",
			err:  errors.New("other error"), respWriter: &responserWriterMock{}, status: 0,
			logs: []string{`{"level":"warn","error":"unable to fetch access data for userID 42; err=other error","user_id":42,"message":"unable to fetch access data"}`,
				`{"level":"warn","error":"unable to fetch access data for userID 42; err=other error","message":"unable to reply: unable to fetch access data"}`}},
	}

	for _, testCase := range testCases {
		var buf bytes.Buffer
		logger := &authrlib.AppLogger{Logger: zerolog.New(&buf)}

		replyAccessError(userID, testCase.err, logger, testCase.respWriter)

		// check logs
		shouldContinue := false
		for i, l := range strings.Split(buf.String(), "\n") {
			if l != "" {
				if result := assert.JSONEq(t, testCase.logs[i], l); !result {
					t.Log("test case failed:", testCase.name)
					shouldContinue = true
					break
				}
			}
		}
		if shouldContinue {
			continue
		}

		if testCase.status > 0 {
			respRec := testCase.respWriter.(*httptest.ResponseRecorder)
			if result := assert.Equal(t, testCase.status, respRec.Code); !result {
				t.Log("test case failed:", testCase.name)
				continue
			}
		}

		t.Log("test case ok:", testCase.name)
	}
}

type serviceMock struct {
	mock.Mock
}

func (m *serviceMock) Access(userID int) (*authorization.Access, error) {
	args := m.Called(userID)
	val := args.Get(0)
	if val == nil {
		return nil, args.Error(1)
	}
	return val.(*authorization.Access), args.Error(1)
}

func TestHandle(t *testing.T) {
	testCases := []struct {
		userID     string
		service    Service
		status     int
		respWriter http.ResponseWriter
		validate   func(http.ResponseWriter, string)
	}{
		{userID: "108a", respWriter: httptest.NewRecorder(),
			service: &serviceMock{},
			status:  http.StatusInternalServerError,
			validate: func(w http.ResponseWriter, logs string) {
				assert.Equal(t, `{"level":"error","rid":"","user_id":"108a","message":"userID has incorrect value"}`, strings.TrimSpace(logs))
			}},
		{userID: "108a", respWriter: &responserWriterMock{}, status: 0,
			service: &serviceMock{},
			validate: func(w http.ResponseWriter, logs string) {
				split := strings.Split(logs, "\n")
				assert.Equal(t, `{"level":"error","rid":"","user_id":"108a","message":"userID has incorrect value"}`, split[0])
				assert.Equal(t, `{"level":"warn","rid":"","error":"write response error","message":"unable to reply: userID has incorrect value"}`, split[1])
			}},
		{userID: "108", respWriter: httptest.NewRecorder(),
			service: func() Service {
				m := &serviceMock{}
				m.On("Access", 108).Return(nil, errors.New("service failed")).Once()
				return m
			}(),
			status: http.StatusInternalServerError,
			validate: func(w http.ResponseWriter, logs string) {
				assert.Equal(t, `{"level":"warn","rid":"","error":"unable to fetch access data for userID 108; err=service failed","user_id":108,"message":"unable to fetch access data"}`, strings.TrimSpace(logs))
			}},
		{userID: "108", respWriter: httptest.NewRecorder(),
			service: func() Service {
				m := &serviceMock{}
				m.On("Access", 108).Return(&authorization.Access{SuperUser: true}, nil).Once()
				return m
			}(),
			status:   http.StatusOK,
			validate: func(w http.ResponseWriter, logs string) {}},
		{userID: "108", respWriter: &responserWriterMock{},
			service: func() Service {
				m := &serviceMock{}
				m.On("Access", 108).Return(&authorization.Access{SuperUser: true}, nil).Once()
				return m
			}(),
			status: 0,
			validate: func(w http.ResponseWriter, logs string) {
				assert.Equal(t, `{"level":"warn","rid":"","error":"write response error","message":"unable to reply success"}`, strings.TrimSpace(logs))
			}},
	}

	for i, testCase := range testCases {
		var buf bytes.Buffer
		logger := &authrlib.AppLogger{Logger: zerolog.New(&buf)}
		ctx := &authrlib.AppContext{Logger: logger}

		handlerFunc := (&accessHandler{ctx: ctx, service: testCase.service}).handlerFunc()

		w := testCase.respWriter
		r := httptest.NewRequest("GET", "/test/"+testCase.userID, nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams = chi.RouteParams{Keys: []string{"userID"}, Values: []string{testCase.userID}}
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		txn := (*newrelicApp()).StartTransaction("/test/"+testCase.userID, w, r)
		r = r.WithContext(context.WithValue(r.Context(), bootstraputils.ContextKey("txn"), txn))

		handlerFunc.ServeHTTP(w, r)

		if testCase.status > 0 {
			respRec := testCase.respWriter.(*httptest.ResponseRecorder)
			assert.Equal(t, testCase.status, respRec.Code)
		}

		testCase.validate(w, buf.String())

		t.Logf("Test case %d: OK", i+1)
	}

}

func newrelicApp() *newrelic.Application {
	newrelicConfig := newrelic.NewConfig("testapp - test", "1234567890123456789012345678901234567890")
	newrelicConfig.Enabled = false
	application, _ := newrelic.NewApplication(newrelicConfig)
	return &application
}
