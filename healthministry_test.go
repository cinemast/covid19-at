package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

var e = newHealthMinistryExporter()

func TestBezirke(t *testing.T) {
	result, err := e.getBezirkMetric()
	assert.Nil(t, err)
	assert.True(t, len(result) > 10, len(result))

	for _, s := range result {
		assert.Equal(t, 4, len(*s.Tags), s.Tags)
	}
}

func TestBundesland(t *testing.T) {
	result, err := e.getBundeslandInfectedMetric()
	assert.Nil(t, err)
	assert.True(t, len(result) == 3*9, len(result))

	for _, s := range result {
		assert.Equal(t, 4, len(*s.Tags), s.Tags)
	}

	vienna := result.findMetric("cov19_detail", "province=Wien")
	assert.NotNil(t, vienna)
	assert.Equal(t, (*vienna.Tags)["country"], "Austria")
	assert.Equal(t, (*vienna.Tags)["latitude"], "48.206351")
	assert.Equal(t, (*vienna.Tags)["longitude"], "16.374817")

	assert.NotNil(t, result.findMetric("cov19_detail_infection_rate", "province=Salzburg"))

	infectionRate := result.findMetric("cov19_detail_infection_rate", "province=Wien")
	assert.NotNil(t, infectionRate)
	assert.True(t, infectionRate.Value > 0 && infectionRate.Value < 1, infectionRate.Value)

	infected100k := result.findMetric("cov19_detail_infected_per_100k", "province=Wien")
	assert.NotNil(t, infected100k)
	assert.True(t, infected100k.Value > 5 && infected100k.Value < 100, infected100k.Value)
}

func TestAltersverteilung(t *testing.T) {
	result, err := e.getAgeMetrics()
	assert.Nil(t, err)
	assert.True(t, len(result) >= 4, len(result))

	for _, s := range result {
		assert.Equal(t, 2, len(*s.Tags), s.Tags)
	}
}

func TestGeschlechtsVerteilung(t *testing.T) {
	result, _ := e.getGeschlechtsVerteilung()
	assert.Equal(t, 2, len(result))
	assert.Equal(t, int64(100), int64(result[0].Value+result[1].Value))
}

func TestSimpleData(t *testing.T) {
	result, err := e.getSimpleData()
	assert.Equal(t, 0, len(err))
	assert.NotNil(t, result.findMetric("cov19_confirmed", ""))
}

func TestHealthMinistryHealth(t *testing.T) {
	errors := e.Health()
	assert.Equal(t, 0, len(errors))
}

func TestHealthMinistryGetMetrics(t *testing.T) {
	result, err := e.GetMetrics()
	assert.Nil(t, err, err)
	assert.NotNil(t, result)
}
