//go:build !dev

package chainlink

import (
	_ "embed"
	"fmt"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	v2 "github.com/smartcontractkit/chainlink/v2/core/config/v2"
	"github.com/smartcontractkit/chainlink/v2/core/utils"
)

func TestTOMLGeneralConfig_Defaults(t *testing.T) {
	config, err := GeneralConfigOpts{}.New()
	require.NoError(t, err)
	assert.Equal(t, (*url.URL)(nil), config.WebServer().BridgeResponseURL())
	assert.Nil(t, config.DefaultChainID())
	assert.False(t, config.EVMRPCEnabled())
	assert.False(t, config.EVMEnabled())
	assert.False(t, config.CosmosEnabled())
	assert.False(t, config.SolanaEnabled())
	assert.False(t, config.StarkNetEnabled())
	assert.Equal(t, false, config.JobPipeline().ExternalInitiatorsEnabled())
	assert.Equal(t, 15*time.Minute, config.WebServer().SessionTimeout().Duration())
}

func TestTOMLGeneralConfig_InsecureConfig(t *testing.T) {
	t.Parallel()

	t.Run("all insecure configs are false by default", func(t *testing.T) {
		config, err := GeneralConfigOpts{}.New()
		require.NoError(t, err)

		assert.False(t, config.Insecure().DevWebServer())
		assert.False(t, config.Insecure().DisableRateLimiting())
		assert.False(t, config.Insecure().InfiniteDepthQueries())
		assert.False(t, config.Insecure().OCRDevelopmentMode())
	})

	t.Run("insecure config ignore override on non-dev builds", func(t *testing.T) {
		config, err := GeneralConfigOpts{
			OverrideFn: func(c *Config, s *Secrets) {
				*c.Insecure.DevWebServer = true
				*c.Insecure.DisableRateLimiting = true
				*c.Insecure.InfiniteDepthQueries = true
				*c.AuditLogger.Enabled = true
			}}.New()
		require.NoError(t, err)

		// Just asserting that override logic work on a safe config
		assert.True(t, config.AuditLogger().Enabled())

		assert.False(t, config.Insecure().DevWebServer())
		assert.False(t, config.Insecure().DisableRateLimiting())
		assert.False(t, config.Insecure().InfiniteDepthQueries())
	})

	t.Run("ValidateConfig fails if insecure config is set on non-dev builds", func(t *testing.T) {
		config := `
		  [insecure]
		  DevWebServer = true
		  DisableRateLimiting = false
		  InfiniteDepthQueries = false
		  OCRDevelopmentMode = false
		`
		opts := GeneralConfigOpts{
			ConfigStrings: []string{config},
		}
		cfg, err := opts.New()
		require.NoError(t, err)
		err = cfg.Validate()
		require.Contains(t, err.Error(), "invalid configuration: Insecure.DevWebServer: invalid value (true): insecure configs are not allowed on secure builds")
	})
}

func TestValidateDB(t *testing.T) {
	t.Setenv(string(v2.EnvConfig), "")

	t.Run("unset db url", func(t *testing.T) {
		t.Setenv(string(v2.EnvDatabaseURL), "")

		config, err := GeneralConfigOpts{}.New()
		require.NoError(t, err)

		err = config.ValidateDB()
		require.Error(t, err)
		require.ErrorIs(t, err, ErrInvalidSecrets)
	})

	t.Run("dev url", func(t *testing.T) {
		t.Setenv(string(v2.EnvDatabaseURL), "postgres://postgres:admin@localhost:5432/chainlink_dev_test?sslmode=disable")

		config, err := GeneralConfigOpts{}.New()
		require.NoError(t, err)
		err = config.ValidateDB()
		require.NoError(t, err)
	})

	t.Run("bad password url", func(t *testing.T) {
		t.Setenv(string(v2.EnvDatabaseURL), "postgres://postgres:pwdTooShort@localhost:5432/chainlink_dev_prod?sslmode=disable")
		t.Setenv(string(v2.EnvDatabaseAllowSimplePasswords), "false")

		config, err := GeneralConfigOpts{}.New()
		require.NoError(t, err)
		err = config.ValidateDB()
		require.Error(t, err)
		require.ErrorIs(t, err, ErrInvalidSecrets)
	})

}

func TestConfig_LogSQL(t *testing.T) {
	config, err := GeneralConfigOpts{}.New()
	require.NoError(t, err)

	config.SetLogSQL(true)
	assert.Equal(t, config.Database().LogSQL(), true)

	config.SetLogSQL(false)
	assert.Equal(t, config.Database().LogSQL(), false)
}

//go:embed testdata/mergingsecretsdata/secrets-database.toml
var databaseSecretsTOML string

//go:embed testdata/mergingsecretsdata/secrets-explorer.toml
var explorerSecretsTOML string

//go:embed testdata/mergingsecretsdata/secrets-password.toml
var passwordSecretsTOML string

//go:embed testdata/mergingsecretsdata/secrets-pyroscope.toml
var pyroscopeSecretsTOML string

//go:embed testdata/mergingsecretsdata/secrets-prometheus.toml
var prometheusSecretsTOML string

//go:embed testdata/mergingsecretsdata/secrets-mercury_a.toml
var mercurySecretsTOML_a string

//go:embed testdata/mergingsecretsdata/secrets-mercury_b.toml
var mercurySecretsTOML_b string

//go:embed testdata/mergingsecretsdata/secrets-threshold.toml
var thresholdSecretsTOML string

