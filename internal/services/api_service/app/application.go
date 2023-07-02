package app

import (
	"Ethereum_Service/c"
	"Ethereum_Service/config"
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

var defaultApplication *Application

type Application struct {
	mu         sync.Mutex
	logger     *zap.SugaredLogger
	srv        *http.Server
	config     *config.Config
	addr       string
	db         *sql.DB
	gormClient *gorm.DB
	// Init and destroy hooks
	initHooks    []ApplicationHook
	destroyHooks []ApplicationHook
}

type ApplicationHook func(*Application) error

// Default get default application
func Default() *Application {
	if defaultApplication != nil {
		return defaultApplication
	}
	o := &Application{
		config:       config.GetConfig(),
		initHooks:    make([]ApplicationHook, 0),
		destroyHooks: make([]ApplicationHook, 0),
		srv:          &http.Server{},
		addr:         config.GetConfig().Service.Host + ":" + config.GetConfig().Service.Port,
	}

	initSlice := []ApplicationHook{initLoggerApplicationHook}
	for _, hook := range initSlice {
		err := hook(o)
		if err != nil {
			panic(err)
		}
	}
	o.AddInitHook(InitDatabaseHook)
	o.AddInitHook(InitGinApplicationHook)

	defaultApplication = o

	return o
}

// Run application
func (app *Application) Run() {
	app.callInitHooks()

	errc := make(chan error)

	go func() {
		if app.logger != nil {
			app.logger.Info("running server on: ", app.addr)
		}
		app.srv.Addr = app.addr
		errc <- app.srv.ListenAndServe()
	}()

	app.logger.Error((fmt.Sprintf("application run error: %s", <-errc)))
}

// AddInitHook add init callback function
func (app *Application) AddInitHook(f ApplicationHook) {
	app.initHooks = append(app.initHooks, f)
}

func (app *Application) AddDestroyHook(f ApplicationHook) {
	app.destroyHooks = append(app.destroyHooks, f)
}

// Shutdown shundown service
func (app *Application) Shutdown() {
	app.mu.Lock()
	defer app.mu.Unlock()

	if app.logger != nil {
		app.logger.Warn("shutdowning")
	}

	c, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := app.srv.Shutdown(c); err != nil {
		app.logger.Error("srv.Shutdown:", err)
	}
	select {
	case <-c.Done():
		app.logger.Info("Graceful Shutdown http server")
		done := make(chan struct{})
		go app.callDestroyHooks(done)
		t := time.NewTimer(5 * time.Second)

		select {
		case <-done:
			break
		case <-t.C:
			if app.logger != nil {
				app.logger.Warn("timeout: application destroy hooks interrupted")
			}
			break
		}
	}

}

func (app *Application) callInitHooks() {
	for _, hook := range app.initHooks {
		if err := hook(app); err != nil {
			panic(err)
		}
	}
}

func (app *Application) callDestroyHooks(done chan struct{}) {
	for i := len(app.destroyHooks); i > 0; i-- {
		hook := app.destroyHooks[i-1]
		if err := hook(app); err != nil {
			app.logger.Error("calling application destroy hook error: ", err.Error())
		}
	}

	done <- struct{}{}
}

func (app *Application) IsProduction() bool {
	return app.Environment() == c.EnvProduction
}

// Environment get Environment
func (app *Application) Environment() string {
	return strings.ToLower(app.config.Env)
}

func (app *Application) SetSrv(handler http.Handler) {
	app.srv.Handler = handler
}

func (app *Application) SetAddr(addr string) {
	app.addr = addr
}
func (app *Application) SetLogger(logger *zap.SugaredLogger) {
	app.logger = logger
}
func (app *Application) GetLogger() *zap.SugaredLogger {
	return app.logger
}

func (app *Application) GetConfig() *config.Config {
	return app.config
}

func (app *Application) SetDatabase(db *sql.DB) {
	app.db = db
}

func (app *Application) GetDatabase() *sql.DB {
	return app.db
}

func (app *Application) SetGormClient(db *gorm.DB) {
	app.gormClient = db
}

func (app *Application) GetGormClient() *gorm.DB {
	return app.gormClient
}
