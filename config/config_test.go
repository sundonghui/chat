package config

import (
	"os"
	"testing"

	"gopkg.in/yaml.v2"

	"github.com/stretchr/testify/assert"

	"github.com/sundonghui/chat/mode"
)

func Test_Load_Example(t *testing.T) {
	mode.Set(mode.Test)
	jsonData, err := os.ReadFile("../config.example.yml")
	assert.Nil(t, err)
	var config Configuration
	err = yaml.Unmarshal(jsonData, &config)
	assert.Nil(t, err)

	assert.Equal(t, 80, config.Server.Port)
	assert.Equal(t, 443, config.Server.SSL.Port)
	assert.Equal(t, 45, config.Server.Stream.PingPeriodSeconds)
	assert.Equal(t, "sqlite3", config.Database.Dialect)
	assert.Equal(t, "data/gotify.db", config.Database.Connection)
}
