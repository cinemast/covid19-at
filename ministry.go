package main

import(
	"net/http"
	"errors"
	"regexp"
	"github.com/PuerkitoBio/goquery"
)

//MinistryExporter for parsing tables from https://www.sozialministerium.at/Informationen-zum-Coronavirus/Neuartiges-Coronavirus-(2019-nCov).html
type MinistryExporter struct {
	url string
}

//NewMinistryExporter creates a new MinistryExporter
func NewMinistryExporter() *MinistryExporter {
	return &MinistryExporter{url: "https://www.sozialministerium.at/Informationen-zum-Coronavirus/Neuartiges-Coronavirus-(2019-nCov).html"}
}

//GetMetrics returns total stats and province details
func (e *MinistryExporter) GetMetrics() ([]Metric, error) {

	response, err := http.Get(e.url)
	if err != nil  {
		return nil, err
	}
	defer response.Body.Close()

	document, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		return nil, err
	}

	summary, err1 := GetTotalStats(document)
	details, err2 := GetProvinceStats(document)


	if err1 != nil && err2 != nil {
		err = errors.New(err1.Error() + " " + err2.Error())
	} else if err1 != nil {
		err = err1
	} else if err2 != nil {
		err = err2
	}

	return append(summary,details...), err
}

//GetProvinceStats exports metrics per "Bundesland"
func GetProvinceStats(document *goquery.Document) ([]Metric, error) {
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
		tags := map[string]string{"country": "Austria", "province": match[1]}
		metric := Metric{Name: "cov19_detail", Value: atoi(match[2]), Tags: &tags}
		result = append(result, metric)
	}

	deathMatch := regexp.MustCompile(`Todesfälle.*`).FindAllString(summary, 1)
	if len(deathMatch) > 0 {
		matches := regexp.MustCompile(`(?P<number>\d+) \((?P<location>\S+)\)`).FindAllStringSubmatch(deathMatch[0], -1)
		for _, match := range matches {
			if len(match) > 2 {
				tags := map[string]string{"country": "Austria", "province": match[2]}
				metric := Metric{Name: "cov19_detail_dead", Value: atoi(match[1]), Tags: &tags}
				result = append(result, metric)
			}
		}
	}
	return result, nil
}


//GetTotalStats gets sumamrized stats and number of tests
func GetTotalStats(document *goquery.Document) ([]Metric, error) {
	result := make([]Metric, 0)
	summary, err := document.Find(".abstract").First().Html()

	if err != nil {
		return nil, err
	}

	confirmedMatch := regexp.MustCompile(`Fälle: [^0-9]*([0-9]+)`).FindStringSubmatch(summary)
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
	return result, nil
}
