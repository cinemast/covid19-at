package exporter

import (
	"errors"
	"net/http"
	"strings"
	"unicode"

	"github.com/PuerkitoBio/goquery"
)

//EcdcExporter for parsing tables
type EcdcExporter struct {
	Url string
	Mp  *MetadataProvider
}

//EcdcStat for Cov19 infections and deaths
type EcdcStat struct {
	CovidStat
	continent string
}

//NewEcdcExporter creates a new exporter
func NewEcdcExporter(lp *MetadataProvider) *EcdcExporter {
	return &EcdcExporter{Url: "https://www.ecdc.europa.eu/en/geographical-distribution-2019-ncov-cases", Mp: lp}
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

//GetMetrics parses the ECDC table
func (e *EcdcExporter) GetMetrics() (Metrics, error) {
	stats, err := getEcdcStat(e.Url)
	if err != nil {
		return nil, err
	}
	result := make([]Metric, 0)
	for i := range stats {
		tags := e.getTags(stats, i)
		deaths := stats[i].deaths
		infected := stats[i].infected
		population := e.Mp.GetPopulation(stats[i].location)
		if deaths > 0 {
			result = append(result, Metric{Name: "cov19_world_death", Value: float64(deaths), Tags: &tags})
			if population > 0 {
				result = append(result, Metric{Name: "cov19_world_fatality_rate", Value: fatalityRate(infected, deaths), Tags: &tags})
			}
		}
		result = append(result, Metric{Name: "cov19_world_infected", Value: float64(infected), Tags: &tags})
		if population > 0 {
			result = append(result, Metric{Name: "cov19_world_infection_rate", Value: infectionRate(infected, population), Tags: &tags})
			result = append(result, Metric{Name: "cov19_world_infected_per_100k", Value: infection100k(infected, population), Tags: &tags})
		}
	}
	return result, nil
}

func (e *EcdcExporter) getTags(stats []EcdcStat, i int) map[string]string {
	var tags map[string]string
	if e.Mp != nil && e.Mp.GetLocation(stats[i].location) != nil {
		location := e.Mp.GetLocation(stats[i].location)
		tags = map[string]string{"country": stats[i].location, "continent": stats[i].continent, "latitude": ftos(location.lat), "longitude": ftos(location.long)}
	} else {
		tags = map[string]string{"country": stats[i].location, "continent": stats[i].continent}
	}
	return tags
}

func getEcdcStat(url string) ([]EcdcStat, error) {
	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	document, _ := goquery.NewDocumentFromReader(response.Body)
	rows := document.Find("table").Find("tbody").Find("tr")
	if rows.Size() == 0 {
		return nil, errors.New("Could not find table")
	}

	result := make([]EcdcStat, rows.Size()-1)

	rows.Each(func(i int, s *goquery.Selection) {
		if i < rows.Size()-1 {
			rowStart := s.Find("td").First()
			result[i] = EcdcStat{
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
