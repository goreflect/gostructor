// Command webservice is the big, realistic example: a full production-style
// microservice configuration assembled from THREE layers at once -
//
//	base defaults in a JSON file  <  operator env-var overrides  <  cf_default
//
// across ~30 fields grouped into nested sub-structs (service, server, TLS,
// database, redis, observability, auth). Secrets are marked cf_secret so
// they stay masked in the resolution trace, and ConfigureWithReport prints
// exactly where every value came from.
//
// This is what wiring gostructor into a real service looks like.
//
// Run it:
//
//	go run ./examples/webservice
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/goreflect/gostructor"
)

type Config struct {
	Service       ServiceConfig
	Server        ServerConfig
	TLS           TLSConfig
	Database      DatabaseConfig
	Redis         RedisConfig
	Observability ObservabilityConfig
	Auth          AuthConfig
}

type ServiceConfig struct {
	Name        string `cf_json:"service.name" cf_default:"payments-api"`
	Environment string `cf_env:"APP_ENV" cf_json:"service.environment" cf_default:"development"`
	Version     string `cf_env:"APP_VERSION" cf_default:"dev"`
}

type ServerConfig struct {
	Host           string        `cf_env:"SERVER_HOST" cf_json:"server.host" cf_default:"0.0.0.0"`
	Port           int           `cf_env:"SERVER_PORT" cf_json:"server.port" cf_default:"8080"`
	ReadTimeout    time.Duration `cf_json:"server.readTimeout" cf_default:"15s"`
	WriteTimeout   time.Duration `cf_json:"server.writeTimeout" cf_default:"15s"`
	MaxHeaderBytes int           `cf_json:"server.maxHeaderBytes" cf_default:"1048576"`
}

type TLSConfig struct {
	Enabled  bool   `cf_json:"tls.enabled" cf_default:"false"`
	CertFile string `cf_json:"tls.certFile" cf_default:""`
	KeyFile  string `cf_json:"tls.keyFile" cf_default:""`
}

type DatabaseConfig struct {
	Driver          string        `cf_json:"database.driver" cf_default:"postgres"`
	Host            string        `cf_env:"DB_HOST" cf_json:"database.host" cf_default:"localhost"`
	Port            int           `cf_env:"DB_PORT" cf_json:"database.port" cf_default:"5432"`
	Name            string        `cf_env:"DB_NAME" cf_json:"database.name" cf_default:"payments"`
	User            string        `cf_env:"DB_USER" cf_json:"database.user" cf_default:"app"`
	Password        string        `cf_env:"DB_PASSWORD" cf_secret:""`
	MaxOpenConns    int           `cf_json:"database.maxOpenConns" cf_default:"25"`
	MaxIdleConns    int           `cf_json:"database.maxIdleConns" cf_default:"5"`
	ConnMaxLifetime time.Duration `cf_json:"database.connMaxLifetime" cf_default:"30m"`
}

type RedisConfig struct {
	Addr     string `cf_env:"REDIS_ADDR" cf_json:"redis.addr" cf_default:"localhost:6379"`
	Password string `cf_env:"REDIS_PASSWORD" cf_secret:"" cf_default:""`
	DB       int    `cf_json:"redis.db" cf_default:"0"`
	PoolSize int    `cf_json:"redis.poolSize" cf_default:"10"`
}

type ObservabilityConfig struct {
	// LogLevel: in prod the operator's env wins; in dev the JSON base wins.
	LogLevel       string  `cf_env:"LOG_LEVEL" cf_json:"observability.logLevel" cf_default:"info" cf_priority:"prod:cf_env,cf_json,cf_default;dev:cf_json,cf_env,cf_default"`
	LogFormat      string  `cf_json:"observability.logFormat" cf_default:"json"`
	MetricsAddr    string  `cf_json:"observability.metricsAddr" cf_default:":9090"`
	TracingEnabled bool    `cf_json:"observability.tracingEnabled" cf_default:"true"`
	SampleRate     float64 `cf_env:"TRACE_SAMPLE_RATE" cf_json:"observability.sampleRate" cf_default:"0.1"`
}

