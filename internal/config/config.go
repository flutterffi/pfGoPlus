package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	App           AppConfig           `mapstructure:"app"`
	HTTP          HTTPConfig          `mapstructure:"http"`
	GRPC          GRPCConfig          `mapstructure:"grpc"`
	Database      DatabaseConfig      `mapstructure:"database"`
	Logger        LoggerConfig        `mapstructure:"logger"`
	Auth          AuthConfig          `mapstructure:"auth"`
	Observability ObservabilityConfig `mapstructure:"observability"`
	TodoBackend   TodoBackendConfig   `mapstructure:"todo_backend"`
}

type AppConfig struct {
	Name string `mapstructure:"name"`
	Env  string `mapstructure:"env"`
}

type HTTPConfig struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
}

type GRPCConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	ClientTarget string `mapstructure:"client_target"`
}

type DatabaseConfig struct {
	Driver      string `mapstructure:"driver"`
	DSN         string `mapstructure:"dsn"`
	AutoMigrate bool   `mapstructure:"auto_migrate"`
}

type LoggerConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

type AuthConfig struct {
	JWTSecret      string        `mapstructure:"jwt_secret"`
	JWTIssuer      string        `mapstructure:"jwt_issuer"`
	AccessTokenTTL time.Duration `mapstructure:"access_token_ttl"`
	DemoUsername   string        `mapstructure:"demo_username"`
	DemoPassword   string        `mapstructure:"demo_password"`
}

type ObservabilityConfig struct {
	Enabled        bool   `mapstructure:"enabled"`
	Exporter       string `mapstructure:"exporter"`
	MetricsPath    string `mapstructure:"metrics_path"`
	OTLPEndpoint   string `mapstructure:"otlp_endpoint"`
	OTLPInsecure   bool   `mapstructure:"otlp_insecure"`
	ServiceVersion string `mapstructure:"service_version"`
}

type TodoBackendConfig struct {
	Mode string `mapstructure:"mode"`
}

func Load() (Config, error) {
	configName := strings.TrimSpace(os.Getenv("PFGO_CONFIG_NAME"))
	if configName == "" {
		configName = "config"
	}
	configFile := strings.TrimSpace(os.Getenv("PFGO_CONFIG_FILE"))
	profile := strings.TrimSpace(os.Getenv("PFGO_APP_ENV"))

	return loadWithOptions([]string{"./configs", "."}, configName, configFile, profile)
}

func loadWithOptions(paths []string, configName, configFile, profile string) (Config, error) {
	v := viper.New()
	v.SetConfigType("yaml")
	v.SetEnvPrefix("PFGO")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	setDefaults(v)

	if err := readBaseConfig(v, paths, configName, configFile); err != nil {
		return Config{}, err
	}

	activeProfile := strings.TrimSpace(profile)
	if activeProfile == "" {
		activeProfile = strings.TrimSpace(v.GetString("app.env"))
	}
	if activeProfile != "" {
		if err := mergeProfileConfig(v, paths, configName, activeProfile); err != nil {
			return Config{}, err
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return Config{}, fmt.Errorf("unmarshal config: %w", err)
	}

	return cfg, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("app.name", "pfGoPlus")
	v.SetDefault("app.env", "local")
	v.SetDefault("http.host", "0.0.0.0")
	v.SetDefault("http.port", 8080)
	v.SetDefault("http.read_timeout", "5s")
	v.SetDefault("http.write_timeout", "10s")
	v.SetDefault("http.idle_timeout", "30s")
	v.SetDefault("grpc.host", "0.0.0.0")
	v.SetDefault("grpc.port", 9090)
	v.SetDefault("grpc.client_target", "127.0.0.1:9090")
	v.SetDefault("database.driver", "sqlite")
	v.SetDefault("database.dsn", "./tmp/pfgo-plus.db")
	v.SetDefault("database.auto_migrate", true)
	v.SetDefault("logger.level", "info")
	v.SetDefault("logger.format", "json")
	v.SetDefault("auth.jwt_secret", "change-me-in-production")
	v.SetDefault("auth.jwt_issuer", "pfGoPlus")
	v.SetDefault("auth.access_token_ttl", "2h")
	v.SetDefault("auth.demo_username", "admin")
	v.SetDefault("auth.demo_password", "admin123")
	v.SetDefault("observability.enabled", true)
	v.SetDefault("observability.exporter", "stdout")
	v.SetDefault("observability.metrics_path", "/metrics")
	v.SetDefault("observability.otlp_endpoint", "127.0.0.1:4317")
	v.SetDefault("observability.otlp_insecure", true)
	v.SetDefault("observability.service_version", "v0.3.0")
	v.SetDefault("todo_backend.mode", "local")
}

func readBaseConfig(v *viper.Viper, paths []string, configName, configFile string) error {
	if configFile != "" {
		v.SetConfigFile(configFile)
		if err := v.ReadInConfig(); err != nil {
			return fmt.Errorf("read config file %s: %w", configFile, err)
		}
		return nil
	}

	v.SetConfigName(configName)
	for _, path := range paths {
		v.AddConfigPath(path)
	}

	if err := v.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if errors.As(err, &notFound) {
			return nil
		}
		return fmt.Errorf("read config: %w", err)
	}
	return nil
}

func mergeProfileConfig(v *viper.Viper, paths []string, configName, profile string) error {
	profileConfig := viper.New()
	profileConfig.SetConfigType("yaml")
	profileConfig.SetConfigName(fmt.Sprintf("%s.%s", configName, profile))
	for _, path := range paths {
		profileConfig.AddConfigPath(path)
	}

	if err := profileConfig.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if errors.As(err, &notFound) {
			return nil
		}
		return fmt.Errorf("read profile config %s.%s: %w", configName, profile, err)
	}

	if err := v.MergeConfigMap(profileConfig.AllSettings()); err != nil {
		return fmt.Errorf("merge profile config %s.%s: %w", configName, profile, err)
	}
	return nil
}
