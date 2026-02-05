package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config 应用配置结构
type Config struct {
	App       AppConfig       `mapstructure:"app"`
	Database  DatabaseConfig  `mapstructure:"database"`
	JWT       JWTConfig       `mapstructure:"jwt"`
	Agent     AgentConfig     `mapstructure:"agent"`
	Log       LogConfig       `mapstructure:"log"`
	CORS      CORSConfig      `mapstructure:"cors"`
	RateLimit RateLimitConfig `mapstructure:"rate_limit"`
	Upload    UploadConfig    `mapstructure:"upload"`
}

// AppConfig 应用基础配置
type AppConfig struct {
	Name  string `mapstructure:"name"`
	Env   string `mapstructure:"env"`
	Port  int    `mapstructure:"port"`
	Debug bool   `mapstructure:"debug"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Postgres PostgresConfig `mapstructure:"postgres"`
	Neo4j    Neo4jConfig    `mapstructure:"neo4j"`
	Redis    RedisConfig    `mapstructure:"redis"`
}

// PostgresConfig PostgreSQL配置
type PostgresConfig struct {
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	Name            string `mapstructure:"name"`
	User            string `mapstructure:"user"`
	Password        string `mapstructure:"password"`
	SSLMode         string `mapstructure:"sslmode"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"`
}

// DSN 返回PostgreSQL连接字符串
func (c *PostgresConfig) DSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Name, c.SSLMode)
}

// Neo4jConfig Neo4j配置
type Neo4jConfig struct {
	URI            string `mapstructure:"uri"`
	User           string `mapstructure:"user"`
	Password       string `mapstructure:"password"`
	Database       string `mapstructure:"database"`
	MaxConnections int    `mapstructure:"max_connections"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

// Addr 返回Redis地址
func (c *RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret        string `mapstructure:"secret"`
	Expiry        string `mapstructure:"expiry"`
	RefreshExpiry string `mapstructure:"refresh_expiry"`
	Issuer        string `mapstructure:"issuer"`
}

// ExpiryDuration 返回Token过期时间
func (c *JWTConfig) ExpiryDuration() time.Duration {
	d, err := time.ParseDuration(c.Expiry)
	if err != nil {
		return 24 * time.Hour
	}
	return d
}

// RefreshExpiryDuration 返回刷新Token过期时间
func (c *JWTConfig) RefreshExpiryDuration() time.Duration {
	d, err := time.ParseDuration(c.RefreshExpiry)
	if err != nil {
		return 7 * 24 * time.Hour
	}
	return d
}

// AgentConfig 智能体服务配置
type AgentConfig struct {
	URL     string `mapstructure:"url"`
	Timeout int    `mapstructure:"timeout"`
	APIKey  string `mapstructure:"api_key"`
}

// TimeoutDuration 返回超时时间
func (c *AgentConfig) TimeoutDuration() time.Duration {
	return time.Duration(c.Timeout) * time.Second
}

// LogConfig 日志配置
type LogConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"`
	Output     string `mapstructure:"output"`
	FilePath   string `mapstructure:"file_path"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`
}

// CORSConfig CORS配置
type CORSConfig struct {
	AllowedOrigins   []string `mapstructure:"allowed_origins"`
	AllowedMethods   []string `mapstructure:"allowed_methods"`
	AllowedHeaders   []string `mapstructure:"allowed_headers"`
	ExposedHeaders   []string `mapstructure:"exposed_headers"`
	AllowCredentials bool     `mapstructure:"allow_credentials"`
	MaxAge           int      `mapstructure:"max_age"`
}

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	Enabled           bool `mapstructure:"enabled"`
	RequestsPerSecond int  `mapstructure:"requests_per_second"`
	Burst             int  `mapstructure:"burst"`
}

// UploadConfig 上传配置
type UploadConfig struct {
	MaxSize      int64    `mapstructure:"max_size"`
	AllowedTypes []string `mapstructure:"allowed_types"`
	StoragePath  string   `mapstructure:"storage_path"`
}

var cfg *Config

// Load 加载配置
func Load(configPath string) (*Config, error) {
	v := viper.New()

	// 设置配置文件路径
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")

	// 读取配置文件
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// 环境变量替换
	for _, key := range v.AllKeys() {
		val := v.GetString(key)
		if strings.HasPrefix(val, "${") && strings.Contains(val, "}") {
			envVal := resolveEnvVar(val)
			v.Set(key, envVal)
		}
	}

	// 绑定环境变量
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 解析配置
	cfg = &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// 从环境变量覆盖配置
	overrideFromEnv(cfg)

	return cfg, nil
}

// resolveEnvVar 解析环境变量格式 ${VAR:default}
func resolveEnvVar(val string) string {
	// 移除 ${ 和 }
	inner := strings.TrimPrefix(val, "${")
	inner = strings.TrimSuffix(inner, "}")

	// 分割环境变量名和默认值
	parts := strings.SplitN(inner, ":", 2)
	envName := parts[0]
	defaultVal := ""
	if len(parts) > 1 {
		defaultVal = parts[1]
	}

	// 获取环境变量值
	if envVal := os.Getenv(envName); envVal != "" {
		return envVal
	}
	return defaultVal
}

// overrideFromEnv 从环境变量覆盖配置
func overrideFromEnv(cfg *Config) {
	// 数据库配置
	if host := os.Getenv("DB_HOST"); host != "" {
		cfg.Database.Postgres.Host = host
	}
	if user := os.Getenv("DB_USER"); user != "" {
		cfg.Database.Postgres.User = user
	}
	if password := os.Getenv("DB_PASSWORD"); password != "" {
		cfg.Database.Postgres.Password = password
	}
	if name := os.Getenv("DB_NAME"); name != "" {
		cfg.Database.Postgres.Name = name
	}

	// Neo4j配置
	if uri := os.Getenv("NEO4J_URI"); uri != "" {
		cfg.Database.Neo4j.URI = uri
	}
	if user := os.Getenv("NEO4J_USER"); user != "" {
		cfg.Database.Neo4j.User = user
	}
	if password := os.Getenv("NEO4J_PASSWORD"); password != "" {
		cfg.Database.Neo4j.Password = password
	}

	// Redis配置
	if host := os.Getenv("REDIS_HOST"); host != "" {
		cfg.Database.Redis.Host = host
	}
	if password := os.Getenv("REDIS_PASSWORD"); password != "" {
		cfg.Database.Redis.Password = password
	}

	// JWT配置
	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		cfg.JWT.Secret = secret
	}

	// Agent配置
	if url := os.Getenv("AGENT_SERVICE_URL"); url != "" {
		cfg.Agent.URL = url
	}
}

// Get 获取配置实例
func Get() *Config {
	return cfg
}

// IsDevelopment 是否开发环境
func (c *Config) IsDevelopment() bool {
	return c.App.Env == "development"
}

// IsProduction 是否生产环境
func (c *Config) IsProduction() bool {
	return c.App.Env == "production"
}
