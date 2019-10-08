package access

import (
	"context"
	"net/http"
	"testing"

	"net/http/httptest"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/go-chi/chi"

	"github.com/stretchr/testify/assert"

	"bitbucket.org/teachingstrategies/authorization-service/internal/pkg/authrlib"
	"bitbucket.org/teachingstrategies/go-svc-bootstrap/authorization"
	"github.com/stretchr/testify/mock"
)

type configServiceMock struct{ mock.Mock }

func (m *configServiceMock) Config() *authrlib.Config {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*authrlib.Config)
}
func (m *configServiceMock) IsProduction() bool { return m.Called().Bool(0) }

func TestMiddlewares(t *testing.T) {
	config := &authrlib.Config{App: authrlib.AppConfig{KeysServer: "http://notrealuri:3333/dev"}}
	mockConfigService := &configServiceMock{}
	mockConfigService.On("Config").Return(config).Once()
	ctx := &authrlib.AppContext{ConfigService: mockConfigService}
	validationMiddlewares := NewAccessValidationMiddlewares(ctx)
	assert.Equal(t, 2, len(validationMiddlewares))
	mockConfigService.AssertExpectations(t)
}

func TestValidateClaims(t *testing.T) {
	middlware := &accessValidationMiddleware{}
	testCases := []struct {
		name   string
		userID string
		sub    interface{}
		err    string
	}{
		{name: "subscription not found", userID: "", sub: nil, err: authorization.ErrSubNotFound.Error()},                                                   //sub doen't exist
		{name: "subscription not found (wrong format)", userID: "", sub: 108, err: authorization.ErrSubNotFound.Error()},                                    // sub has wrong type
		{name: "userID path param is not a number", userID: "abc", sub: "108", err: "internal server error: strconv.Atoi: parsing \"abc\": invalid syntax"}, // {userID} path param is not a number
		{name: "subscription <> userID", userID: "1008", sub: "108", err: authorization.ErrSubInvalid.Error()},                                              //sub <> userID
		{name: "success case", userID: "1008", sub: "1008", err: ""},                                                                                        // success
	}

	for _, testCase := range testCases {
		req := createRequest(testCase.userID)
		claims := map[string]interface{}{}
		if testCase.sub != nil {
			claims["sub"] = testCase.sub
		}
		if _, err := middlware.validateClaims(jwt.MapClaims(claims), req); testCase.err != "" && (err == nil || err.Error() != testCase.err) {
			t.Fatalf("test case failed: '%s' [expected '%s'; got '%v']", testCase.name, testCase.err, err)
		}
		t.Log("test case ok:", testCase.name)
	}
}

func createRequest(userID string) *http.Request {
	request := httptest.NewRequest("GET", "/test", nil)
	rctx := chi.NewRouteContext()
	if userID != "" {
		rctx.URLParams = chi.RouteParams{Keys: []string{"userID"}, Values: []string{userID}}
	}
	request = request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, rctx))
	return request
}
