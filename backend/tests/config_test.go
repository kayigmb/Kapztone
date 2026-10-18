package tests

import (
	"strings"
	"testing"
	"time"

	"github.com/kayigmb/chaintrack/backend/internal/auth"
	"github.com/kayigmb/chaintrack/backend/internal/config"
)

const (
	testChainKey = "0x1111111111111111111111111111111111111111111111111111111111111111" // public test key
	testAddress  = "0x00000000000000000000000000000000000C0FFE"
)

// fullEnv is a complete, valid configuration.
func fullEnv() map[string]string {
	return map[string]string{
		config.EnvPort:                "8080",
		config.EnvAppEnv:              config.AppDev,
		config.EnvDatabaseURL:         "postgres://u:p@localhost:5478/db",
		config.EnvRedisURL:            "redis://localhost:6379/0",
		config.EnvStorageEndpoint:     "http://localhost:9000",
		config.EnvStorageRegion:       "us-east-1",
		config.EnvStorageBucket:       "chaintrack",
		config.EnvStorageAccessKey:    "access",
		config.EnvStorageSecretKey:    "secret-key",
		config.EnvJWTSecret:           testSecret,
		config.EnvJWTTTL:              "90",
		config.EnvJWTTTLRemember:      "43200",
		config.EnvEncryptionKey:       testKey,
		config.EnvSuperadminEmail:     "root@chaintrack.test",
		config.EnvSuperadminPassword:  testPassword,
		config.EnvChainRPCURL:         "http://127.0.0.1:8545",
		config.EnvChainID:             "31337",
		config.EnvChainRelayerKey:     testChainKey,
		config.EnvOrgRegistry:         testAddress,
		config.EnvFacilityRegistry:    testAddress,
		config.EnvPassportRegistry:    testAddress,
		config.EnvCustodyManager:      testAddress,
		config.EnvAttestationRegistry: testAddress,
		config.EnvPinataJWT:           config.Off,
		config.EnvSMTPHost:            config.Off,
	}
}

func setEnv(t *testing.T, env map[string]string) {
	t.Helper()
	for k, v := range env {
		t.Setenv(k, v)
	}
}

func TestConfigLoad(t *testing.T) {
	setEnv(t, fullEnv())
	c, err := config.Load()
	if err != nil {
		t.Fatalf("complete configuration rejected: %v", err)
	}
	// JWT_TTL is plain minutes.
	if c.TokenTTL != 90*time.Minute || !c.Dev || c.Secrets == nil || string(c.JWTSecret) != testSecret {
		t.Fatalf("configuration read wrongly: %+v", c)
	}
	if d := c.Domain(); d.ChainID.Int64() != 31337 || d.Contract != c.Contracts.Custody || c.RelayerKey == nil {
		t.Fatalf("chain settings read wrongly: %+v", c.Contracts)
	}
}

func TestConfigReportsEveryProblemAtOnce(t *testing.T) {
	env := fullEnv()
	env[config.EnvDatabaseURL] = ""
	env[config.EnvSuperadminEmail] = "  "
	env[config.EnvJWTTTL] = "90m"
	setEnv(t, env)

	_, err := config.Load()
	if err == nil {
		t.Fatal("incomplete configuration accepted")
	}
	for _, name := range []string{config.EnvDatabaseURL, config.EnvSuperadminEmail, config.EnvJWTTTL} {
		if !strings.Contains(err.Error(), name) {
			t.Errorf("error does not mention %s:\n%v", name, err)
		}
	}
	if strings.Contains(err.Error(), config.EnvPort) {
		t.Errorf("error mentions a setting that is fine:\n%v", err)
	}
}

