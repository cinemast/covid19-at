package main

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

//TotalStat for Cov19 in Austria
type TotalStat struct {
	tests     int
	confirmed int
	healed    int
}

//ProvinceStat for Cov19
type ProvinceStat struct {
	name     string
	infected int
	death    int
}

//WorldStat for Cov19 infections and deaths
type WorldStat struct {
	continent string
	country   string
	infected  int
	deaths    int
}

var ecdcURL = "https://www.ecdc.europa.eu/en/geographical-distribution-2019-ncov-cases"
var healthministryURL = "https://www.sozialministerium.at/Informationen-zum-Coronavirus/Neuartiges-Coronavirus-(2019-nCov).html"

func parseStats(reader io.Reader) TotalStat {
	document, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		return TotalStat{0, 0, 0}
	}

	summary, err := document.Find(".abstract").First().Html()

	confirmedMatch := regexp.MustCompile(`Fälle: [^0-9]*([0-9]+)`).FindStringSubmatch(summary)
	healed := 0
	confirmed := 0
	tests := 0
	if len(confirmedMatch) >= 2 {
		confirmed = atoi(confirmedMatch[1])
	}

	testsMatch := regexp.MustCompile(`Testungen: [^0-9]*(?P<number>[0-9\.]+)`).FindAllStringSubmatch(summary, -1)
	if len(testsMatch) >= 1 && len(testsMatch[0]) >= 2 {
		tests = atoi(testsMatch[0][1])
	}

	healedMatch := regexp.MustCompile(`Genesene Personen: [^0-9]*([0-9\.]+)`).FindStringSubmatch(summary)
	if len(healedMatch) >= 2 {
		healed = atoi(healedMatch[1])
	}
	return TotalStat{tests: tests, confirmed: confirmed, healed: healed}
}

func getStats(c chan TotalStat) {
	response, err := http.Get(healthministryURL)
	if err != nil {
		c <- TotalStat{0, 0, 0}
		return
	}
	defer response.Body.Close()
	c <- parseStats(response.Body)
}

func mapToSlice(m map[string]ProvinceStat) []ProvinceStat {
	r := make([]ProvinceStat, len(m))
	i := 0
	for _, v := range m {
		r[i] = v
		i++
	}
	return r
}

func parseProvinceStats(r io.Reader) []ProvinceStat {
	document, _ := goquery.NewDocumentFromReader(r)
	summary, _ := document.Find(".infobox").Html()
	result := make(map[string]ProvinceStat)
	summaryMatch := regexp.MustCompile(`Bestätigte Fälle.*`).FindAllString(summary, 1)
	if len(summaryMatch) == 0 {
		return []ProvinceStat{}
	}

	re := regexp.MustCompile(`(?P<location>\S+) \((?P<number>\d+)\)`)
	matches := re.FindAllStringSubmatch(summaryMatch[0], -1)

	for _, match := range matches {
		stat := ProvinceStat{name: match[1], infected: atoi(match[2]), death: 0}
		result[stat.name] = stat
	}

	deathMatch := regexp.MustCompile(`Todesfälle.*`).FindAllString(summary, 1)
	if len(deathMatch) > 0 {
		matches := regexp.MustCompile(`(?P<number>\d+) \((?P<location>\S+)\)`).FindAllStringSubmatch(deathMatch[0], -1)
		for _, match := range matches {
			if len(match) > 2 {
				name := match[2]
				death := atoi(match[1])
				if val, ok := result[name]; ok {
					val.death = death
					result[name] = val
				} else {
					result[name] = ProvinceStat{name: name, infected: 0, death: death}
				}
			}
		}
	}
	return mapToSlice(result)
}

func getDetails(c chan []ProvinceStat) {
	response, err := http.Get(healthministryURL)
	if err != nil {
		c <- []ProvinceStat{}
		return
	}
	defer response.Body.Close()
	c <- parseProvinceStats(response.Body)
}

