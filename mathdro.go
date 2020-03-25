package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"
)

type mathdroExporter struct {
	url string
}

type recoveredStats []struct {
	ProvinceState *string
	CountryRegion string
	Recovered     uint64
	Lat           float64
	Long          float64
}

func newMathdroExporter() *mathdroExporter {
	return &mathdroExporter{url: "https://covid19.mathdro.id/api/"}
}

func (me *mathdroExporter) GetMetrics() (metrics, error) {
	recovered, err := me.getRecoveredStats()
	if err != nil {
		return nil, err
	}
	result := make(metrics, 0)

	for _, r := range recovered {
		tags := map[string]string{"country": r.CountryRegion, "latitude": ftos(r.Lat), "longitude": ftos(r.Long)}
		if r.ProvinceState != nil {
			tags["province"] = *r.ProvinceState
		}
		result = append(result, metric{Name: "cov19_world_recovered", Tags: &tags, Value: float64(r.Recovered)})
	}
	return result, nil
}

func (me *mathdroExporter) Health() []error {
	_, err := me.GetMetrics()
	if err != nil {
		return []error{err}
	}
	return nil
}

func (me *mathdroExporter) getRecoveredStats() (recoveredStats, error) {
	client := http.Client{Timeout: 5 * time.Second}
	response, err := client.Get(me.url + "recovered")
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	jsonString, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	recoveredStats := make(recoveredStats, 0)
	err = json.Unmarshal(jsonString, &recoveredStats)
	if err != nil {
		return nil, err
	}
	return recoveredStats, nil
}
