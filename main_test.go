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
	healthMinistryExporter := exporters[0].(*healthMinistryExporter)
	socialMinistry := exporters[1].(*socialMinistryExporter)
	ecdcExporter := exporters[2].(*ecdcExporter)

	ecdcURL := ecdcExporter.Url
	ministryURL := socialMinistry.url
	healthMinistryURL := healthMinistryExporter.url

	mockServer := httptest.NewServer(http.HandlerFunc(emptyPage))
	defer mockServer.Close()
	ecdcExporter.Url = mockServer.URL
	socialMinistry.url = mockServer.URL
	healthMinistryExporter.url = mockServer.URL

	ts := httptest.NewServer(http.HandlerFunc(handleHealth))

	defer ts.Close()
	response, err := ts.Client().Get(ts.URL)
	assert.Nil(t, err)
	assert.Equal(t, 500, response.StatusCode)
	//greeting, err := ioutil.ReadAll(response.Body)

	assert.Nil(t, err)
	//assert.Equal(t, "<html><body><img width=\"500\" src=\"https://spiessknafl.at/fine.jpg\"/><pre>Could not find beginning of array\nNot enough Bezirke Results: 0\nCould not find beginning of array\nMissing Bundesland result 0\nCould not find beginning of array\nMissing age metrics\nCould not find beginning of array\nGeschlechtsverteilung failed\nCould not find \"Bestätigte Fälle\"\nCould not find \"Hospitalisiert\"\nCould not find \"Intensivstation\"\nCould not find \"Bestätigte Fälle\"\nCould not find \"Bestätigte Fälle\"\nMissing ministry stats\nCould not find metric cov19_healed / ()\nWorld stats are failing\n</pre></body></html>", string(greeting))

	ecdcExporter.Url = ecdcURL
	socialMinistry.url = ministryURL
	healthMinistryExporter.url = healthMinistryURL
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
	//assert.True(t, strings.Contains(metricResult, "cov19_healed"))
	assert.True(t, strings.Contains(metricResult, "cov19_world_infected"))
	assert.True(t, strings.Contains(metricResult, "cov19_world_death"))
	assert.True(t, strings.Contains(metricResult, "cov19_detail"))
	assert.True(t, strings.Contains(metricResult, "cov19_detail_dead"))
}

func TestApiOverall(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(handleApiTotal))
	defer ts.Close()
	_, err := ts.Client().Get(ts.URL)
	assert.Nil(t, err)
}

func TestApiBezirk(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(handleApiBezirk))
	defer ts.Close()
	_, err := ts.Client().Get(ts.URL)
	assert.Nil(t, err)
}

func TestApiBundesland(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(handleApiBundesland))
	defer ts.Close()
	_, err := ts.Client().Get(ts.URL)
	assert.Nil(t, err)
}
