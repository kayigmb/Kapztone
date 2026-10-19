// Package config is the only place that reads the environment. Every setting
// is required: there are no defaults, so a missing or invalid value stops the
// API at start with one message listing everything to fix.
package config

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/big"
	"net/url"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/kayigmb/chaintrack/backend/internal/auth"
	"github.com/kayigmb/chaintrack/backend/internal/chain"
	"github.com/kayigmb/chaintrack/backend/internal/notify"
	"github.com/kayigmb/chaintrack/backend/internal/secret"
	"github.com/kayigmb/chaintrack/backend/internal/upload"
)

// Names of the environment variables. `.env.example` documents each one.
const (
	EnvPort               = "PORT"
	EnvLogFormat          = "LOG_FORMAT" // json (default) or text
	EnvLogLevel           = "LOG_LEVEL"  // debug, info (default), warn, error
	EnvAppEnv             = "APP_ENV"
	EnvDatabaseURL        = "DATABASE_URL"
	EnvRedisURL           = "REDIS_URL"
	EnvJWTSecret          = "JWT_SECRET"
	EnvJWTTTL             = "JWT_TTL"          // sign-in lifetime, whole minutes
	EnvJWTTTLRemember     = "JWT_TTL_REMEMBER" // sign-in lifetime with "keep me signed in", whole minutes
	EnvEncryptionKey      = "ENCRYPTION_KEY"
	EnvSuperadminEmail    = "SUPERADMIN_EMAIL"
	EnvSuperadminPassword = "SUPERADMIN_PASSWORD"

	// The bucket that holds uploaded files: AWS S3, Google Cloud Storage or RustFS.
	EnvStorageEndpoint  = "STORAGE_ENDPOINT" // http(s) URL; the scheme decides TLS
	EnvStorageRegion    = "STORAGE_REGION"
	EnvStorageBucket    = "STORAGE_BUCKET"
	EnvStorageAccessKey = "STORAGE_ACCESS_KEY"
	EnvStorageSecretKey = "STORAGE_SECRET_KEY"

	// The chain the platform writes to, and the contracts deployed on it.
	EnvChainRPCURL         = "CHAIN_RPC_URL"
	EnvChainID             = "CHAIN_ID"
	EnvChainRelayerKey     = "CHAIN_RELAYER_KEY" // sends every transaction; relayer + operator on-chain, never admin
	EnvOrgRegistry         = "ORG_REGISTRY_ADDRESS"
	EnvFacilityRegistry    = "FACILITY_REGISTRY_ADDRESS"
	EnvPassportRegistry    = "PASSPORT_REGISTRY_ADDRESS"
	EnvCustodyManager      = "CUSTODY_MANAGER_ADDRESS"
	EnvAttestationRegistry = "ATTESTATION_REGISTRY_ADDRESS"

	// Outside services. Each is required like every setting; the value Off
	// says, in writing, that the service is not used.
	EnvPinataJWT = "PINATA_JWT" // pins certificate files to public IPFS

	// The mail server for notification e-mail. SMTP_HOST=none turns e-mail
	// off, and the other four are then not needed.
	EnvSMTPHost     = "SMTP_HOST"
	EnvSMTPPort     = "SMTP_PORT"
	EnvSMTPUser     = "SMTP_USER" // also the sender address
	EnvSMTPPassword = "SMTP_PASSWORD"
	EnvSMTPUseTLS   = "SMTP_USE_TLS" // true or false
)

// Off is the value that turns an outside service off.
const Off = "none"

// Values of APP_ENV.
const (
	AppDev  = "dev"
	AppProd = "prod"
)

var AppEnvs = []string{AppDev, AppProd}

const MinJWTSecretLen = 32

type Config struct {
	Port               string
	Dev                bool
	DatabaseURL        string
	RedisURL           string
	Storage            upload.Config
	JWTSecret          []byte
	TokenTTL           time.Duration
	RememberTTL        time.Duration
	Secrets            *secret.Box
	SuperadminEmail    string
	SuperadminPassword string

	ChainRPCURL string
	ChainID     *big.Int
	RelayerKey  *ecdsa.PrivateKey
	Contracts   chain.Addresses

	// PinataJWT is empty and SMTP nil when the service is off.
	PinataJWT string
	SMTP      *notify.SMTP
}

// Domain is the signing domain for handovers on the configured chain.
func (c Config) Domain() chain.Domain {
	return chain.Domain{ChainID: c.ChainID, Contract: c.Contracts.Custody}
}

// Logging is the log format and level, read before anything else so start-up
// problems are logged the same way. Both are optional.
func Logging() (format, level string) {
	format, level = strings.TrimSpace(os.Getenv(EnvLogFormat)), strings.TrimSpace(os.Getenv(EnvLogLevel))
	if format == "" {
		format = "json"
	}
	if level == "" {
		level = "info"
	}
	return format, level
}

