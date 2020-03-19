package main

import (
	"errors"
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"regexp"
	"strings"
)

//MinistryExporter for parsing tables from https://www.sozialministerium.at/Informationen-zum-Coronavirus/Neuartiges-Coronavirus-(2019-nCov).html
type MinistryExporter struct {
	url string
	lp  *MetadataProvider
}

//NewMinistryExporter creates a new MinistryExporter
func NewMinistryExporter(lp *MetadataProvider) *MinistryExporter {
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
	provinceStats, err2 := e.getProvinceStats(document)

	if err1 != nil && err2 != nil {
		err = errors.New(err1.Error() + " " + err2.Error())
	} else if err1 != nil {
		err = err1
	} else if err2 != nil {
		err = err2
	}

	provinceMetrics := make(Metrics, 0)
	for _, s := range provinceStats {
		tags := e.getTags(s.location)
		provinceMetrics = append(provinceMetrics, Metric{Name: "cov19_detail", Value: float64(s.infected), Tags: tags})
		population := e.lp.GetPopulation(s.location)

		if population > 0 {
			provinceMetrics = append(provinceMetrics, Metric{Name: "cov19_detail_infection_rate", Value: infectionRate(s.infected, population), Tags: tags})
			provinceMetrics = append(provinceMetrics, Metric{Name: "cov19_detail_infected_per_100k", Value: infection100k(s.infected, population), Tags: tags})
		}
		if s.deaths > 0 {
			provinceMetrics = append(provinceMetrics, Metric{Name: "cov19_detail_dead", Value: float64(s.deaths), Tags: tags})
			if population > 0 {
				provinceMetrics = append(provinceMetrics, Metric{Name: "cov19_detail_fatality_rate", Value: fatalityRate(s.infected, s.deaths), Tags: tags})
			}
		}
	}

	return append(summary, provinceMetrics...), err
}

func (e *MinistryExporter) getTags(province string) *map[string]string {
	if e.lp != nil && e.lp.GetLocation(province) != nil {
		location := e.lp.GetLocation(province)
		return &map[string]string{"country": "Austria", "province": province, "latitude": ftos(location.lat), "longitude": ftos(location.long)}
	}
	return &map[string]string{"country": "Austria", "province": province}
}

//GetProvinceStats exports metrics per "Bundesland"
func (e *MinistryExporter) getProvinceStats(document *goquery.Document) (map[string]CovidStat, error) {
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
func (e *MinistryExporter) GetTotalStats(document *goquery.Document) (Metrics, error) {
	result := make([]Metric, 0)
	summary, err := document.Find(".abstract").First().Html()

	if err != nil {
		return nil, err
	}

	confirmedMatch := regexp.MustCompile(`Fälle:[^0-9]*([0-9\.]+)`).FindStringSubmatch(summary)
	if len(confirmedMatch) >= 2 {
		result = append(result, Metric{Name: "cov19_confirmed", Value: atoif(confirmedMatch[1])})
	}

	testsMatch := regexp.MustCompile(`Testungen:[^0-9]*(?P<number>[0-9\.]+)`).FindAllStringSubmatch(summary, -1)
	if len(testsMatch) >= 1 && len(testsMatch[0]) >= 2 {
		result = append(result, Metric{Name: "cov19_tests", Value: atoif(testsMatch[0][1])})
	}

	healedMatch := regexp.MustCompile(`Genesene Personen:[^0-9]*([0-9\.]+)`).FindStringSubmatch(summary)
	if len(healedMatch) >= 2 {
		result = append(result, Metric{Name: "cov19_healed", Value: atoif(healedMatch[1])})
	}

	deadMatch := regexp.MustCompile(`Todesfälle:[^0-9]*([0-9\.]+)`).FindStringSubmatch(summary)
	if len(deadMatch) >= 2 {
		result = append(result, Metric{Name: "cov19_dead", Value: atoif(deadMatch[1])})
	}

	return result, nil
}

//Angefragte Adresse:https://info.gesundheitsministerium.at/api/tsdb/query
//Anfragemethode:POST
//Externe Adresse:[2a00:1610:0:50::1bb:4a]:443
//Status-Code:
//200
//Version:HTTP/1.1
//Referrer Policy:no-referrer-when-downgrade
//curl 'https://info.gesundheitsministerium.at/api/tsdb/query' -H 'User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:74.0) Gecko/20100101 Firefox/74.0' -H 'Accept: application/json, text/plain, */*' -H 'Accept-Language: de,en-US;q=0.7,en;q=0.3' --compressed -H 'X-Grafana-Org-Id: 1' -H 'Content-Type: application/json;charset=utf-8' -H 'X-TS-AJAX-Request: true' -H 'Origin: https://info.gesundheitsministerium.at' -H 'DNT: 1' -H 'Connection: keep-alive' -H 'Referer: https://info.gesundheitsministerium.at/d/87BA8__Wk1/covid-19-offentlich-zugangliche-information?orgId=1&refresh=5m' -H 'Cookie: BIGipServerpool_radw34=335816970.20480.0000; TS01ad2292=013fa7d1cc3dc0776143d03262ddf9eb60deb8e777c2652a1f3598ee675e56f9a76d0804033acddc7b642bcfbdc81544cb995ea13d' -H 'Pragma: no-cache' -H 'Cache-Control: no-cache' --data '{"from":"1582498800000","to":"1584559998509","queries":[{"refId":"A","intervalMs":3600000,"maxDataPoints":512,"datasourceId":3,"rawSql":"select Bezirk, count(*) as Anzahl from GRAFANA_Covid_Faelle group by Bezirk order by Bezirk","format":"table"}]}'
//{"from":"1582498800000","to":"1584559998509","queries":[{"refId":"A","intervalMs":3600000,"maxDataPoints":512,"datasourceId":3,"rawSql":"select Bezirk, count(*) as Anzahl from GRAFANA_Covid_Faelle group by Bezirk order by Bezirk","format":"table"}]}
