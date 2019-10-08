package authrlib

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateNewRelicService(t *testing.T) {
	// success
	ctx := &AppContext{
		ConfigService: &configService{
			config: &Config{},
		},
	}
	service, err := CreateNewRelicService(ctx)
	assert.Nil(t, err)
	assert.NotNil(t, service)
	assert.NotNil(t, service.Application())

	// error
	ctx.ConfigService.Config().App.Name = "1;2;3;4;5"
	service, err = CreateNewRelicService(ctx)
	assert.Nil(t, service)
	assert.NotNil(t, err)
}
