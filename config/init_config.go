package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/AuthMe01/authme-go-kit/logger"

	"github.com/spf13/viper"
)

// Op op
type Op struct {
	FileName      string
	VerifyElement []string
}

// Option option
type Option func(op *Op)

func LoadConf(paths string, globalConfig interface{}, opts ...Option) {
	var op Op
	for _, option := range opts {
		option(&op)
	}
	v := viper.GetViper()
	loadLocal(v, paths, globalConfig)
	loadEnv(v, globalConfig)
}

// loadLocal load local file
func loadLocal(vp *viper.Viper, dirPaths string, globalConfig interface{}) {

	defaultLoggingKeys := map[string]interface{}{
		"log_key": "local_config",
		"service": viper.GetString("SERVICE.NAME"),
		"env":     viper.GetString("ENV"),
	}
	dirPathSlice := strings.Split(dirPaths, "/")

	vp.AddConfigPath(strings.Join(dirPathSlice[:len(dirPathSlice)-1], "/"))
	vp.SetConfigName(dirPathSlice[len(dirPathSlice)-1])
	err := vp.ReadInConfig()
	if err != nil {
		logger.GetLoggerWithKeys(defaultLoggingKeys).Panic(fmt.Sprintf("ReadInConfig err: %s ", err))
	}

	err = vp.Unmarshal(globalConfig)
	if err != nil {
		logger.GetLoggerWithKeys(defaultLoggingKeys).Panic(fmt.Sprintf("fatal error config unmarshal:  err: %s ", err))
	}
}

func loadEnv(vp *viper.Viper, globalConfig interface{}) {
	defaultLoggingKeys := map[string]interface{}{
		"log_key": "local_config",
		"service": viper.GetString("SERVICE.NAME"),
		"env":     viper.GetString("ENV"),
	}
	vp.AutomaticEnv()

	for _, envstr := range os.Environ() {
		parts := strings.SplitN(envstr, "=", 2)
		key := parts[0]
		value := ""

		if len(parts) == 2 {
			value = parts[1]
		}
		vp.Set(key, value)
	}

	err := vp.Unmarshal(globalConfig)
	if err != nil {
		logger.GetLoggerWithKeys(defaultLoggingKeys).Panic(fmt.Sprintf("fatal error apollo unmarshall err: %s ", err))
	}
}
