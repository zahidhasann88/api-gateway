package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	LogLevel string
	Server   ServerConfig
	CORS     CORSConfig
	Proxy    ProxyConfig
	Auth     AuthConfig
	Services map[string]ServiceConfig
}

type ServerConfig struct {
	Address string
	Timeout int
}

type CORSConfig struct {
	AllowedOrigins []string
	AllowedMethods []string
	AllowedHeaders []string
}

type ProxyConfig struct {
	ReadTimeout  int
	WriteTimeout int
	IdleTimeout  int
}

type AuthConfig struct {
	Enabled    bool
	JWTSecret  string
	Expiration string
	Issuer     string
}

type ServiceConfig struct {
	URL             string
	Timeout         int
	RetryCount      int
	RateLimit       int
	Authentication  bool
	Authorization   AuthorizationConfig
	CircuitBreaker  CircuitBreakerConfig
	Transformations *TransformationConfig
}

type AuthorizationConfig struct {
	Roles []string
}

type CircuitBreakerConfig struct {
	Enabled                  bool
	FailureThreshold         int
	ResetTimeout             string
	HalfOpenSuccessThreshold int
}

type TransformationConfig struct {
	Request  *TransformConfig
	Response *TransformConfig
}

type TransformConfig struct {
	FieldMapping map[string]string
	HeaderToBody map[string]string
	BodyToHeader map[string]string
}

func Load(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(path)
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
