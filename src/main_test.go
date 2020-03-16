package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"strings"
	"github.com/stretchr/testify/assert"
)


func TestHealth(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(handleHealth))
	defer ts.Close()
	response, err := ts.Client().Get(ts.URL)
	assert.Nil(t, err)
	assert.Equal(t, 200, response.StatusCode)
	greeting, err := ioutil.ReadAll(response.Body)
	assert.Equal(t, "<html><body><img width=\"500\" src=\"https://spiessknafl.at/helth.png\"/></body></html>", string(greeting))
}
/*
func TestFailingHealth(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("<html></html>"))
	}))
	defer ts.Close()
	ecdcURLOrig := ecdcURL
	healthministryURLOrig := healthministryURL

	ecdcURL = ts.URL
	healthministryURL = ts.URL

	ts2 := httptest.NewServer(http.HandlerFunc(handleHealth))
	defer ts2.Close()

	response, err := ts2.Client().Get(ts2.URL)

	ecdcURL = ecdcURLOrig
	healthministryURL = healthministryURLOrig

	assert.Nil(t, err)
	assert.Equal(t, 500, response.StatusCode)
	errorDescription, err := ioutil.ReadAll(response.Body)
	assert.Equal(t,
		"<html><body><img width=\"500\" src=\"https://spiessknafl.at/fine.jpg\"/><pre>Summary confirmed are failing\nSummary healed are failing\nSummary tests are failing\nDetails Austria are failing\nWorld stats are failing\n</pre></body></html>", string(errorDescription))

}*/

func TestMetrics(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(handleMetrics))
	defer ts.Close()
	response, err := ts.Client().Get(ts.URL)
	assert.Nil(t, err)
	assert.Equal(t, 200, response.StatusCode)
	metricsString, err := ioutil.ReadAll(response.Body)
	metricResult := string(metricsString)
	assert.True(t, strings.Contains(metricResult, "cov19_tests"))
	assert.True(t, strings.Contains(metricResult, "cov19_confirmed"))
	assert.True(t, strings.Contains(metricResult, "cov19_healed"))
	assert.True(t, strings.Contains(metricResult, "cov19_world_infected"))
	assert.True(t, strings.Contains(metricResult, "cov19_world_death"))
	assert.True(t, strings.Contains(metricResult, "cov19_detail"))
	assert.True(t, strings.Contains(metricResult, "cov19_detail_dead"))
}
