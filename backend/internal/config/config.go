// Package config is the only place that reads the environment. Every setting
// is required: there are no defaults, so a missing or invalid value stops the
// API at start with one message listing everything to fix.
package config

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/big"
	"net/mail"
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
	EnvAppURL             = "APP_URL"    // the web app's bare address; e-mailed links start with it
	EnvRateLimit          = "RATE_LIMIT" // true or false; empty: on outside dev
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
	EnvSMTPUser     = "SMTP_USER" // the login; the sender is MAIL_FROM
	EnvSMTPPassword = "SMTP_PASSWORD"
	EnvSMTPUseTLS   = "SMTP_USE_TLS" // true or false
	// MAIL_FROM is the sender every recipient sees, e.g. ChainTrack <noreply@chaintrack.digital>.
	EnvMailFrom = "MAIL_FROM"
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

// HardhatKeys are Hardhat's public test accounts #0 and #1 (as in
// .env.example): fine on a local chain, refused in prod.
var HardhatKeys = []string{
	"ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80",
	"59c6995e998f97a5a0044966f0945389dc9e86dae88c7a8412f4603b6b78690d",
}

type Config struct {
	Port               string
	Dev                bool
	AppURL             string // no trailing slash
	DatabaseURL        string
	RedisURL           string
	Storage            upload.Config
	JWTSecret          []byte
	TokenTTL           time.Duration
	RememberTTL        time.Duration
	Secrets            *secret.Box
	SuperadminEmail    string
	SuperadminPassword string
	RateLimit          bool

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
	r := &reader{}
	c := Config{
		Port:               r.get(EnvPort),
		DatabaseURL:        r.get(EnvDatabaseURL),
		RedisURL:           r.get(EnvRedisURL),
		SuperadminEmail:    r.get(EnvSuperadminEmail),
		SuperadminPassword: r.get(EnvSuperadminPassword),
	}
	r.storage(&c)
	r.site(&c)
	switch v := strings.TrimSpace(os.Getenv(EnvRateLimit)); v {
	case "":
		c.RateLimit = !c.Dev
	case "true", "false":
		c.RateLimit = v == "true"
	default:
		r.invalid(v, EnvRateLimit, "must be true or false")
	}
	r.sessions(&c)
	r.chain(&c)
	c.PinataJWT = r.service(EnvPinataJWT)
	r.mail(&c)
	if len(r.problems) > 0 {
		return Config{}, errors.New("configuration is incomplete; set these in .env:\n  - " +
			strings.Join(r.problems, "\n  - "))
	}
	return c, nil
}

// reader collects every problem, so one start-up lists them all.
type reader struct{ problems []string }

func (r *reader) get(name string) string {
	v := strings.TrimSpace(os.Getenv(name))
	if v == "" {
		r.problems = append(r.problems, name+" is required")
	}
	return v
}

// invalid records a problem for a value that is present but wrong.
func (r *reader) invalid(value, name, rule string) {
	if value != "" {
		r.problems = append(r.problems, name+" "+rule)
	}
}

// service reads a setting that may be Off.
func (r *reader) service(name string) string {
	if v := r.get(name); v != Off {
		return v
	}
	return ""
}

func isWebURL(v string) bool {
	u, err := url.Parse(v)
	return err == nil && (u.Scheme == "http" || u.Scheme == "https") && u.Host != ""
}

func (r *reader) storage(c *Config) {
	c.Storage = upload.Config{
		Endpoint:  r.get(EnvStorageEndpoint),
		Region:    r.get(EnvStorageRegion),
		Bucket:    r.get(EnvStorageBucket),
		AccessKey: r.get(EnvStorageAccessKey),
		SecretKey: r.get(EnvStorageSecretKey),
	}
	if !isWebURL(c.Storage.Endpoint) {
		r.invalid(c.Storage.Endpoint, EnvStorageEndpoint, "must be an http or https URL, e.g. http://localhost:9000")
	}
}

// site reads the environment and APP_URL. E-mailed links start with APP_URL,
// never with a request header: a header is the caller's claim and would let
// anyone aim a reset link at their site.
func (r *reader) site(c *Config) {
	appEnv := r.get(EnvAppEnv)
	if !slices.Contains(AppEnvs, appEnv) {
		r.invalid(appEnv, EnvAppEnv, "must be one of: "+strings.Join(AppEnvs, ", "))
	}
	c.Dev = appEnv == AppDev

	c.AppURL = strings.TrimRight(r.get(EnvAppURL), "/")
	if u, err := url.Parse(c.AppURL); !isWebURL(c.AppURL) || err != nil || u.Path != "" || u.RawQuery != "" || u.Fragment != "" {
		r.invalid(c.AppURL, EnvAppURL, "must be an http or https address with no path or query, e.g. https://chaintrack.app")
	}
}

