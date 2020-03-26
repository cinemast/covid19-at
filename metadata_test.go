package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var p = newMetadataProvider()

func TestAustriaLocations(t *testing.T) {
	assert.NotNil(t, p.getLocation("Wien"))
	assert.NotNil(t, p.getLocation("Kärnten"))
	assert.NotNil(t, p.getLocation("Steiermark"))
	assert.NotNil(t, p.getLocation("Burgenland"))
	assert.NotNil(t, p.getLocation("Niederösterreich"))
	assert.NotNil(t, p.getLocation("Oberösterreich"))
	assert.NotNil(t, p.getLocation("Salzburg"))
	assert.NotNil(t, p.getLocation("Tirol"))
	assert.NotNil(t, p.getLocation("Vorarlberg"))
}

func TestLocationsForMetrics(t *testing.T) {
	metrics, err := newEcdcExporter(p).GetMetrics()
	assert.Nil(t, err)
	for _, m := range metrics {
		country := (*m.Tags)["country"]
		assert.NotNil(t, p.getLocation(country), "Country lookup: "+country)
	}
}

func TestLocationsPopulationForMetrics(t *testing.T) {
	metrics, err := newEcdcExporter(p).GetMetrics()
	assert.Nil(t, err)
	for _, m := range metrics {
		country := (*m.Tags)["country"]
		assert.True(t, p.getPopulation(country) > 0, "Population lookup: "+country)
	}
}

func TestMetadataForBezirke(t *testing.T) {
	healthMinistryExporter := newHealthMinistryExporter()
	metrics, err := healthMinistryExporter.getBezirkMetric()
	assert.Nil(t, err)
	for _, m := range metrics {
		bezirk := (*m.Tags)["bezirk"]
		assert.True(t, healthMinistryExporter.mp.getPopulation(bezirk) > 0, "Population lookup: "+bezirk)
		assert.True(t, healthMinistryExporter.mp.getLocation(bezirk) != nil, "location lookup: "+bezirk)
	}
}

func TestUnknownLocation(t *testing.T) {
	assert.Nil(t, p.getLocation("xxxxx"))
	assert.Equal(t, uint64(0), p.getPopulation("xxxxx"))
	assert.Nil(t, p.getMetadata("xxxxx"))

	assert.Nil(t, newMetadataProviderWithFilename("someinvalidfile"))
}
