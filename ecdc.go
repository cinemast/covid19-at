package main

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
	"unicode"

	"github.com/PuerkitoBio/goquery"
)

type ecdcExporter struct {
	Url string
	Mp  *metadataProvider
}

type ecdcStat struct {
	CovidStat
	continent string
}

func newEcdcExporter(lp *metadataProvider) *ecdcExporter {
	return &ecdcExporter{Url: "https://www.ecdc.europa.eu/en/geographical-distribution-2019-ncov-cases", Mp: lp}
}

//GetMetrics parses the ECDC table
func (e *ecdcExporter) GetMetrics() (metrics, error) {
	stats, err := getEcdcStat(e.Url)
	if err != nil {
		return nil, err
	}
	result := make([]metric, 0)
	for i := range stats {
		tags := e.getTags(stats, i)
		deaths := stats[i].deaths
		infected := stats[i].infected
		population := e.Mp.getPopulation(stats[i].location)
		if deaths > 0 {
			result = append(result, metric{Name: "cov19_world_death", Value: float64(deaths), Tags: &tags})
			if population > 0 {
				result = append(result, metric{Name: "cov19_world_fatality_rate", Value: fatalityRate(infected, deaths), Tags: &tags})
			}
		}
		result = append(result, metric{Name: "cov19_world_infected", Value: float64(infected), Tags: &tags})
		if population > 0 {
			result = append(result, metric{Name: "cov19_world_infection_rate", Value: infectionRate(infected, population), Tags: &tags})
			result = append(result, metric{Name: "cov19_world_infected_per_100k", Value: infection100k(infected, population), Tags: &tags})
		}
	}
	return result, nil
}

//Health checks the functionality of the exporter
func (e *ecdcExporter) Health() []error {
	errors := make([]error, 0)
	worldStats, _ := e.GetMetrics()

	if len(worldStats) < 200 {
		errors = append(errors, fmt.Errorf("World stats are failing"))
	}

	for _, m := range worldStats {
		country := (*m.Tags)["country"]
		if mp.getLocation(country) == nil {
			errors = append(errors, fmt.Errorf("Could not find location for country: %s", country))
		}
	}
	return errors
}

func normalizeCountryName(name string) string {
	name = strings.TrimSpace(name)
	parts := strings.FieldsFunc(name, func(r rune) bool { return r == ' ' || r == '_' })
	for i, part := range parts {
		if strings.ToUpper(part) == "AND" || strings.ToUpper(part) == "OF" {
			parts[i] = strings.ToLower(part)
		} else {
			runes := []rune(part)
			parts[i] = string(unicode.ToUpper(runes[0])) + strings.ToLower(string(runes[1:]))
		}
	}

	return strings.Join(parts, " ")
}

func (e *ecdcExporter) getTags(stats []ecdcStat, i int) map[string]string {
	var tags map[string]string
	if e.Mp != nil && e.Mp.getLocation(stats[i].location) != nil {
		location := e.Mp.getLocation(stats[i].location)
		tags = map[string]string{"country": stats[i].location, "continent": stats[i].continent, "latitude": ftos(location.lat), "longitude": ftos(location.long)}
	} else {
		tags = map[string]string{"country": stats[i].location, "continent": stats[i].continent}
	}
	return tags
}

func getEcdcStat(url string) ([]ecdcStat, error) {
	client := http.Client{Timeout: 3 * time.Second}
	response, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	document, _ := goquery.NewDocumentFromReader(response.Body)
	rows := document.Find("table").Find("tbody").Find("tr")
	if rows.Size() == 0 {
		return nil, errors.New("Could not find table")
	}

	result := make([]ecdcStat, rows.Size()-1)

	rows.Each(func(i int, s *goquery.Selection) {
		if i < rows.Size()-1 {
			rowStart := s.Find("td").First()
			result[i] = ecdcStat{
				CovidStat{
					location: normalizeCountryName(rowStart.Next().Text()),
					infected: atoi(rowStart.Next().Next().Text()),
					deaths:   atoi(rowStart.Next().Next().Next().Text()),
				},
				rowStart.Text(),
			}
		}
	})
	return result, nil
}
