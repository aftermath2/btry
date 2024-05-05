// Package config contains the configuration schema and some utilities.
package config

import (
	"crypto/tls"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"
	"google.golang.org/grpc/credentials"
	"gopkg.in/macaroon.v2"
	"gopkg.in/yaml.v3"
)

// Config represents the configuration for the BTRY application.
type Config struct {
	Notifier  Notifier  `yaml:"notifier"`
	DB        DB        `yaml:"db"`
	Lottery   Lottery   `yaml:"lottery"`
	Tor       Tor       `yaml:"tor"`
	Lightning Lightning `yaml:"lightning"`
	API       API       `yaml:"api"`
	Server    Server    `yaml:"server"`
}

// API configuration.
type API struct {
	Logger      Logger      `yaml:"logger"`
	SSE         SSE         `yaml:"sse"`
	RateLimiter RateLimiter `yaml:"rate_limiter"`
}

// DB database configuration.
type DB struct {
	Path            string        `yaml:"path"`
	Logger          Logger        `yaml:"logger"`
	MaxIdleConns    int           `yaml:"max_idle_conns"`
	ConnMaxIdleTime time.Duration `yaml:"conn_max_idle_time"`
}

// Lightning configuration.
type Lightning struct {
	RPCAddress   string `yaml:"rpc_address"`
	TLSCertPath  string `yaml:"tls_cert_path"`
	MacaroonPath string `yaml:"macaroon_path"`
	Logger       Logger `yaml:"logger"`
	MaxFeePPM    int64  `yaml:"max_fee_ppm"`
}

// Logger configuration.
type Logger struct {
	Label   string `yaml:"label"`
	OutFile string `yaml:"out_file"`
	Level   uint8  `yaml:"level"`
}

// Lottery configuration.
type Lottery struct {
	Logger   Logger `yaml:"logger"`
	Duration uint32 `yaml:"duration"`
}

// Nostr configuration.
type Nostr struct {
	PrivateKey string   `yaml:"private_key"`
	Relays     []string `yaml:"relays"`
}

// Notifier configuration.
type Notifier struct {
	Telegram Telegram `yaml:"telegram"`
	Nostr    Nostr    `yaml:"nostr"`
	Logger   Logger   `yaml:"logger"`
	Enabled  bool     `yaml:"enabled"`
}

// RateLimiter configuration.
type RateLimiter struct {
	Tokens   uint64        `yaml:"tokens"`
	Interval time.Duration `yaml:"interval"`
}

// Server configuration.
type Server struct {
	Address         string            `yaml:"address"`
	TLSCertificates []tls.Certificate `yaml:"tls_certificates"`
	Logger          Logger            `yaml:"logger"`
	Timeout         struct {
		Read     time.Duration `yaml:"read"`
		Write    time.Duration `yaml:"write"`
		Shutdown time.Duration `yaml:"shutdown"`
		Idle     time.Duration `yaml:"idle"`
	} `yaml:"timeout"`
}

// SSE configuration.
type SSE struct {
	Logger   Logger        `yaml:"logger"`
	Deadline time.Duration `yaml:"deadline"`
}

// Telegram configuration.
type Telegram struct {
	BotAPIToken string `yaml:"bot_api_token"`
	BotName     string `yaml:"bot_name"`
}

// Tor configuration.
type Tor struct {
	Address string        `yaml:"address"`
	Timeout time.Duration `yaml:"timeout"`
}

// New returns a configuration object loaded from a file.
func New() (Config, error) {
	configPath := os.Getenv("BTRY_CONFIG")
	if configPath == "" {
		dir, err := os.Getwd()
		if err != nil {
			return Config{}, err
		}
		configPath = filepath.Join(dir, "btry.yml")
	}

	f, err := os.OpenFile(configPath, os.O_RDONLY, 0o600)
	if err != nil {
		return Config{}, errors.Wrap(err, "opening file")
	}
	defer f.Close()

	var config Config
	if err := yaml.NewDecoder(f).Decode(&config); err != nil {
		return Config{}, errors.Wrap(err, "decoding configuration")
	}

	if err := config.Validate(); err != nil {
		return Config{}, err
	}

	return config, nil
}

// Validate returns an error if the configuration is not valid.
func (c Config) Validate() error {
	if err := validateLoggers(
		c.API.Logger,
		c.API.SSE.Logger,
		c.DB.Logger,
		c.Lightning.Logger,
		c.Lottery.Logger,
		c.Server.Logger,
	); err != nil {
		return err
	}

	if _, err := credentials.NewClientTLSFromFile(c.Lightning.TLSCertPath, ""); err != nil {
		return errors.Wrap(err, "invalid tls certificate")
	}

	macBytes, err := os.ReadFile(c.Lightning.MacaroonPath)
	if err != nil {
		return errors.Wrap(err, "macaroon file missing")
	}

	mac := &macaroon.Macaroon{}
	if err := mac.UnmarshalBinary(macBytes); err != nil {
		return errors.Wrap(err, "invalid macaroon encoding")
	}

	if c.Lottery.Duration == 0 {
		return errors.New("invalid lottery duration, must be higher than zero")
	}

	return validateAddresses(c.Lightning.RPCAddress, c.Server.Address, c.Tor.Address)
}

func validateLoggers(loggers ...Logger) error {
	for _, logger := range loggers {
		// Not importing logger constants to avoid cycle
		if logger.Level < 0 || logger.Level > 5 {
			return errors.Errorf("invalid logger %q. Level should be between 0 and 5", logger.Label)
		}
	}

	return nil
}

func validateAddresses(addresses ...string) error {
	for _, address := range addresses {
		if !strings.HasPrefix(address, "http") {
			address = "http://" + address
		}
		if _, err := url.Parse(address); err != nil {
			return errors.Wrapf(err, "invalid address: %s", address)
		}
	}

	return nil
}