// sessions reads the sign-in secrets. Any length in dev; in prod a short JWT
// secret would let anyone forge sign-ins. bcrypt reads at most MaxPasswordLen
// bytes everywhere.
func (r *reader) sessions(c *Config) {
	jwtSecret := r.get(EnvJWTSecret)
	if !c.Dev && len(jwtSecret) < MinJWTSecretLen {
		r.invalid(jwtSecret, EnvJWTSecret, fmt.Sprintf("must be at least %d characters in prod", MinJWTSecretLen))
	}
	c.JWTSecret = []byte(jwtSecret)
	c.TokenTTL = r.minutes(EnvJWTTTL, "90")
	c.RememberTTL = r.minutes(EnvJWTTTLRemember, "43200 (30 days)")

	key := r.get(EnvEncryptionKey)
	var err error
	if c.Secrets, err = secret.New(key); err != nil {
		r.invalid(key, EnvEncryptionKey, fmt.Sprintf("must be %d hex characters", secret.KeyBytes*2))
	}
	if n := len(c.SuperadminPassword); n > auth.MaxPasswordLen || (!c.Dev && n < auth.MinPasswordLen) {
		r.invalid(c.SuperadminPassword, EnvSuperadminPassword,
			fmt.Sprintf("must be %d to %d characters (in dev: up to %d)", auth.MinPasswordLen, auth.MaxPasswordLen, auth.MaxPasswordLen))
	}
}

func (r *reader) minutes(name, example string) time.Duration {
	v := r.get(name)
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		r.invalid(v, name, "must be a whole number of minutes above 0, e.g. "+example)
	}
	return time.Duration(n) * time.Minute
}

// chain reads the network, the relayer key and the contract addresses.
// Prod refuses Hardhat's public test keys.
func (r *reader) chain(c *Config) {
	c.ChainRPCURL = r.get(EnvChainRPCURL)
	chainID := r.get(EnvChainID)
	id, ok := new(big.Int).SetString(chainID, 10)
	if !ok || id.Sign() <= 0 {
		r.invalid(chainID, EnvChainID, "must be a whole number above 0, e.g. 31337 for a local chain")
		id = new(big.Int)
	}
	c.ChainID = id

	relayerKey := r.get(EnvChainRelayerKey)
	var err error
	if c.RelayerKey, err = chain.ParseKey(relayerKey); err != nil {
		r.invalid(relayerKey, EnvChainRelayerKey, "must be a private key: 64 hex characters")
	}
	if !c.Dev && slices.Contains(HardhatKeys, strings.ToLower(strings.TrimPrefix(relayerKey, "0x"))) {
		r.invalid(relayerKey, EnvChainRelayerKey, "is one of Hardhat's public test keys: anyone can sign with it")
	}
	c.Contracts = chain.Addresses{
		Orgs:         r.address(EnvOrgRegistry),
		Facilities:   r.address(EnvFacilityRegistry),
		Passports:    r.address(EnvPassportRegistry),
		Custody:      r.address(EnvCustodyManager),
		Attestations: r.address(EnvAttestationRegistry),
	}
}

func (r *reader) address(name string) common.Address {
	v := r.get(name)
	if !common.IsHexAddress(v) {
		r.invalid(v, name, "must be a contract address: 0x followed by 40 hex characters")
	}
	return common.HexToAddress(v)
}

func (r *reader) mail(c *Config) {
	if host := r.get(EnvSMTPHost); host != "" && host != Off {
		c.SMTP = r.smtp(host, c.Dev)
	}
}

func (r *reader) smtp(host string, dev bool) *notify.SMTP {
	s := notify.SMTP{Host: host, Port: r.get(EnvSMTPPort), User: r.get(EnvSMTPUser), Password: os.Getenv(EnvSMTPPassword)}
	if from, err := mail.ParseAddress(r.get(EnvMailFrom)); err != nil || from.Address == "" {
		r.invalid(r.get(EnvMailFrom), EnvMailFrom, "must be a name and address, e.g. ChainTrack <noreply@chaintrack.digital>")
	} else {
		s.FromName, s.FromAddr = from.Name, from.Address
	}
	if p, err := strconv.Atoi(s.Port); err != nil || p < 1 || p > 65535 {
		r.invalid(s.Port, EnvSMTPPort, "must be a port number, e.g. 587")
	}
	// Optional in dev (empty is false); prod mail is always encrypted.
	useTLS := strings.TrimSpace(os.Getenv(EnvSMTPUseTLS))
	switch {
	case !dev && useTLS != "true":
		r.problems = append(r.problems, EnvSMTPUseTLS+" must be true in prod (mail is always encrypted)")
	case useTLS == "" || useTLS == "true" || useTLS == "false":
		s.TLS = useTLS == "true"
	default:
		r.invalid(useTLS, EnvSMTPUseTLS, "must be true or false")
	}
	return &s
}