// Load reads and checks every setting. The error lists all problems at once.
func Load() (Config, error) {
	var problems []string
	get := func(name string) string {
		v := strings.TrimSpace(os.Getenv(name))
		if v == "" {
			problems = append(problems, name+" is required")
		}
		return v
	}
	// invalid records a problem for a value that is present but wrong.
	invalid := func(value, name, rule string) {
		if value != "" {
			problems = append(problems, name+" "+rule)
		}
	}

	c := Config{
		Port:               get(EnvPort),
		DatabaseURL:        get(EnvDatabaseURL),
		RedisURL:           get(EnvRedisURL),
		SuperadminEmail:    get(EnvSuperadminEmail),
		SuperadminPassword: get(EnvSuperadminPassword),
		Storage: upload.Config{
			Endpoint:  get(EnvStorageEndpoint),
			Region:    get(EnvStorageRegion),
			Bucket:    get(EnvStorageBucket),
			AccessKey: get(EnvStorageAccessKey),
			SecretKey: get(EnvStorageSecretKey),
		},
	}
	if u, err := url.Parse(c.Storage.Endpoint); err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		invalid(c.Storage.Endpoint, EnvStorageEndpoint, "must be an http or https URL, e.g. http://localhost:9000")
	}

	appEnv := get(EnvAppEnv)
	if !slices.Contains(AppEnvs, appEnv) {
		invalid(appEnv, EnvAppEnv, "must be one of: "+strings.Join(AppEnvs, ", "))
	}
	c.Dev = appEnv == AppDev

	jwtSecret := get(EnvJWTSecret)
	// Any length in dev; a short secret would let anyone forge sign-ins in prod.
	if !c.Dev && len(jwtSecret) < MinJWTSecretLen {
		invalid(jwtSecret, EnvJWTSecret, fmt.Sprintf("must be at least %d characters in prod", MinJWTSecretLen))
	}
	c.JWTSecret = []byte(jwtSecret)

	ttl := get(EnvJWTTTL)
	minutes, err := strconv.Atoi(ttl)
	if err != nil || minutes <= 0 {
		invalid(ttl, EnvJWTTTL, "must be a whole number of minutes above 0, e.g. 90")
	}
	c.TokenTTL = time.Duration(minutes) * time.Minute

	remember := get(EnvJWTTTLRemember)
	rememberMinutes, err := strconv.Atoi(remember)
	if err != nil || rememberMinutes <= 0 {
		invalid(remember, EnvJWTTTLRemember, "must be a whole number of minutes above 0, e.g. 43200 (30 days)")
	}
	c.RememberTTL = time.Duration(rememberMinutes) * time.Minute

	key := get(EnvEncryptionKey)
	if c.Secrets, err = secret.New(key); err != nil {
		invalid(key, EnvEncryptionKey, fmt.Sprintf("must be %d hex characters", secret.KeyBytes*2))
	}

	// Any length in dev; bcrypt reads at most MaxPasswordLen bytes everywhere.
	if n := len(c.SuperadminPassword); n > auth.MaxPasswordLen || (!c.Dev && n < auth.MinPasswordLen) {
		invalid(c.SuperadminPassword, EnvSuperadminPassword,
			fmt.Sprintf("must be %d to %d characters (in dev: up to %d)", auth.MinPasswordLen, auth.MaxPasswordLen, auth.MaxPasswordLen))
	}

	c.ChainRPCURL = get(EnvChainRPCURL)

	chainID := get(EnvChainID)
	id, ok := new(big.Int).SetString(chainID, 10)
	if !ok || id.Sign() <= 0 {
		invalid(chainID, EnvChainID, "must be a whole number above 0, e.g. 31337 for a local chain")
		id = new(big.Int)
	}
	c.ChainID = id

	relayerKey := get(EnvChainRelayerKey)
	if c.RelayerKey, err = chain.ParseKey(relayerKey); err != nil {
		invalid(relayerKey, EnvChainRelayerKey, "must be a private key: 64 hex characters")
	}

	address := func(name string) common.Address {
		v := get(name)
		if !common.IsHexAddress(v) {
			invalid(v, name, "must be a contract address: 0x followed by 40 hex characters")
		}
		return common.HexToAddress(v)
	}
	c.Contracts = chain.Addresses{
		Orgs:         address(EnvOrgRegistry),
		Facilities:   address(EnvFacilityRegistry),
		Passports:    address(EnvPassportRegistry),
		Custody:      address(EnvCustodyManager),
		Attestations: address(EnvAttestationRegistry),
	}

	// service reads a setting that may be Off.
	service := func(name string) string {
		if v := get(name); v != Off {
			return v
		}
		return ""
	}
	c.PinataJWT = service(EnvPinataJWT)
	if host := get(EnvSMTPHost); host != "" && host != Off {
		s := notify.SMTP{Host: host, Port: get(EnvSMTPPort), User: get(EnvSMTPUser), Password: os.Getenv(EnvSMTPPassword)}
		if p, err := strconv.Atoi(s.Port); err != nil || p < 1 || p > 65535 {
			invalid(s.Port, EnvSMTPPort, "must be a port number, e.g. 587")
		}
		// Optional in dev (empty is false); prod mail is always encrypted.
		useTLS := strings.TrimSpace(os.Getenv(EnvSMTPUseTLS))
		switch {
		case !c.Dev && useTLS != "true":
			problems = append(problems, EnvSMTPUseTLS+" must be true in prod (mail is always encrypted)")
		case useTLS == "" || useTLS == "true" || useTLS == "false":
			s.TLS = useTLS == "true"
		default:
			invalid(useTLS, EnvSMTPUseTLS, "must be true or false")
		}
		c.SMTP = &s
	}

	if len(problems) > 0 {
		return Config{}, errors.New("configuration is incomplete; set these in .env:\n  - " +
			strings.Join(problems, "\n  - "))
	}
	return c, nil
}
