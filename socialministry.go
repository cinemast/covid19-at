package main

import (
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type socialMinistryExporter struct {
	url string
	mp  *metadataProvider
}

func newSocialMinistryExporter(lp *metadataProvider) *socialMinistryExporter {
	return &socialMinistryExporter{url: "https://www.sozialministerium.at/Informationen-zum-Coronavirus/Neuartiges-Coronavirus-(2019-nCov).html", mp: lp}
}

func (e *socialMinistryExporter) Health() []error {
	errors := make([]error, 0)

	ministryStats, err := e.GetMetrics()
	if err != nil {
		errors = append(errors, err)
	}

	if len(ministryStats) < 10 {
		errors = append(errors, fmt.Errorf("Missing ministry stats"))
	}

	err = ministryStats.checkMetric("cov19_healed", "", func(x float64) bool { return x > 5 })
	if err != nil {
		errors = append(errors, err)
	}
	return errors
}

//GetMetrics returns total stats and province details
func (e *socialMinistryExporter) GetMetrics() (metrics, error) {
	client := http.Client{Timeout: 3 * time.Second}
	response, err := client.Get(e.url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	document, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		return nil, err
	}

	summary, err1 := e.GetTotalStats(document)
	provinceStats, err2 := e.getProvinceStats(document)

	if err1 != nil && err2 != nil {
		err = errors.New(err1.Error() + " " + err2.Error())
	} else if err1 != nil {
		err = err1
	} else if err2 != nil {
		err = err2
	}

	provinceMetrics := make(metrics, 0)
	for _, s := range provinceStats {
		tags := e.getTags(s.location)
		population := e.mp.getPopulation(s.location)
		if s.deaths > 0 {
			provinceMetrics = append(provinceMetrics, metric{Name: "cov19_detail_dead", Value: float64(s.deaths), Tags: tags})
			if population > 0 {
				provinceMetrics = append(provinceMetrics, metric{Name: "cov19_detail_fatality_rate", Value: fatalityRate(s.infected, s.deaths), Tags: tags})
			}
		}
	}

	return append(summary, provinceMetrics...), err
}

func (e *socialMinistryExporter) getTags(province string) *map[string]string {
	if e.mp != nil && e.mp.getLocation(province) != nil {
		location := e.mp.getLocation(province)
		return &map[string]string{"country": "Austria", "province": province, "latitude": ftos(location.lat), "longitude": ftos(location.long)}
	}
	return &map[string]string{"country": "Austria", "province": province}
}

//GetProvinceStats exports metrics per "Bundesland"
func (e *socialMinistryExporter) getProvinceStats(document *goquery.Document) (map[string]CovidStat, error) {
	result := make(map[string]CovidStat)
	summary, err := document.Find(".infobox").Html()
	if err != nil {
		return nil, err
	}
	summaryMatch := regexp.MustCompile(`Bestätigte Fälle.*`).FindAllString(summary, 1)
	if len(summaryMatch) == 0 {
		return nil, errors.New(`Could not find "Bestätigte Fälle"`)
	}

	re := regexp.MustCompile(`(?P<location>\S+) \((?P<number>\d+)\)`)
	matches := re.FindAllStringSubmatch(summaryMatch[0], -1)

	for _, match := range matches {
		infected := atoi(match[2])
		province := strings.TrimSpace(strings.ReplaceAll(match[1], ",", ""))
		result[province] = CovidStat{province, infected, 0}

	}

	deathMatch := regexp.MustCompile(`Todesfälle.*`).FindAllString(summary, 1)
	if len(deathMatch) > 0 {
		matches := regexp.MustCompile(`(?P<number>\d+) \((?P<location>\S+)\)`).FindAllStringSubmatch(deathMatch[0], -1)
		provinceIndex := 2
		valueIndex := 1
		if len(matches) == 0 {
			matches = regexp.MustCompile(`(?P<location>\S+) \((?P<number>\d+)\)`).FindAllStringSubmatch(deathMatch[0], -1)
			provinceIndex = 1
			valueIndex = 2
		}
		for _, match := range matches {
			if len(match) > 2 {
				location := strings.TrimSpace(strings.ReplaceAll(match[provinceIndex], ",", ""))
				stat := result[location]
				stat.deaths = atoi(match[valueIndex])
				result[location] = stat
			}
		}
	}
	return result, nil
}

//GetTotalStats gets sumamrized stats and number of tests
func (e *socialMinistryExporter) GetTotalStats(document *goquery.Document) (metrics, error) {
	result := make([]metric, 0)
	summary, err := document.Find(".abstract").First().Html()

	if err != nil {
		return nil, err
	}

	confirmedMatch := regexp.MustCompile(`Fälle:[^0-9]*([0-9\.]+)`).FindStringSubmatch(summary)
	if len(confirmedMatch) >= 2 {
		result = append(result, metric{Name: "cov19_confirmed", Value: atoif(confirmedMatch[1])})
	}

	testsMatch := regexp.MustCompile(`Testungen:[^0-9]*(?P<number>[0-9\.]+)`).FindAllStringSubmatch(summary, -1)
	if len(testsMatch) >= 1 && len(testsMatch[0]) >= 2 {
		result = append(result, metric{Name: "cov19_tests", Value: atoif(testsMatch[0][1])})
	}

	healedMatch := regexp.MustCompile(`Genesene Personen:[^0-9]*([0-9\.]+)`).FindStringSubmatch(summary)
	if len(healedMatch) >= 2 {
		result = append(result, metric{Name: "cov19_healed", Value: atoif(healedMatch[1])})
	}

	deadMatch := regexp.MustCompile(`Todesfälle:[^0-9]*([0-9\.]+)`).FindStringSubmatch(summary)
	if len(deadMatch) >= 2 {
		result = append(result, metric{Name: "cov19_dead", Value: atoif(deadMatch[1])})
	}

	return result, nil
}
