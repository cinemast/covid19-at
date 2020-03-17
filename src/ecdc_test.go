package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEcdcStats(t *testing.T) {

	ecdc := NewEcdcExporter(nil)
	result, err := ecdc.GetMetrics()

	assert.Nil(t, err)
	assert.True(t, len(result) > 0)
	assert.True(t, len(result)%2 == 0)

	china := result.FindMetric("cov19_world_death", "country=China")
	assert.NotNil(t, china)
	assert.Equal(t, (*china.Tags)["continent"], "Asia")
	assert.True(t, china.Value > 3000)

	china = result.FindMetric("cov19_world_infected", "country=China")
	assert.NotNil(t, china)
	assert.Equal(t, (*china.Tags)["continent"], "Asia")
	assert.True(t, china.Value > 10000)
}
