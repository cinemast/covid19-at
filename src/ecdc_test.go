package main

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestEcdcStats(t *testing.T) {

	ecdc := NewEcdcExporter(nil)
	result, err := ecdc.GetMetrics()

	assert.Nil(t, err)
	assert.True(t, len(result) > 0)
	assert.True(t, len(result) % 2 == 0)

	assert.Equal(t, (*result[0].Tags)["country"], "China")
	assert.Equal(t, (*result[0].Tags)["continent"], "Asia")
	assert.Equal(t, result[0].Name, "cov19_world_death")
	assert.True(t, result[0].Value > 3000)

	assert.Equal(t, (*result[1].Tags)["country"], "China")
	assert.Equal(t, (*result[1].Tags)["continent"], "Asia")
	assert.Equal(t, result[1].Name, "cov19_world_infected")
	assert.True(t, result[1].Value > 10000)
}
