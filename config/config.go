package config

import (
	"sync"
	"time"
)

var globalConfig *Config
var configOnce sync.Once

// ResetConfig set config to Nil, used for tests
func ResetConfig() {
	globalConfig = nil
}

// GetConfig 獲取該服務相關配置
func GetConfig() *Config {
	configOnce.Do(func() {
		globalConfig = &Config{}
	})
	return globalConfig
}

// Config 該服務相關配置
type Config struct {
	Env              string  `mapstructure:"ENV"`
	ShortenerService Service `mapstructure:"SHORTENER_SERVICE"`

	LogLevel  string         `mapstructure:"LOG_LEVEL"`
	LogFile   string         `mapstructure:"LOG_FILE"`
	EndPoints EndPoints      `mapstructure:"ENDPOINTS"`
	Databases DatabaseOption `mapstructure:"DATABASES"`

	Redis RedisOption `mapstructure:"REDIS"`

	WorkerNumber      int    `mapstructure:"WORKER_NUMBER"`
	StoreBufferSize   int    `mapstructure:"STORE_BUFFER_SIZE"`
	MigrationFilePath string `mapstructure:"MIGRATION_FILE_PATH"`
	Trace             Trace  `mapstructure:"TRACE"`
	MaxRetryTime      int    `mapstructure:"MAX_RETRY_TIME"`
	RCPEndpoint       string `mapstructure:"RCP_ENDPOINT"`
}

// Service defines service configuration struct.
type Service struct {
	Name string `mapstructure:"NAME"`
	Host string `mapstructure:"HOST"`
	Port string `mapstructure:"PORT"`
}

type EndPoints struct {
	KeyServer Endpoint `mapstructure:"KEY_SERVER"`
}
type Endpoint struct {
	EnableGRCP    bool              `mapstructure:"ENABLE_GRCP"`
	Grpc          GrpcEndpoint      `mapstructure:"GRPC"`
	Http          HttpEndpoint      `mapstructure:"HTTP"`
	ExtensionInfo map[string]string `mapstructure:"EXTENSION_INFO"`
}

type HttpEndpoint struct {
	Host      string `mapstructure:"HOST"`
	Port      string `mapstructure:"PORT"`
	EnableTls bool   `mapstructure:"ENABLE_TLS"`
}

type GrpcEndpoint struct {
	Host     string `mapstructure:"HOST"`
	Port     int    `mapstructure:"PORT"`
	Insecure bool   `mapstructure:"INSECURE"`
}

type DatabaseOption struct {
	Driver       string        `mapstructure:"DRIVER"`
	Host         string        `mapstructure:"HOST"`
	Port         uint16        `mapstructure:"PORT"`
	Username     string        `mapstructure:"USERNAME"`
	Password     string        `mapstructure:"PASSWORD"`
	DBName       string        `mapstructure:"DBNAME"`
	Timezone     string        `mapstructure:"TIMEZONE"`
	Charset      string        `mapstructure:"CHARSET"`
	PoolSize     int           `mapstructure:"POOL_SIZE"`
	Timeout      time.Duration `mapstructure:"TIMEOUT"`
	ReadTimeout  time.Duration `mapstructure:"READ_TIMEOUT"`
	WriteTimeout time.Duration `mapstructure:"WRITE_TIMEOUT"`
}

type RedisOption struct {
	Enable   bool   `mapstructure:"ENABLE"`
	Host     string `mapstructure:"HOST"`
	Port     string `mapstructure:"PORT"`
	Password string `mapstructure:"PASSWORD"`
}

type Ratelimit struct {
	Enable bool  `mapstructure:"ENABLE"`
	Secend int   `mapstructure:"SECEND"`
	Number int64 `mapstructure:"NUMBER"`
}

type Trace struct {
	Enable   bool   `mapstructure:"ENABLE"`
	Endpoint string `mapstructure:"ENDPOINT"`
}
