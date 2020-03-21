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

//func TestErrors(t *testing.T) {
//	healthMinistryExporter := exporters[0].(*healthMinistryExporter)
//	socialMinistry := exporters[1].(*SocialMinistryExporter)
//	ecdcExporter := exporters[2].(*EcdcExporter)
//
//	ecdcURL := ecdcExporter.Url
//	ministryURL := socialMinistry.Url
//	healthMinistryURL := healthMinistryExporter.url
//
//	mockServer := httptest.NewServer(http.HandlerFunc(emptyPage))
//	defer mockServer.Close()
//	ecdcExporter.Url = mockServer.URL
//	socialMinistry.Url = mockServer.URL
//	healthMinistryExporter.url = mockServer.URL
//
//	ts := httptest.NewServer(http.HandlerFunc(handleHealth))
//
//	defer ts.Close()
//	response, err := ts.Client().Get(ts.URL)
//	assert.Nil(t, err)
//	assert.Equal(t, 500, response.StatusCode)
//	greeting, err := ioutil.ReadAll(response.Body)
//
//	assert.Nil(t, err)
//	assert.Equal(t, "<html><body><img width=\"500\" src=\"https://spiessknafl.at/fine.jpg\"/><pre>Could not find \"Bestätigte Fälle\"\nMissing ministry stats\nCould not find metric cov19_confirmed / ()\nCould not find metric cov19_tests / ()\nCould not find metric cov19_healed / ()\nWorld stats are failing\n</pre></body></html>", string(greeting))
//
//	ecdcExporter.Url = ecdcURL
//	socialMinistry.Url = ministryURL
//	healthMinistryExporter.url = healthMinistryURL
//}

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
