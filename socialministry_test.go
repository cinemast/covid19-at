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
}
