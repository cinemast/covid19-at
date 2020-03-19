package exporter

import (
	"encoding/json"
	"strings"
)

var grafanaQueryUrl = "https://info.gesundheitsministerium.at/api/tsdb/query"

type grafanaExporter struct {
	url string
	mp  *MetadataProvider
}

type grafanaResult struct {
	Results struct {
		A struct {
			Tables []struct {
				Rows [][]interface{} `json:"rows"`
			} `json:"tables"`
		} `json:"A"`
	} `json:"results"`
}

func NewGrafanaExporter() *grafanaExporter {
	return &grafanaExporter{url: grafanaQueryUrl, mp: NewMetadataProviderWithFilename("bezirke.csv")}
}

func (g *grafanaExporter) getTags(location string) *map[string]string {
	data := g.mp.GetMetadata(location)
	if data == nil {
		return &map[string]string{"bezirk": location, "country": "Austria"}
	}

	return &map[string]string{"bezirk": location, "country": "Austria", "latitude": ftos(data.location.lat), "longitude": ftos(data.location.long)}
}

func (g *grafanaExporter) GetMetrics() (Metrics, error) {
	json, _ := readJsonFromFile("response.json")
	data, err := getBezirkData(json)
	if err != nil {
		return nil, err
	}

	result := make(Metrics, len(data))
	for i, v := range data {
		result[i].Value = float64(v.infected)
		result[i].Name = "cov19_bezirk_infected"
		result[i].Tags = g.getTags(v.location)
	}
	return result, nil

}

func getBezirkData(jsonString []byte) ([]CovidStat, error) {
	grafanaResult := grafanaResult{}
	err := json.Unmarshal(jsonString, &grafanaResult)
	if err != nil {
		return nil, err
	}

	rows := grafanaResult.Results.A.Tables[0].Rows
	stats := make([]CovidStat, len(rows))
	for i, v := range rows {
		stats[i].location = strings.TrimSpace(v[0].(string))
		stats[i].infected = uint64(v[1].(float64))
	}
	return stats, nil
}
