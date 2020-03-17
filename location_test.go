package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

var p = NewLocationProvider()

func TestAustriaLocations(t *testing.T) {
	assert.NotNil(t, p.GetLocation("Wien"))
	assert.NotNil(t, p.GetLocation("Kärnten"))
	assert.NotNil(t, p.GetLocation("Steiermark"))
	assert.NotNil(t, p.GetLocation("Burgenland"))
	assert.NotNil(t, p.GetLocation("Niederösterreich"))
	assert.NotNil(t, p.GetLocation("Oberösterreich"))
	assert.NotNil(t, p.GetLocation("Salzburg"))
	assert.NotNil(t, p.GetLocation("Tirol"))
	assert.NotNil(t, p.GetLocation("Vorarlberg"))
}

func TestLocationsForMetrics(t *testing.T) {
	metrics, err := NewEcdcExporter(p).GetMetrics()
	assert.Nil(t, err)
	for _, m := range metrics {
		country := (*m.Tags)["country"]
		assert.NotNil(t, p.GetLocation(country), "Country lookup: "+country)
	}
}
