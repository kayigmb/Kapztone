// Package config is the only place that reads the environment. Every setting
// is required: there are no defaults, so a missing or invalid value stops the
// API at start with one message listing everything to fix.
package config

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/big"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/kayigmb/chaintrack/backend/internal/auth"
	"github.com/kayigmb/chaintrack/backend/internal/chain"
	"github.com/kayigmb/chaintrack/backend/internal/secret"
)

// Names of the environment variables. `.env.example` documents each one.
const (
	EnvPort               = "PORT"
	EnvAppEnv             = "APP_ENV"
	EnvDatabaseURL        = "DATABASE_URL"
	EnvRedisURL           = "REDIS_URL"
	EnvUploadDir          = "UPLOAD_DIR"
	EnvJWTSecret          = "JWT_SECRET"
	EnvJWTTTL             = "JWT_TTL" // sign-in lifetime, whole minutes
	EnvEncryptionKey      = "ENCRYPTION_KEY"
	EnvSuperadminEmail    = "SUPERADMIN_EMAIL"
	EnvSuperadminPassword = "SUPERADMIN_PASSWORD"

	// The chain the platform writes to, and the contracts deployed on it.
	EnvChainRPCURL         = "CHAIN_RPC_URL"
	EnvChainID             = "CHAIN_ID"
	EnvChainPlatformKey    = "CHAIN_PLATFORM_KEY" // sends every transaction; holds the admin and relayer roles
	EnvOrgRegistry         = "ORG_REGISTRY_ADDRESS"
	EnvFacilityRegistry    = "FACILITY_REGISTRY_ADDRESS"
	EnvPassportRegistry    = "PASSPORT_REGISTRY_ADDRESS"
	EnvCustodyManager      = "CUSTODY_MANAGER_ADDRESS"
	EnvAttestationRegistry = "ATTESTATION_REGISTRY_ADDRESS"
)

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
	UploadDir          string
	JWTSecret          []byte
	TokenTTL           time.Duration
	Secrets            *secret.Box
	SuperadminEmail    string
	SuperadminPassword string

	ChainRPCURL string
	ChainID     *big.Int
	PlatformKey *ecdsa.PrivateKey
	Contracts   chain.Addresses
}

// Domain is the signing domain for handovers on the configured chain.
func (c Config) Domain() chain.Domain {
	return chain.Domain{ChainID: c.ChainID, Contract: c.Contracts.Custody}
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
		UploadDir:          get(EnvUploadDir),
		SuperadminEmail:    get(EnvSuperadminEmail),
		SuperadminPassword: get(EnvSuperadminPassword),
	}

	appEnv := get(EnvAppEnv)
	if !slices.Contains(AppEnvs, appEnv) {
		invalid(appEnv, EnvAppEnv, "must be one of: "+strings.Join(AppEnvs, ", "))
	}
	c.Dev = appEnv == AppDev

	jwtSecret := get(EnvJWTSecret)
	if len(jwtSecret) < MinJWTSecretLen {
		invalid(jwtSecret, EnvJWTSecret, fmt.Sprintf("must be at least %d characters", MinJWTSecretLen))
	}
	c.JWTSecret = []byte(jwtSecret)

	ttl := get(EnvJWTTTL)
	minutes, err := strconv.Atoi(ttl)
	if err != nil || minutes <= 0 {
		invalid(ttl, EnvJWTTTL, "must be a whole number of minutes above 0, e.g. 90")
	}
	c.TokenTTL = time.Duration(minutes) * time.Minute

	key := get(EnvEncryptionKey)
	if c.Secrets, err = secret.New(key); err != nil {
		invalid(key, EnvEncryptionKey, fmt.Sprintf("must be %d hex characters", secret.KeyBytes*2))
	}

	if n := len(c.SuperadminPassword); n < auth.MinPasswordLen || n > auth.MaxPasswordLen {
		invalid(c.SuperadminPassword, EnvSuperadminPassword,
			fmt.Sprintf("must be %d to %d characters", auth.MinPasswordLen, auth.MaxPasswordLen))
	}

	c.ChainRPCURL = get(EnvChainRPCURL)

	chainID := get(EnvChainID)
	id, ok := new(big.Int).SetString(chainID, 10)
	if !ok || id.Sign() <= 0 {
		invalid(chainID, EnvChainID, "must be a whole number above 0, e.g. 31337 for a local chain")
		id = new(big.Int)
	}
	c.ChainID = id

	platformKey := get(EnvChainPlatformKey)
	if c.PlatformKey, err = chain.ParseKey(platformKey); err != nil {
		invalid(platformKey, EnvChainPlatformKey, "must be a private key: 64 hex characters")
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

	if len(problems) > 0 {
		return Config{}, errors.New("configuration is incomplete; set these in .env:\n  - " +
			strings.Join(problems, "\n  - "))
	}
	return c, nil
}
