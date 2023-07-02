package app

import (
	"Ethereum_Service/pkg/utils/logger"
)

func initLoggerApplicationHook(app *Application) error {
	l, err := logger.New(logger.Options{
		Level:   app.GetConfig().LogLevel,
		Outputs: []string{app.GetConfig().LogFile},
	})

	if err != nil {
		panic(err)
	}

	logger.SetLogger(l)
	app.SetLogger(l.Sugar())

	return nil
}
