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

	tests := result.findMetric("cov19_tests", "")
	assert.NotNil(t, tests)
	assert.True(t, tests.Value > 1000)

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
