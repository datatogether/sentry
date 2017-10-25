package main

import (
	"fmt"
	conf "github.com/datatogether/config"
	"github.com/datatogether/core"
	"os"
	"path/filepath"
	"time"
)

// server modes
const (
	DEVELOP_MODE    = "develop"
	PRODUCTION_MODE = "production"
	TEST_MODE       = "test"
)

// config holds all configuration for the server. It pulls from three places (in order):
// 		1. environment variables
// 		2. .[MODE].env OR .env
//
// globally-set env variables win.
// it's totally fine to not have, say, .env.develop defined, and just
// rely on a base ".env" file. But if you're in production mode & ".env.production"
// exists, that will be read *instead* of .env
//
// configuration is read at startup and cannot be alterd without restarting the server.
type config struct {
	Debug bool
	// port to listen on, will be read from PORT env variable if present.
	Port string

	// root url for service
	UrlRoot string

	// url of postgres app db
	PostgresDbUrl string

	// Public Key to use for signing metablocks. required.
	PublicKey string

	// TLS (HTTPS) enable support via LetsEncrypt, default false
	// should be true in production
	TLS bool

	// How long before a url is considered stale, in hours.
	StaleDurationHours int

	// crawl urls?
	Crawl bool
	// Weather or not the crawler respects robots.txt
	Polite bool
	// how long to wait between requests. one day this'll be dynamically
	// modifiable
	CrawlDelaySeconds int
	// Content Types to Store
	StoreContentTypes []string

	// read from env variable: AWS_REGION
	// the region your bucket is in, eg "us-east-1"
	AwsRegion string
	// read from env variable: AWS_S3_BUCKET_NAME
	// should be just the name of your bucket, no protocol prefixes or paths
	AwsS3BucketName string
	// read from env variable: AWS_ACCESS_KEY_ID
	AwsAccessKeyId string
	// read from env variable: AWS_SECRET_ACCESS_KEY
	AwsSecretAccessKey string
	// path to store & retrieve data from
	AwsS3BucketPath string

	// seed        = flag.String("seed", "", "seed URL")
	// cancelAfter = flag.Duration("cancelafter", 0, "automatically cancel the fetchbot after a given time")
	// cancelAtURL = flag.String("cancelat", "", "automatically cancel the fetchbot at a given URL")
	// stopAfter   = flag.Duration("stopafter", 0, "automatically stop the fetchbot after a given time")
	// stopAtURL   = flag.String("stopat", "", "automatically stop the fetchbot at a given URL")
	// memStats    = flag.Duration("memstats", 0, "display memory statistics at a given interval")

	// setting HTTP_AUTH_USERNAME & HTTP_AUTH_PASSWORD
	// will enable basic http auth for the server. This is a single
	// username & password that must be passed in with every request.
	// leaving these values blank will disable http auth
	// read from env variable: HTTP_AUTH_USERNAME
	HttpAuthUsername string
	// read from env variable: HTTP_AUTH_PASSWORD
	HttpAuthPassword string

	// if true, requests that have X-Forwarded-Proto: http will be redirected
	// to their https variant
	ProxyForceHttps bool
	// CertbotResponse is only for doing manual SSL certificate generation via LetsEncrypt.
	CertbotResponse string
}

// StaleDuration turns cfg.StaleDurationHours into a time.Duration
func (cfg *config) StaleDuration() time.Duration {
	return 72 * time.Hour
	// return cfg.StaleDurationHours * time.Hour
}

// initConfig pulls configuration from config.json
func initConfig(mode string) (cfg *config, err error) {
	cfg = &config{}

	if path := configFilePath(mode, cfg); path != "" {
		log.Infof("loading config file: %s", filepath.Base(path))
		if err := conf.Load(cfg, path); err != nil {
			log.Info("error loading config:", err)
		}
	} else {
		if err := conf.Load(cfg, path); err != nil {
			log.Info("error loading config:", err)
		}
	}

	// make sure port is set
	if cfg.Port == "" {
		cfg.Port = "8080"
	}

	err = requireConfigStrings(map[string]string{
		"PORT":            cfg.Port,
		"POSTGRES_DB_URL": cfg.PostgresDbUrl,
		"PUBLIC_KEY":      cfg.PublicKey,
	})

	// transfer settings to core library
	core.AwsRegion = cfg.AwsRegion
	core.AwsAccessKeyId = cfg.AwsAccessKeyId
	core.AwsS3BucketName = cfg.AwsS3BucketName
	core.AwsS3BucketPath = cfg.AwsS3BucketPath
	core.AwsSecretAccessKey = cfg.AwsSecretAccessKey

	return
}

// requireConfigStrings panics if any of the passed in values aren't set
func requireConfigStrings(values map[string]string) error {
	for key, value := range values {
		if value == "" {
			return fmt.Errorf("%s env variable or config key must be set", key)
		}
	}

	return nil
}

func packagePath(path string) string {
	return filepath.Join(os.Getenv("GOPATH"), "src/github.com/datatogether/sentry", path)
}

// checks for .[mode].env file to read configuration from if the file exists
// defaults to .env, returns "" if no file is present
func configFilePath(mode string, cfg *config) string {
	fileName := packagePath(fmt.Sprintf(".%s.env", mode))
	if !fileExists(fileName) {
		fileName = packagePath(".env")
		if !fileExists(fileName) {
			return ""
		}
	}
	return fileName
}

// Does this file exist?
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// outputs any notable settings to stdout
func printConfigInfo() {
	// TODO
}