func TestConfig_SecretsMerging(t *testing.T) {
	setInFile := "set in config file"
	testConfigFileContents := Config{
		Core: v2.Core{
			RootDir: &setInFile,
			P2P: v2.P2P{
				V2: v2.P2PV2{
					AnnounceAddresses: &[]string{setInFile},
					ListenAddresses:   &[]string{setInFile},
				},
			},
		},
	}

	t.Run("verify secrets merging in GeneralConfigOpts.New()", func(t *testing.T) {
		databaseSecrets, err := parseSecrets(databaseSecretsTOML)
		require.NoErrorf(t, err, "error: %s", err)
		explorerSecrets, err1 := parseSecrets(explorerSecretsTOML)
		require.NoErrorf(t, err1, "error: %s", err1)
		passwordSecrets, err2 := parseSecrets(passwordSecretsTOML)
		require.NoErrorf(t, err2, "error: %s", err2)
		pyroscopeSecrets, err3 := parseSecrets(pyroscopeSecretsTOML)
		require.NoErrorf(t, err3, "error: %s", err3)
		prometheusSecrets, err4 := parseSecrets(prometheusSecretsTOML)
		require.NoErrorf(t, err4, "error: %s", err4)
		mercurySecrets_a, err5 := parseSecrets(mercurySecretsTOML_a)
		require.NoErrorf(t, err5, "error: %s", err5)
		mercurySecrets_b, err6 := parseSecrets(mercurySecretsTOML_b)
		require.NoErrorf(t, err6, "error: %s", err6)
		thresholdSecrets, err7 := parseSecrets(thresholdSecretsTOML)
		require.NoErrorf(t, err7, "error: %s", err7)

		opts := new(GeneralConfigOpts)
		configFiles := []string{utils.MakeTestFile(t, testConfigFileContents, "test.toml")}
		secretsFiles := []string{
			"testdata/mergingsecretsdata/secrets-database.toml",
			"testdata/mergingsecretsdata/secrets-explorer.toml",
			"testdata/mergingsecretsdata/secrets-password.toml",
			"testdata/mergingsecretsdata/secrets-pyroscope.toml",
			"testdata/mergingsecretsdata/secrets-prometheus.toml",
			"testdata/mergingsecretsdata/secrets-mercury_a.toml",
			"testdata/mergingsecretsdata/secrets-mercury_b.toml",
			"testdata/mergingsecretsdata/secrets-threshold.toml",
		}
		err = opts.Setup(configFiles, secretsFiles)
		require.NoErrorf(t, err, "error: %s", err)

		err = opts.parse()
		require.NoErrorf(t, err, "error testing: %s, %s", configFiles, secretsFiles)

		assert.Equal(t, databaseSecrets.Database.URL.URL().String(), opts.Secrets.Database.URL.URL().String())
		assert.Equal(t, databaseSecrets.Database.BackupURL.URL().String(), opts.Secrets.Database.BackupURL.URL().String())

		assert.Equal(t, (string)(*explorerSecrets.Explorer.AccessKey), (string)(*opts.Secrets.Explorer.AccessKey))
		assert.Equal(t, (string)(*explorerSecrets.Explorer.Secret), (string)(*opts.Secrets.Explorer.Secret))
		assert.Equal(t, (string)(*passwordSecrets.Password.Keystore), (string)(*opts.Secrets.Password.Keystore))
		assert.Equal(t, (string)(*passwordSecrets.Password.VRF), (string)(*opts.Secrets.Password.VRF))
		assert.Equal(t, (string)(*pyroscopeSecrets.Pyroscope.AuthToken), (string)(*opts.Secrets.Pyroscope.AuthToken))
		assert.Equal(t, (string)(*prometheusSecrets.Prometheus.AuthToken), (string)(*opts.Secrets.Prometheus.AuthToken))
		assert.Equal(t, (string)(*thresholdSecrets.Threshold.ThresholdKeyShare), (string)(*opts.Secrets.Threshold.ThresholdKeyShare))

		assertDeepEqualityMercurySecrets(*merge(mercurySecrets_a.Mercury, mercurySecrets_b.Mercury), opts.Secrets.Mercury)
	})
}

func parseSecrets(secrets string) (*Secrets, error) {
	var s Secrets
	if err := v2.DecodeTOML(strings.NewReader(secrets), &s); err != nil {
		return nil, fmt.Errorf("failed to decode secrets TOML: %w", err)
	}

	return &s, nil
}

func assertDeepEqualityMercurySecrets(expected v2.MercurySecrets, actual v2.MercurySecrets) error {
	if len(expected.Credentials) != len(actual.Credentials) {
		return fmt.Errorf("maps are not equal in length: len(expected): %d, len(actual): %d", len(expected.Credentials), len(actual.Credentials))
	}

	for key, value := range expected.Credentials {
		if actual.Credentials[key] != value {
			return fmt.Errorf("maps are not equal: expected[%s] = %s, actual[%s] = %s", key, value, key, actual.Credentials[key])
		}
	}
	return nil
}

func merge(map1 v2.MercurySecrets, map2 v2.MercurySecrets) *v2.MercurySecrets {
	combinedMap := make(map[string]v2.MercuryCredentials)

	for key, value := range map1.Credentials {
		combinedMap[key] = value
	}

	for key, value := range map2.Credentials {
		combinedMap[key] = value
	}

	return &v2.MercurySecrets{Credentials: combinedMap}
}
