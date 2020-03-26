package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMinistryStats(t *testing.T) {
	ministry := newSocialMinistryExporter(newMetadataProvider())
	result, err := ministry.GetMetrics()

	assert.Nil(t, err)
	assert.True(t, len(result) >= 10)

	totalConfirmed := result.findMetric("cov19_confirmed", "")
	assert.NotNil(t, totalConfirmed)
	assert.True(t, totalConfirmed.Value > 1000)

	tests := result.findMetric("cov19_tests", "")
	assert.NotNil(t, tests)
	assert.True(t, tests.Value > 1000)

	//healed := result.findMetric("cov19_healed", "")
	//assert.NotNil(t, healed)
	//assert.True(t, healed.Value > 5)

	dead := result.findMetric("cov19_dead", "")
	assert.NotNil(t, dead)
	assert.True(t, dead.Value >= 4)
}

func TestHospitalized(t *testing.T) {
	ministry := newSocialMinistryExporter(newMetadataProvider())
	result, err := ministry.getHospitalizedMetrics()

	assert.Nil(t, err)
	assert.True(t, len(result) > 0, len(result))
	assert.NotNil(t, result.findMetric("cov19_hospitalized", ""))
	assert.NotNil(t, result.findMetric("cov19_intensive_care", ""))
	assert.NotNil(t, result.findMetric("cov19_intensive_care_detail", "province=Wien"))

}
