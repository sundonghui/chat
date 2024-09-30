package config

import (
	"log"
	"os"
	"reflect"
	"strconv"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

type Configuration struct {
	// Server
	Server struct {
		KeepAlivePeriodSeconds int    `yaml:"keepaliveperiodseconds"`
		ListenAddr             string `yaml:"listenaddr" default:"127.0.0.1"`
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

// Get returns the configuration extracted from env variables or config file.
func Get() *Configuration {
	// Load config file
	currentDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error getting current directory: %v", err)
	}

	v := viper.New()
	v.SetConfigName("config.example")
	v.SetConfigType("yaml")
	v.AddConfigPath(currentDir)
	if err := v.ReadInConfig(); err != nil {
		log.Fatalf("config file not exists or read error %s", err)
	}

	config := new(Configuration)
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToSliceHookFunc(","),
			func(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
				if f.Kind() == reflect.Ptr && f.Elem().Kind() == reflect.Bool {
					if data == nil {
						return false, nil
					}
				}

				if t.Kind() == reflect.String && data == "" {
					return reflect.Zero(t).Interface(), nil
				}

				return data, nil
			},
		),
		Result: config,
	})
	if err != nil {
		log.Fatalf("Error creating decoder: %s", err)
	}

	if err := decoder.Decode(v.AllSettings()); err != nil {
		log.Fatalf("Error decoding config: %s", err)
	}

	return config
}

// initializeDefaults sets the default values for the configuration.
func initializeDefaults(cfg Configuration) Configuration {
	val := reflect.ValueOf(&cfg).Elem()
	setDefaults(val)
	return cfg
}

func setDefaults(val reflect.Value) {
	for i := 0; i < val.NumField(); i++ {
		field := val.Type().Field(i)
		defaultValue := field.Tag.Get("default")

		if field.Type.Kind() == reflect.Struct {
			setDefaults(val.Field(i))
			continue
		}

		if defaultValue != "" {
			setFieldValue(val.Field(i), defaultValue)
		}
	}
}

func setFieldValue(field reflect.Value, defaultValue string) {
	switch field.Kind() {
	case reflect.String:
		field.SetString(defaultValue)
	case reflect.Int:
		if intValue, err := strconv.Atoi(defaultValue); err == nil {
			field.SetInt(int64(intValue))
		}
	case reflect.Bool:
		if boolValue, err := strconv.ParseBool(defaultValue); err == nil {
			field.SetBool(boolValue)
		}
	case reflect.Ptr:
		if field.Type().Elem().Kind() == reflect.Bool {
			boolValue := defaultValue == "true"
			ptr := reflect.New(field.Type().Elem())
			ptr.Elem().SetBool(boolValue)
			field.Set(ptr)
		}
	case reflect.Slice:
		for i := 0; i < field.Len(); i++ {
			setFieldValue(field.Index(i), defaultValue)
		}
	}
}
