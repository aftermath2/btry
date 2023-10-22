package config_test

import (
	"os"
	"testing"

	"github.com/aftermath2/BTRY/config"
	"github.com/aftermath2/BTRY/logger"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	key := "BTRY_CONFIG"
	initialValue := os.Getenv(key)
	os.Setenv(key, "./testdata/mock_config.yml")

	_, err := config.New()
	assert.NoError(t, err)

	os.Setenv(key, initialValue)
}

func TestValidate(t *testing.T) {
	validConfig := config.Config{
		Lightning: config.Lightning{
			RPCAddress: "127.0.0.1:10001",
			Logger: config.Logger{
				Level: uint8(logger.INFO),
			},
			TLSCertPath:  "./testdata/tls.cert",
			MacaroonPath: "./testdata/readonly.macaroon",
		},
		Lottery: config.Lottery{
			Time: "00:00",
		},
		Server: config.Server{
			Address: "127.0.0.1:4000",
		},
		Tor: config.Tor{
			Address: "127.0.0.1:9050",
		},
	}

	cases := []struct {
		getConfig func(c config.Config) config.Config
		desc      string
		fail      bool
	}{
		{
			desc: "Valid",
			getConfig: func(c config.Config) config.Config {
				return c
			},
			fail: false,
		},
		{
			desc: "Invalid logger level",
			getConfig: func(c config.Config) config.Config {
				c.Lightning.Logger.Level = 7
				return c
			},
			fail: true,
		},
		{
			desc: "Invalid TLS certificate path",
			getConfig: func(c config.Config) config.Config {
				c.Lightning.TLSCertPath = "tls"
				return c
			},
			fail: true,
		},
		{
			desc: "Invalid macaroon path",
			getConfig: func(c config.Config) config.Config {
				c.Lightning.MacaroonPath = "macaroon"
				return c
			},
			fail: true,
		},
		{
			desc: "Invalid time",
			getConfig: func(c config.Config) config.Config {
				c.Lottery.Time = "03/01/2009"
				return c
			},
			fail: true,
		},
		{
			desc: "Invalid address",
			getConfig: func(c config.Config) config.Config {
				c.Server.Address = "4000:127.0.0.1"
				return c
			},
			fail: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			configuration := tc.getConfig(validConfig)
			err := configuration.Validate()
			if tc.fail {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
