// Command webservice is a production-style microservice configuration
// assembled from three layers:
//
//	operator env-var overrides  <  base config in a JSON file  <  gos default
//
// The layer order is the WithSources order (Env, JSON, Default): the first
// source that resolves a field wins. Fields are grouped into nested
// sub-structs (service, server, TLS, database, redis, observability, auth).
// Secrets are marked gos:"secret" so they stay masked in the resolution
// trace, and ConfigureWithReport prints a summary of where values came from.
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
	Name        string `cfg:"name,json:service.name" gos:"default:payments-api"`
	Environment string `cfg:"environment,env:APP_ENV,json:service.environment" gos:"default:development"`
	Version     string `cfg:"version,env:APP_VERSION" gos:"default:dev"`
}

type ServerConfig struct {
	Host           string        `cfg:"host,env:SERVER_HOST,json:server.host" gos:"default:0.0.0.0"`
	Port           int           `cfg:"port,env:SERVER_PORT,json:server.port" gos:"default:8080"`
	ReadTimeout    time.Duration `cfg:"readTimeout,json:server.readTimeout" gos:"default:15s"`
	WriteTimeout   time.Duration `cfg:"writeTimeout,json:server.writeTimeout" gos:"default:15s"`
	MaxHeaderBytes int           `cfg:"maxHeaderBytes,json:server.maxHeaderBytes" gos:"default:1048576"`
}

type TLSConfig struct {
	Enabled  bool   `cfg:"enabled,json:tls.enabled" gos:"default:false"`
	CertFile string `cfg:"certFile,json:tls.certFile" gos:"optional"`
	KeyFile  string `cfg:"keyFile,json:tls.keyFile" gos:"optional"`
}

type DatabaseConfig struct {
	Driver          string        `cfg:"driver,json:database.driver" gos:"default:postgres"`
	Host            string        `cfg:"host,env:DB_HOST,json:database.host" gos:"default:localhost"`
	Port            int           `cfg:"port,env:DB_PORT,json:database.port" gos:"default:5432"`
	Name            string        `cfg:"name,env:DB_NAME,json:database.name" gos:"default:payments"`
	User            string        `cfg:"user,env:DB_USER,json:database.user" gos:"default:app"`
	Password        string        `cfg:"password,env:DB_PASSWORD" gos:"secret"`
	MaxOpenConns    int           `cfg:"maxOpenConns,json:database.maxOpenConns" gos:"default:25"`
	MaxIdleConns    int           `cfg:"maxIdleConns,json:database.maxIdleConns" gos:"default:5"`
	ConnMaxLifetime time.Duration `cfg:"connMaxLifetime,json:database.connMaxLifetime" gos:"default:30m"`
}

type RedisConfig struct {
	Addr     string `cfg:"addr,env:REDIS_ADDR,json:redis.addr" gos:"default:localhost:6379"`
	Password string `cfg:"password,env:REDIS_PASSWORD" gos:"secret,optional"`
	DB       int    `cfg:"db,json:redis.db" gos:"default:0"`
	PoolSize int    `cfg:"poolSize,json:redis.poolSize" gos:"default:10"`
}

type ObservabilityConfig struct {
	// LogLevel: Env is listed before JSON in WithSources, so the operator's
	// LOG_LEVEL wins over the JSON base whenever it is set.
	LogLevel       string  `cfg:"logLevel,env:LOG_LEVEL,json:observability.logLevel" gos:"default:info"`
	LogFormat      string  `cfg:"logFormat,json:observability.logFormat" gos:"default:json"`
	MetricsAddr    string  `cfg:"metricsAddr,json:observability.metricsAddr" gos:"default::9090"`
	TracingEnabled bool    `cfg:"tracingEnabled,json:observability.tracingEnabled" gos:"default:true"`
	SampleRate     float64 `cfg:"sampleRate,env:TRACE_SAMPLE_RATE,json:observability.sampleRate" gos:"default:0.1"`
}

type AuthConfig struct {
	JWTSecret string        `cfg:"jwtSecret,env:JWT_SECRET" gos:"secret"`
	TokenTTL  time.Duration `cfg:"tokenTTL,json:auth.tokenTTL" gos:"default:1h"`
	Issuer    string        `cfg:"issuer,json:auth.issuer" gos:"default:payments-api"`
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
	os.Setenv("LOG_LEVEL", "warn") // Env before JSON => this wins over JSON's "info"

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
