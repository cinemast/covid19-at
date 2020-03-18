package main

import (
	"errors"
	"net/http"
	"regexp"

	"github.com/PuerkitoBio/goquery"
)

//MinistryExporter for parsing tables from https://www.sozialministerium.at/Informationen-zum-Coronavirus/Neuartiges-Coronavirus-(2019-nCov).html
type MinistryExporter struct {
	url string
	lp  *LocationProvider
}

//NewMinistryExporter creates a new MinistryExporter
func NewMinistryExporter(lp *LocationProvider) *MinistryExporter {
	return &MinistryExporter{url: "https://www.sozialministerium.at/Informationen-zum-Coronavirus/Neuartiges-Coronavirus-(2019-nCov).html", lp: lp}
}

//GetMetrics returns total stats and province details
func (e *MinistryExporter) GetMetrics() (Metrics, error) {
	response, err := http.Get(e.url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	document, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		return nil, err
	}

	summary, err1 := e.GetTotalStats(document)
	details, err2 := e.GetProvinceStats(document)

	if err1 != nil && err2 != nil {
		err = errors.New(err1.Error() + " " + err2.Error())
	} else if err1 != nil {
		err = err1
	} else if err2 != nil {
		err = err2
	}

	return append(summary, details...), err
}

func (e *MinistryExporter) getTags(province string) *map[string]string {
	if e.lp != nil && e.lp.GetLocation(province) != nil {
		location := e.lp.GetLocation(province)
		return &map[string]string{"country": "Austria", "province": province, "latitude": ftos(location.lat), "longitude": ftos(location.long)}
	}
	return &map[string]string{"country": "Austria", "province": province}
}

//GetProvinceStats exports metrics per "Bundesland"
func (e *MinistryExporter) GetProvinceStats(document *goquery.Document) (Metrics, error) {
	summary, err := document.Find(".infobox").Html()
	if err != nil {
		return nil, err
	}
	result := make([]Metric, 0)
	summaryMatch := regexp.MustCompile(`Bestätigte Fälle.*`).FindAllString(summary, 1)
	if len(summaryMatch) == 0 {
		return nil, errors.New(`Could not find "Bestätigte Fälle"`)
	}

	re := regexp.MustCompile(`(?P<location>\S+) \((?P<number>\d+)\)`)
	matches := re.FindAllStringSubmatch(summaryMatch[0], -1)

	for _, match := range matches {
		metric := Metric{Name: "cov19_detail", Value: atoi(match[2]), Tags: e.getTags(match[1])}
		result = append(result, metric)
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
				metric := Metric{Name: "cov19_detail_dead", Value: atoi(match[valueIndex]), Tags: e.getTags(match[provinceIndex])}
				result = append(result, metric)
			}
		}
	}
	return result, nil
}

//GetTotalStats gets sumamrized stats and number of tests
func (e *MinistryExporter) GetTotalStats(document *goquery.Document) (Metrics, error) {
	result := make([]Metric, 0)
	summary, err := document.Find(".abstract").First().Html()

	if err != nil {
		return nil, err
	}

	confirmedMatch := regexp.MustCompile(`Fälle: [^0-9]*([0-9\.]+)`).FindStringSubmatch(summary)
	if len(confirmedMatch) >= 2 {
		result = append(result, Metric{Name: "cov19_confirmed", Value: atoi(confirmedMatch[1])})
	}

	testsMatch := regexp.MustCompile(`Testungen: [^0-9]*(?P<number>[0-9\.]+)`).FindAllStringSubmatch(summary, -1)
	if len(testsMatch) >= 1 && len(testsMatch[0]) >= 2 {
		result = append(result, Metric{Name: "cov19_tests", Value: atoi(testsMatch[0][1])})
	}

	healedMatch := regexp.MustCompile(`Genesene Personen: [^0-9]*([0-9\.]+)`).FindStringSubmatch(summary)
	if len(healedMatch) >= 2 {
		result = append(result, Metric{Name: "cov19_healed", Value: atoi(healedMatch[1])})
	}

	deadMatch := regexp.MustCompile(`Todesfälle: [^0-9]*([0-9\.]+)`).FindStringSubmatch(summary)
	if len(deadMatch) >= 2 {
		result = append(result, Metric{Name: "cov19_dead", Value: atoi(deadMatch[1])})
	}

	return result, nil
}
