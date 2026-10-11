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
		config.EnvUploadDir:           "uploads",
		config.EnvJWTSecret:           testSecret,
		config.EnvJWTTTL:              "90",
		config.EnvEncryptionKey:       testKey,
		config.EnvSuperadminEmail:     "root@chaintrack.test",
		config.EnvSuperadminPassword:  testPassword,
		config.EnvChainRPCURL:         "http://127.0.0.1:8545",
		config.EnvChainID:             "31337",
		config.EnvChainPlatformKey:    testChainKey,
		config.EnvOrgRegistry:         testAddress,
		config.EnvFacilityRegistry:    testAddress,
		config.EnvPassportRegistry:    testAddress,
		config.EnvCustodyManager:      testAddress,
		config.EnvAttestationRegistry: testAddress,
		config.EnvPinataJWT:           config.Off,
		config.EnvSMTPURL:             config.Off,
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
	if d := c.Domain(); d.ChainID.Int64() != 31337 || d.Contract != c.Contracts.Custody || c.PlatformKey == nil {
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
		{"short signing secret", config.EnvJWTSecret, strings.Repeat("a", config.MinJWTSecretLen-1)},
		{"short encryption key", config.EnvEncryptionKey, "abcd"},
		{"encryption key not hex", config.EnvEncryptionKey, strings.Repeat("z", len(testKey))},
		{"unknown environment", config.EnvAppEnv, "staging"},
		{"short superadmin password", config.EnvSuperadminPassword, strings.Repeat("a", auth.MinPasswordLen-1)},
		{"chain id in words", config.EnvChainID, "amoy"},
		{"zero chain id", config.EnvChainID, "0"},
		{"platform key too short", config.EnvChainPlatformKey, "abcd1234"},
		{"contract address too short", config.EnvCustodyManager, "0x1234"},
		{"mail server without a sender", config.EnvSMTPURL, "smtp://mail.test:587"},
		{"mail server that is not smtp", config.EnvSMTPURL, "https://mail.test:587?from=a@b.test"},
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
