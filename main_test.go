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
	ecdcExporter := exporters[1].(*ecdcExporter)

	ecdcURL := ecdcExporter.Url
	healthMinistryURL := healthMinistryExporter.url

	mockServer := httptest.NewServer(http.HandlerFunc(emptyPage))
	defer mockServer.Close()
	ecdcExporter.Url = mockServer.URL
	healthMinistryExporter.url = mockServer.URL

	ts := httptest.NewServer(http.HandlerFunc(handleHealth))

	defer ts.Close()
	response, err := ts.Client().Get(ts.URL)
	assert.Nil(t, err)
	assert.Equal(t, 500, response.StatusCode)
	greeting, err := ioutil.ReadAll(response.Body)

	assert.Nil(t, err)
	assert.Equal(t, `<html><body><img width="500" src="https://spiessknafl.at/fine.jpg"/><pre>Could not find beginning of array
Not enough Bezirke Results: 0
Could not find beginning of array
Missing Bundesland result 0
Could not find beginning of array
Missing age metrics
Could not find beginning of array
Geschlechtsverteilung failed
Erkrankungen not found in /SimpleData.js
dpGenesen not found in /Genesen.js
dpTot not found in /Verstorben.js
dpGesNBBel not found in /GesamtzahlNormalbettenBel.js
dpGesIBBel not found in /GesamtzahlIntensivBettenBel.js
dpGesTestungen not found in /GesamtzahlTestungen.js
Could not find "Bestätigte Fälle"
World stats are failing
</pre></body></html>`, string(greeting))

	ecdcExporter.Url = ecdcURL
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
	assert.True(t, strings.Contains(metricResult, "cov19_healed"))
	assert.True(t, strings.Contains(metricResult, "cov19_world_infected"))
	assert.True(t, strings.Contains(metricResult, "cov19_world_death"))
	assert.True(t, strings.Contains(metricResult, "cov19_detail"))
	assert.True(t, strings.Contains(metricResult, "cov19_detail_dead"))
}
