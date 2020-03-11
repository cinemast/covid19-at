package main

import (
	"bufio"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func loadHTML(path string) io.Reader {
	file, _ := os.Open(path)
	return bufio.NewReader(file)
}

func TestSummaryError(t *testing.T) {
	result := parseStats(loadHTML("test/summary_error.html"))
	assert.Equal(t, 0, result.confirmed)
	assert.Equal(t, 0, result.tests)
	assert.Equal(t, 0, result.healed)
}

func TestSummarySuccess(t *testing.T) {
	result := parseStats(loadHTML("test/austria_stats_2020_03_10.html"))
	assert.Equal(t, 182, result.confirmed)
	assert.Equal(t, 5026, result.tests)
	assert.Equal(t, 4, result.healed)
}

func TestProvinceError(t *testing.T) {
	result := parseProvinceStats(loadHTML("test/summary_error.html"))
	assert.Equal(t, 0, len(result))
}

func TestProvinceSuccess(t *testing.T) {
	result := parseProvinceStats(loadHTML("test/austria_stats_2020_03_10.html"))
	assert.Equal(t, 9, len(result))
	assert.Equal(t, 40, result[0].count)
	assert.Equal(t, "Niederösterreich", result[0].name)
}

func TestWorldError(t *testing.T) {
	result := parseWorldStats(loadHTML("test/summary_error.html"))
	assert.Equal(t, 0, len(result))
}

func TestWorldSuccess(t *testing.T) {
	result := parseWorldStats(loadHTML("test/world_stats_2020_03_10.html"))
	assert.Equal(t, 105, len(result))
	assert.Equal(t, "China", result[0].country)
	assert.Equal(t, "Asia", result[0].continent)
	assert.Equal(t, 3139, result[0].deaths)
	assert.Equal(t, 80879, result[0].infected)

	assert.Equal(t, "Togo", result[len(result)-1].country)
	assert.Equal(t, "Africa", result[len(result)-1].continent)
	assert.Equal(t, 1, result[len(result)-1].infected)
	assert.Equal(t, 0, result[len(result)-1].deaths)
}

func TestHealth(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(handleHealth))
	defer ts.Close()
	response, err := ts.Client().Get(ts.URL)
	assert.Nil(t, err)
	assert.Equal(t, 200, response.StatusCode)
	greeting, err := ioutil.ReadAll(response.Body)
	assert.Equal(t, "<html><body><img width=\"500\" src=\"https://spiessknafl.at/helth.png\"/></body></html>", string(greeting))
}

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

}

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
}
