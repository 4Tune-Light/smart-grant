package config

import "time"

type Config struct {
	App      App      `mapstructure:"app"`
	Server   Server   `mapstructure:"server"`
	Gateway  Gateway  `mapstructure:"gateway"`
	Database Database `mapstructure:"database"`
	Redis    Redis    `mapstructure:"redis"`
	Storage  Storage  `mapstructure:"storage"`
	OTel     OTel     `mapstructure:"otel"`
	JWT      JWT      `mapstructure:"jwt"`
}

type Storage struct {
	Minio Minio `mapstructure:"minio"`
}

type Minio struct {
	Endpoint  string `mapstructure:"endpoint"`
	AccessKey string `mapstructure:"access_key"`
	SecretKey string `mapstructure:"secret_key"`
	Bucket    string `mapstructure:"bucket"`
	UseSSL    bool   `mapstructure:"use_ssl"`
	Region    string `mapstructure:"region"`
}

type App struct {
	Name  string `mapstructure:"name"`
	Env   string `mapstructure:"env"`
	Debug bool   `mapstructure:"debug"`
}

type Server struct {
	HTTP ServerHTTP `mapstructure:"http"`
	GRPC ServerGRPC `mapstructure:"grpc"`
}

type ServerHTTP struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
}

type ServerGRPC struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

type Gateway struct {
	HTTP      GatewayHTTP `mapstructure:"http"`
	RateLimit int         `mapstructure:"rate_limit"`
}

type GatewayHTTP struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

type Database struct {
	Postgres Postgres `mapstructure:"postgres"`
}

type Postgres struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	DBName          string        `mapstructure:"dbname"`
	SSLMode         string        `mapstructure:"sslmode"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `mapstructure:"conn_max_idle_time"`
}

type Redis struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	Password     string `mapstructure:"password"`
	DB           int    `mapstructure:"db"`
	PoolSize     int    `mapstructure:"pool_size"`
	MinIdleConns int    `mapstructure:"min_idle_conns"`
}

type OTel struct {
	ServiceName string  `mapstructure:"service_name"`
	Environment string  `mapstructure:"environment"`
	Endpoint    string  `mapstructure:"endpoint"`
	Insecure    bool    `mapstructure:"insecure"`
	TraceRatio  float64 `mapstructure:"trace_ratio"`
}

type JWT struct {
	Secret      string        `mapstructure:"secret"`
	AccessTTL   time.Duration `mapstructure:"access_ttl"`
	RefreshTTL  time.Duration `mapstructure:"refresh_ttl"`
}
