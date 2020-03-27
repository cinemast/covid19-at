package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeName(t *testing.T) {
	assert.Equal(t, "Saudi Arabia", normalizeCountryName("Saudi_Arabia"))
	assert.Equal(t, "Canada", normalizeCountryName("CANADA"))
	assert.Equal(t, "United States of America", normalizeCountryName("United_States_of_America"))
	assert.Equal(t, "Antigua and Barbuda", normalizeCountryName("Antigua_and_Barbuda"))
}

func TestEcdcStats(t *testing.T) {

	ecdc := newEcdcExporter(newMetadataProvider())
	result, err := ecdc.GetMetrics()

	assert.Nil(t, err)
	assert.True(t, len(result) > 0)

	china := result.findMetric("cov19_world_death", "country=China")
	assert.NotNil(t, china)
	//assert.Equal(t, (*china.Tags)["continent"], "Asia")
	assert.Equal(t, (*china.Tags)["latitude"], "35.861660")
	assert.Equal(t, (*china.Tags)["longitude"], "104.195397")
	assert.True(t, china.Value > 3000)

	china = result.findMetric("cov19_world_infected", "country=China")
	assert.NotNil(t, china)
	//assert.Equal(t, (*china.Tags)["continent"], "Asia")
	assert.Equal(t, (*china.Tags)["latitude"], "35.861660")
	assert.Equal(t, (*china.Tags)["longitude"], "104.195397")
	assert.True(t, china.Value > 10000)

	china = result.findMetric("cov19_world_infected", "country=Bosnia and Herzegovina")
	assert.NotNil(t, china)
	assert.Equal(t, (*china.Tags)["continent"], "Europe")
	assert.Equal(t, (*china.Tags)["latitude"], "43.915886")
	assert.Equal(t, (*china.Tags)["longitude"], "17.679076")
	assert.True(t, china.Value > 10)
}
