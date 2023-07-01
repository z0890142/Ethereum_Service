package common

import (
	"Ethereum_Service/c"
	"Ethereum_Service/config"
	"database/sql"
	"fmt"
	"time"

	"Ethereum_Service/pkg/utils/logger"

	"github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	migration_mysql "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	mysql_driver "github.com/go-sql-driver/mysql"
)

func SqlErrCode(err error) int {
	mysqlErr, ok := err.(*mysql_driver.MySQLError)
	if !ok {
		return 0
	}
	return int(mysqlErr.Number)
}

func OpenMysqlDatabase(option *config.DatabaseOption) (db *sql.DB, err error) {

	connection, err := GetConnectionString(option)
	if err != nil {
		return nil, fmt.Errorf("openMysqlDatabase: %v", err)
	}

	if db, err = sql.Open(c.DriverMysql, connection); err != nil {
		return nil, fmt.Errorf("openMysqlDatabase: %v", err)
	} else {
		err = db.Ping()
		if err != nil {
			return nil, fmt.Errorf("openMysqlDatabase: %v", err)
		}
	}

	// Set connection pool
	if option.PoolSize > 0 {
		db.SetMaxIdleConns(option.PoolSize)
		db.SetMaxOpenConns(option.PoolSize)
	}

	return
}

func GetConnectionString(option *config.DatabaseOption) (string, error) {

	var loc = time.Local
	var err error
	if len(option.Timezone) > 0 {
		if loc, err = time.LoadLocation(option.Timezone); err != nil {
			return "", fmt.Errorf("GetConnectionString: %v", err)
		}
	}
	c := mysql.Config{
		User:                 option.Username,
		Passwd:               option.Password,
		Net:                  "tcp",
		Addr:                 fmt.Sprintf("%s:%d", option.Host, option.Port),
		DBName:               option.DBName,
		Loc:                  loc,
		Timeout:              option.Timeout,
		ReadTimeout:          option.ReadTimeout,
		WriteTimeout:         option.WriteTimeout,
		ParseTime:            true,
		CheckConnLiveness:    true,
		AllowNativePasswords: true,
		MaxAllowedPacket:     4 << 20, // 4MB
		Collation:            "utf8mb4_general_ci",
		MultiStatements:      true,
	}
	if len(option.Charset) > 0 {
		c.Params = make(map[string]string)
		c.Params["charset"] = option.Charset
	}
	return c.FormatDSN(), nil

}

func Migration(instance *sql.DB) error {
	driver, err := migration_mysql.WithInstance(instance, &migration_mysql.Config{})
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