type AuthConfig struct {
	JWTSecret string        `cf_env:"JWT_SECRET" cf_secret:""`
	TokenTTL  time.Duration `cf_json:"auth.tokenTTL" cf_default:"1h"`
	Issuer    string        `cf_json:"auth.issuer" cf_default:"payments-api"`
}

// baseConfig is the JSON base layer a service would ship in its image or pull
// from a config repo. Operators override selected keys via env at deploy time.
const baseConfig = `{
  "service":   { "name": "payments-api", "environment": "staging" },
  "server":    { "host": "0.0.0.0", "port": 8080, "readTimeout": "10s", "writeTimeout": "10s" },
  "tls":       { "enabled": true, "certFile": "/etc/tls/tls.crt", "keyFile": "/etc/tls/tls.key" },
  "database":  { "host": "db.internal", "port": 5432, "name": "payments", "maxOpenConns": 50 },
  "redis":     { "addr": "redis.internal:6379", "poolSize": 20 },
  "observability": { "logLevel": "info", "sampleRate": 0.25, "tracingEnabled": true },
  "auth":      { "issuer": "https://auth.internal", "tokenTTL": "30m" }
}`

func main() {
	// 1. The JSON base layer (would be a mounted file / config-repo pull).
	path := filepath.Join(os.TempDir(), "gostructor-webservice.json")
	if err := os.WriteFile(path, []byte(baseConfig), 0o644); err != nil {
		fmt.Println("write base config:", err)
		os.Exit(1)
	}
	defer os.Remove(path)

	// 2. Operator overrides + secrets injected via the environment at deploy.
	os.Setenv("APP_ENV", "production")
	os.Setenv("APP_VERSION", "1.8.3")
	os.Setenv("SERVER_PORT", "443")
	os.Setenv("DB_HOST", "prod-db.internal")
	os.Setenv("DB_PASSWORD", "s3cr3t-db-pw")
	os.Setenv("REDIS_PASSWORD", "s3cr3t-redis-pw")
	os.Setenv("JWT_SECRET", "super-secret-signing-key")
	os.Setenv("LOG_LEVEL", "warn") // prod priority => this wins over JSON's "info"
	os.Setenv(gostructor.PriorityEnvVar, "prod")

	// 3. Resolve, with a trace so we can see where each value came from.
	cfg, report, err := gostructor.ConfigureWithReport(&Config{},
		gostructor.WithSources(
			gostructor.Env(),
			gostructor.JSONFile(path),
			gostructor.Default(),
		),
		// Reveal only the last three chars of secrets in the trace.
		gostructor.WithMasker(func(_ gostructor.FieldContext, v any) string {
			s, _ := v.(string)
			if len(s) >= 3 {
				return "•••" + s[len(s)-3:]
			}
			return "•••"
		}),
	)
	if err != nil {
		fmt.Println("configure failed:", err)
		os.Exit(1)
	}

	fmt.Printf("%s v%s starting in %q\n", cfg.Service.Name, cfg.Service.Version, cfg.Service.Environment)
	fmt.Printf("  listen:   %s:%d  (TLS %v)\n", cfg.Server.Host, cfg.Server.Port, cfg.TLS.Enabled)
	fmt.Printf("  database: %s@%s:%d/%s (pool %d, password set: %d chars)\n",
		cfg.Database.User, cfg.Database.Host, cfg.Database.Port, cfg.Database.Name,
		cfg.Database.MaxOpenConns, len(cfg.Database.Password))
	fmt.Printf("  redis:    %s (pool %d)\n", cfg.Redis.Addr, cfg.Redis.PoolSize)
	fmt.Printf("  logging:  level=%s format=%s sample=%.2f\n",
		cfg.Observability.LogLevel, cfg.Observability.LogFormat, cfg.Observability.SampleRate)
	fmt.Printf("  auth:     issuer=%s ttl=%s\n", cfg.Auth.Issuer, cfg.Auth.TokenTTL)

	fmt.Println("\n── resolution trace (who won each field; secrets masked) ──")
	fmt.Println(report.String())
}