func TestConfigRejectsInvalidValues(t *testing.T) {
	cases := []struct{ name, key, value string }{
		{"lifetime with a unit", config.EnvJWTTTL, "90m"},
		{"lifetime with a fraction", config.EnvJWTTTL, "1.5"},
		{"zero lifetime", config.EnvJWTTTL, "0"},
		{"negative lifetime", config.EnvJWTTTL, "-5"},
		{"lifetime in words", config.EnvJWTTTL, "soon"},
		{"remembered lifetime with a unit", config.EnvJWTTTLRemember, "30d"},
		{"zero remembered lifetime", config.EnvJWTTTLRemember, "0"},
		{"short encryption key", config.EnvEncryptionKey, "abcd"},
		{"encryption key not hex", config.EnvEncryptionKey, strings.Repeat("z", len(testKey))},
		{"unknown environment", config.EnvAppEnv, "staging"},
		{"superadmin password past the bcrypt limit", config.EnvSuperadminPassword, strings.Repeat("a", auth.MaxPasswordLen+1)},
		{"chain id in words", config.EnvChainID, "amoy"},
		{"zero chain id", config.EnvChainID, "0"},
		{"relayer key too short", config.EnvChainRelayerKey, "abcd1234"},
		{"contract address too short", config.EnvCustodyManager, "0x1234"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			env := fullEnv()
			env[tc.key] = tc.value
			setEnv(t, env)
			_, err := config.Load()
			if err == nil || !strings.Contains(err.Error(), tc.key) {
				t.Fatalf("%s=%q: err = %v", tc.key, tc.value, err)
			}
			// The value itself must not be echoed: it may be a secret.
			if strings.Contains(err.Error(), tc.value) && len(tc.value) > 4 {
				t.Fatalf("error echoes the value: %v", err)
			}
		})
	}
}

// The superadmin password and the signing secret may be short in dev, never in prod.
func TestConfigShortSecretsByEnvironment(t *testing.T) {
	for key, short := range map[string]string{
		config.EnvSuperadminPassword: "abc",
		config.EnvJWTSecret:          strings.Repeat("a", config.MinJWTSecretLen-1),
	} {
		env := fullEnv()
		env[key] = short
		setEnv(t, env)
		if _, err := config.Load(); err != nil {
			t.Errorf("short %s refused in dev: %v", key, err)
		}
		t.Setenv(config.EnvAppEnv, config.AppProd)
		if _, err := config.Load(); err == nil || !strings.Contains(err.Error(), key) {
			t.Errorf("short %s accepted in prod: %v", key, err)
		}
	}
}

// E-mail on: host, port, user and TLS are read and checked; none turns it off.
func TestConfigSMTP(t *testing.T) {
	on := func() map[string]string {
		env := fullEnv()
		env[config.EnvSMTPHost] = "smtp.gmail.com"
		env[config.EnvSMTPPort] = "587"
		env[config.EnvSMTPUser] = "sender@chaintrack.test"
		env[config.EnvSMTPPassword] = "abcd efgh ijkl mnop"
		env[config.EnvSMTPUseTLS] = "true"
		return env
	}
	setEnv(t, on())
	c, err := config.Load()
	if err != nil || c.SMTP == nil || c.SMTP.Port != "587" || !c.SMTP.TLS || c.SMTP.Password != "abcd efgh ijkl mnop" {
		t.Fatalf("mail settings read wrongly: %+v (%v)", c.SMTP, err)
	}

	for key, value := range map[string]string{
		config.EnvSMTPPort:   "smtp",
		config.EnvSMTPUser:   "",
		config.EnvSMTPUseTLS: "maybe",
	} {
		env := on()
		env[key] = value
		setEnv(t, env)
		if _, err := config.Load(); err == nil || !strings.Contains(err.Error(), key) {
			t.Errorf("%s=%q accepted: %v", key, value, err)
		}
	}

	// SMTP_USE_TLS: optional in dev (empty is false), true required in prod.
	for _, tc := range []struct {
		appEnv, useTLS string
		ok, tls        bool
	}{
		{config.AppDev, "", true, false},
		{config.AppDev, "false", true, false},
		{config.AppProd, "true", true, true},
		{config.AppProd, "false", false, false},
		{config.AppProd, "", false, false},
	} {
		env := on()
		env[config.EnvAppEnv] = tc.appEnv
		env[config.EnvSMTPUseTLS] = tc.useTLS
		setEnv(t, env)
		c, err := config.Load()
		if tc.ok != (err == nil) || (err == nil && c.SMTP.TLS != tc.tls) {
			t.Errorf("%s with SMTP_USE_TLS=%q: %+v, %v", tc.appEnv, tc.useTLS, c.SMTP, err)
		}
	}

	env := fullEnv() // SMTP_HOST=none and nothing else set
	setEnv(t, env)
	for _, k := range []string{config.EnvSMTPPort, config.EnvSMTPUser, config.EnvSMTPPassword, config.EnvSMTPUseTLS} {
		t.Setenv(k, "")
	}
	if c, err := config.Load(); err != nil || c.SMTP != nil {
		t.Fatalf("e-mail off still needs settings: %v", err)
	}
}
