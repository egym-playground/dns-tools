package main

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/egymgmbh/dns-tools/config"
	"github.com/egymgmbh/dns-tools/rrdb"
)

func TestValidConfiguration(t *testing.T) {
	// Given
	loadedConfig := loadConfig("testdata/pass/config.yml")
	db, err := rrdb.NewFromDirectory(loadedConfig.ZoneDataDirectory)
	// When
	err = checkCnames(db, loadedConfig.ManagedZones)
	// Then
	assert.Nil(t, err)
}

func TestInvalidConfiguration(t *testing.T) {
	// Given
	loadedConfig := loadConfig("testdata/fail/config.yml")
	db, err := rrdb.NewFromDirectory(loadedConfig.ZoneDataDirectory)
	// When
	err = checkCnames(db, loadedConfig.ManagedZones)
	// Then
	assert.NotNil(t, err)
}

func loadConfig(configFileLocation string) *config.Config {
	loadedConfig, _ := config.New(configFileLocation)
	return loadedConfig
}