func parseWorldStats(r io.Reader) []WorldStat {
	document, _ := goquery.NewDocumentFromReader(r)
	rows := document.Find("table").Find("tbody").Find("tr")
	if rows.Size() == 0 {
		return make([]WorldStat, 0)
	}

	result := make([]WorldStat, rows.Size()-1)

	rows.Each(func(i int, s *goquery.Selection) {
		if i < rows.Size()-1 {
			rowStart := s.Find("td").First()
			result[i] = WorldStat{
				continent: rowStart.Text(),
				country:   rowStart.Next().Text(),
				infected:  atoi(rowStart.Next().Next().Text()),
				deaths:    atoi(rowStart.Next().Next().Next().Text()),
			}
		}
	})
	return result
}

func getWorldStats(c chan []WorldStat) {
	response, err := http.Get(ecdcURL)
	if err != nil {
		c <- make([]WorldStat, 0)
		return
	}
	defer response.Body.Close()
	c <- parseWorldStats(response.Body)
}

func atoi(s string) int {
	s = strings.ReplaceAll(s, ".", "")
	s = strings.ReplaceAll(s, ",", "")
	result, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return result
}

func getStatsAsync() (TotalStat, []ProvinceStat, []WorldStat) {
	statsChannel := make(chan TotalStat)
	provinceChannel := make(chan []ProvinceStat)
	worldChannel := make(chan []WorldStat)
	go getStats(statsChannel)
	go getDetails(provinceChannel)
	go getWorldStats(worldChannel)
	return <-statsChannel, <-provinceChannel, <-worldChannel
}

func handleMetrics(w http.ResponseWriter, r *http.Request) {

	stats, details, worldStats := getStatsAsync()

	fmt.Fprintf(w, "cov19_tests %d\n", stats.tests)
	fmt.Fprintf(w, "cov19_confirmed %d\n", stats.confirmed)
	fmt.Fprintf(w, "cov19_healed %d\n", stats.healed)

	for _, detail := range details {
		if detail.infected > 0 {
			fmt.Fprintf(w, "cov19_detail{country=\"Austria\",province=\"%s\"} %d\n", detail.name, detail.infected)
		}
		if detail.death > 0 {
			fmt.Fprintf(w, "cov19_detail_dead{country=\"Austria\",province=\"%s\"} %d\n", detail.name, detail.death)
		}
	}

	for _, s := range worldStats {
		fmt.Fprintf(w, "cov19_world_death{continent=\"%s\",country=\"%s\"} %d\n", s.continent, s.country, s.deaths)
		fmt.Fprintf(w, "cov19_world_infected{continent=\"%s\",country=\"%s\"} %d\n", s.continent, s.country, s.infected)
	}
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	summary, details, world := getStatsAsync()
	failures := 0

	errorResponse := ""

	if summary.confirmed == 0 {
		failures++
		errorResponse = errorResponse + "Summary confirmed are failing\n"
	}

	if summary.healed == 0 {
		failures++
		errorResponse = errorResponse + "Summary healed are failing\n"
	}

	if summary.tests == 0 {
		failures++
		errorResponse = errorResponse + "Summary tests are failing\n"
	}

	if len(details) == 0 || (details[0].infected == 0 && details[0].death == 0) {
		failures++
		errorResponse = errorResponse + "Details Austria are failing\n"
	}

	if len(world) == 0 {
		failures++
		errorResponse = errorResponse + "World stats are failing\n"
	}

	if failures > 0 {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, `<html><body><img width="500" src="https://spiessknafl.at/fine.jpg"/><pre>%s</pre></body></html>`, errorResponse)
	} else {
		fmt.Fprintf(w, `<html><body><img width="500" src="https://spiessknafl.at/helth.png"/></body></html>`)
	}
}

func main() {
	http.HandleFunc("/metrics", handleMetrics)
	http.HandleFunc("/health", handleHealth)
	http.ListenAndServe(":8282", nil)
}
