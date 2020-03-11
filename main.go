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
	name  string
	count int
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

func parseProvinceStats(r io.Reader) []ProvinceStat {
	document, _ := goquery.NewDocumentFromReader(r)
	summary, _ := document.Find(".infobox").Html()

	summaryMatch := regexp.MustCompile(`Bestätigte Fälle.*`).FindAllString(summary, 1)
	if len(summaryMatch) == 0 {
		return make([]ProvinceStat, 0)
	}

	re := regexp.MustCompile(`(?P<location>\S+) \((?P<number>\d+)\)`)
	matches := re.FindAllStringSubmatch(summaryMatch[0], -1)

	result := make([]ProvinceStat, len(matches))
	for i, match := range matches {
		number := atoi(match[2])
		result[i] = ProvinceStat{match[1], number}
	}

	return result
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
		fmt.Fprintf(w, "cov19_detail{country=\"Austria\",province=\"%s\"} %d\n", detail.name, detail.count)
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

	if len(details) == 0 || details[0].count == 0 {
		failures++
		errorResponse = errorResponse + "Details Austria are failing\n"
	}

	if len(world) == 0 {
		failures++
		errorResponse = errorResponse + "World stats are failing\n"
	}

	if failures > 0 {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, errorResponse)
	} else {
		fmt.Fprintf(w, "Everything is fine :)\n")
	}
}

func main() {
	http.HandleFunc("/metrics", handleMetrics)
	http.HandleFunc("/health", handleHealth)
	http.ListenAndServe(":8282", nil)
}
