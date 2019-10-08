package authr

import (
	"testing"

	"github.com/stretchr/testify/mock"

	"bitbucket.org/teachingstrategies/authorization-service/internal/pkg/authrlib"
	"github.com/go-chi/chi"

	"github.com/stretchr/testify/assert"
)

type mockConfigService struct{ mock.Mock }

func (m *mockConfigService) Config() *authrlib.Config { return m.Called().Get(0).(*authrlib.Config) }
func (m *mockConfigService) IsProduction() bool       { return m.Called().Bool(0) }

func TestCreateRouter(t *testing.T) {
	mockConfigSvc := &mockConfigService{}
	mockConfigSvc.On("Config").Return(&authrlib.Config{}).Times(3)
	ctx := &authrlib.AppContext{ConfigService: mockConfigSvc,
		Logger: authrlib.NewLogger(false)}
	var err error
	ctx.NewRelicService, err = authrlib.CreateNewRelicService(ctx)
	assert.Nil(t, err)
	handler := CreateRouter(ctx)
	assert.NotNil(t, handler)
	mux := handler.(*chi.Mux)
	assert.Equal(t, 8, len(mux.Middlewares()))
	assert.Equal(t, 2, len(mux.Routes())) // /access and /health
	mockConfigSvc.AssertExpectations(t)

}
