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

func emptyPage(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("<html></html>"))
}

func TestErrors(t *testing.T) {

	ecdcURL := ecdcExporter.url
	ministryURL := ministryExporter.url

	mockServer := httptest.NewServer(http.HandlerFunc(emptyPage))
	defer mockServer.Close()
	ecdcExporter.url = mockServer.URL
	ministryExporter.url = mockServer.URL

	ts := httptest.NewServer(http.HandlerFunc(handleHealth))

	defer ts.Close()
	response, err := ts.Client().Get(ts.URL)
	assert.Nil(t, err)
	assert.Equal(t, 500, response.StatusCode)
	greeting, err := ioutil.ReadAll(response.Body)

	assert.Nil(t, err)
	assert.Equal(t, "<html><body><img width=\"500\" src=\"https://spiessknafl.at/fine.jpg\"/><pre>World stats are failing\nCould not find \"Bestätigte Fälle\"\nMissing ministry stats\nCould not find metric cov19_confirmed / ()\nCould not find metric cov19_tests / ()\nCould not find metric cov19_healed / ()\n</pre></body></html>", string(greeting))

	ecdcExporter.url = ecdcURL
	ministryExporter.url = ministryURL
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
	//assert.True(t, strings.Contains(metricResult, "cov19_detail_dead"))
}
