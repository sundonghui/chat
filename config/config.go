package config

import (
	"path/filepath"
	"strings"
)

type Configuration struct {
	// Server
	Server struct {
		KeepAlivePeriodSeconds int    `yaml:"keepaliveperiodseconds"`
		ListenAddr             string `yaml:"listenaddr" default:""`
		Port                   int    `yaml:"port" default:"80"`

		SSL struct {
			Enabled         *bool  `yaml:"enabled" default:"false"`
			RedirectToHTTPS *bool  `yaml:"redirecttohttps"  default:"true"`
			ListenAddr      string `yaml:"listenaddr" default:""`
			Port            int    `yaml:"port" default:"443"`
			CertFile        string `yaml:"certfile" default:""`
			CertKey         string `yaml:"certkey" default:""`
			LetsEncrypt     struct {
				Enabled   *bool    `yaml:"enabled" default:"false"`
				AcceptTOS *bool    `yaml:"accepttos" default:"false"`
				Cache     string   `yaml:"cache" default:"data/certs"`
				Hosts     []string `yaml:"hosts"`
			} `yaml:"letsencrypt"`
		} `yaml:"ssl"`

		ResponseHeaders map[string]string `yaml:"responseheaders"`
		TrustedProxies  []string          `yaml:"trustedproxies"`
		Cors            struct {
			AllowOrigins []string `yaml:"alloworigins"`
			AllowMethods []string `yaml:"allowmethods"`
			AllowHeaders []string `yaml:"allowheaders"`
		} `yaml:"cors"`
		Stream struct {
			PingPeriodSeconds int      `yaml:"pingperiodseconds" default:"45"`
			AllowedOrigins    []string `yaml:"allowedorigins"`
		} `yaml:"stream"`
	} `yaml:"server"`

	// Database
	Database struct {
		Dialect    string `yaml:"dialect" default:"sqlite3"`           // sqlite3, mysql, postgres
		Connection string `yaml:"connection" default:"data/gotify.db"` // sqlite3: data/gotify.db, mysql: user:password@tcp(127.0.0.1:3306)/gotify, postgres: user:password@tcp(127.0.0.1:5432)/gotify
	} `yaml:"database"`
	// Security
	DefaultUser struct {
		Name string `yaml:"name" default:"admin"`
		Pass string `yaml:"pass" default:"password"`
	} `yaml:"defaultuser"`
	PassStrength      int    `yaml:"passstrength" default:"8"`
	UploadedImagesDir string `yaml:"uploadedimagesdir" default:"data/images"`
	PluginsDir        string `yaml:"pluginsdir" default:"data/plugins"`
	Registration      bool   `yaml:"registration" default:"true"`
}

func configFiles() []string {
	return []string{"config.yml"}
}

// Get returns the configuration extracted from env variables or config file.
func Get() *Configuration {
	conf := new(Configuration)
	err := New(&Config{EnvironmentPrefix: "chat"}).Load(conf, configFiles()...)
	if err != nil {
		panic(err)
	}
	addTrailingSlashToPaths(conf)
	return conf
}

func addTrailingSlashToPaths(conf *Configuration) {
	if !strings.HasSuffix(conf.UploadedImagesDir, "/") && !strings.HasSuffix(conf.UploadedImagesDir, "\\") {
		conf.UploadedImagesDir += string(filepath.Separator)
	}
}
