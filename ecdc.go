package main

import (
	"errors"
	"net/http"
	"strings"
	"unicode"

	"github.com/PuerkitoBio/goquery"
)

//EcdcExporter for parsing tables
type EcdcExporter struct {
	url string
	lp  *MetadataProvider
}

//EcdcStat for Cov19 infections and deaths
type EcdcStat struct {
	continent string
	country   string
	infected  uint64
	deaths    uint64
}

//NewEcdcExporter creates a new exporter
func NewEcdcExporter(lp *MetadataProvider) *EcdcExporter {
	return &EcdcExporter{url: "https://www.ecdc.europa.eu/en/geographical-distribution-2019-ncov-cases", lp: lp}
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
	stats, err := getEcdcStat(e.url)
	if err != nil {
		return nil, err
	}
	result := make([]Metric, 0)
	for i := range stats {
		var tags map[string]string
		if e.lp != nil && e.lp.GetLocation(stats[i].country) != nil {
			location := e.lp.GetLocation(stats[i].country)
			tags = map[string]string{"country": stats[i].country, "continent": stats[i].continent, "latitude": ftos(location.lat), "longitude": ftos(location.long)}
		} else {
			tags = map[string]string{"country": stats[i].country, "continent": stats[i].continent}
		}
		deaths := stats[i].deaths
		infected := stats[i].infected
		population := e.lp.GetPopulation(stats[i].country)
		if deaths > 0 {
			result = append(result, Metric{Name: "cov19_world_death", Value: deaths, Tags: &tags})
			if population > 0 {
				result = append(result, Metric{Name: "cov19_world_fatality_rate", Value: deaths / infected, Tags: &tags})
			}
		}
		result = append(result, Metric{Name: "cov19_world_infected", Value: infected, Tags: &tags})

		//TODO: fatality rate
		//percent infected

		if population > 0 {

		}
	}
	return result, nil
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
				continent: rowStart.Text(),
				country:   normalizeCountryName(rowStart.Next().Text()),
				infected:  atoi(rowStart.Next().Next().Text()),
				deaths:    atoi(rowStart.Next().Next().Next().Text()),
			}
		}
	})
	return result, nil
}
