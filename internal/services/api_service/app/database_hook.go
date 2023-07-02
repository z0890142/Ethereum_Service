package app

import (
	"Ethereum_Service/config"
	"Ethereum_Service/pkg/utils/common"
	"Ethereum_Service/pkg/utils/logger"
	"database/sql"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	gormMysql "gorm.io/driver/mysql"

	"gorm.io/gorm"
)

func InitDatabaseHook(app *Application) error {
	db, err := common.OpenMysqlDatabase(&config.GetConfig().Databases)
	if err != nil {
		return fmt.Errorf("InitDatabaseHook: %s", err)
	}

	if err := migration(db); err != nil {
		return fmt.Errorf("InitDatabaseHook: %s", err)
	}

	if err := InitGormClientHook(app, db); err != nil {
		return fmt.Errorf("InitDatabaseHook: %s", err)
	}
	return nil
}

func InitGormClientHook(app *Application, db *sql.DB) error {
	gormClient, err := gorm.Open(gormMysql.New(gormMysql.Config{
		SkipInitializeWithVersion: true,
		Conn:                      db,
	}), &gorm.Config{SkipDefaultTransaction: true})
	if err != nil {
		return fmt.Errorf("InitGormClientHook: %v", err)
	}
	app.gormClient = gormClient
	return nil
}

func migration(db *sql.DB) error {
	driver, err := mysql.WithInstance(db, &mysql.Config{})
	if err != nil {
		return fmt.Errorf("Migrate: %v", err)
	}
	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", config.GetConfig().MigrationFilePath),
		config.GetConfig().Databases.DBName,
		driver)
	if err != nil {
		return fmt.Errorf("Migrate: %v", err)
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("Migrate: %v", err)
	}

	logger.Info("Migrate: Migrate successfully")
	return nil
}
