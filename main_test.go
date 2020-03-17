package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHealth(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(handleHealth))
	defer ts.Close()
	response, err := ts.Client().Get(ts.URL)
	assert.Nil(t, err)
	assert.Equal(t, 200, response.StatusCode)
	greeting, err := ioutil.ReadAll(response.Body)

	assert.Nil(t, err)
	assert.Equal(t, "<html><body><img width=\"500\" src=\"https://spiessknafl.at/helth.png\"/></body></html>", string(greeting))
}

func TestMetrics(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(handleMetrics))
	defer ts.Close()
	response, err := ts.Client().Get(ts.URL)
	assert.Nil(t, err)
	assert.Equal(t, 200, response.StatusCode)
	metricsString, err := ioutil.ReadAll(response.Body)
	assert.Nil(t, err)

	metricResult := string(metricsString)
	assert.True(t, strings.Contains(metricResult, "cov19_tests"))
	assert.True(t, strings.Contains(metricResult, "cov19_confirmed"))
	assert.True(t, strings.Contains(metricResult, "cov19_healed"))
	assert.True(t, strings.Contains(metricResult, "cov19_world_infected"))
	assert.True(t, strings.Contains(metricResult, "cov19_world_death"))
	assert.True(t, strings.Contains(metricResult, "cov19_detail"))
	assert.True(t, strings.Contains(metricResult, "cov19_detail_dead"))
}
