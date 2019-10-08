package authrlib

import (
	newrelic "github.com/newrelic/go-agent"
)

// NewRelicService represents interface
// for interaction with new relic
type NewRelicService interface {
	Application() newrelic.Application
}

type newRelicService struct {
	application newrelic.Application
	ctx         *AppContext
}

// CreateNewRelicService instantiates NewRelicService instance based on confuration
func CreateNewRelicService(ctx *AppContext) (NewRelicService, error) {
	conf := ctx.ConfigService.Config().NewRelic
	appConf := ctx.ConfigService.Config().App
	newrelicConfig := newrelic.NewConfig(appConf.Name+" - "+appConf.Env, conf.APIKey)
	newrelicConfig.Enabled = conf.Enabled
	application, err := newrelic.NewApplication(newrelicConfig)
	if err != nil {
		return nil, err
	}
	return &newRelicService{application, ctx}, nil
}

func (nrc *newRelicService) Application() newrelic.Application {
	return nrc.application
}
