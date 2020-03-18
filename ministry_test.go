package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMinistryStats(t *testing.T) {
	ministry := NewMinistryExporter(NewMetadataProvider())
	result, err := ministry.GetMetrics()

	assert.Nil(t, err)
	assert.True(t, len(result) >= 10)

	totalConfirmed := result.FindMetric("cov19_confirmed", "")
	assert.NotNil(t, totalConfirmed)
	assert.True(t, totalConfirmed.Value > 1000)

	tests := result.FindMetric("cov19_tests", "")
	assert.NotNil(t, tests)
	assert.True(t, tests.Value > 1000)

	healed := result.FindMetric("cov19_healed", "")
	assert.NotNil(t, healed)
	assert.True(t, healed.Value > 5)

	dead := result.FindMetric("cov19_dead", "")
	assert.NotNil(t, dead)
	assert.True(t, dead.Value >= 4)

	vienna := result.FindMetric("cov19_detail", "province=Wien")
	assert.NotNil(t, vienna)
	assert.Equal(t, (*vienna.Tags)["country"], "Austria")
	assert.Equal(t, (*vienna.Tags)["latitude"], "48.206351")
	assert.Equal(t, (*vienna.Tags)["longitude"], "16.374817")

	infectionRate := result.FindMetric("cov19_detail_infection_rate", "province=Wien")
	assert.NotNil(t, infectionRate)
	assert.True(t, infectionRate.Value > 0 && infectionRate.Value < 1, infectionRate.Value)

	infected100k := result.FindMetric("cov19_detail_infected_per_100k", "province=Wien")
	assert.NotNil(t, infected100k)
	assert.True(t, infected100k.Value > 5 && infected100k.Value < 100, infected100k.Value)
}
