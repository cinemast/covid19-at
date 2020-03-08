package main

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"

	"github.com/PuerkitoBio/goquery"
)

//Stats for Cov19 in Austria
type Stats struct {
	tests     int
	confirmed int
}

type Stat struct {
	name   string
	count  int
	deaths int
}

type WorldStat struct {
	continent string
	country   string
	infected  int
	deaths    int
}

func getStats() Stats {
	response, err := http.Get("https://www.sozialministerium.at/Informationen-zum-Coronavirus/Neuartiges-Coronavirus-(2019-nCov).html")
	if err != nil {
		return Stats{0, 0}
	}
	defer response.Body.Close()

	// Create a goquery document from the HTTP response
	document, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		return Stats{0, 0}
	}

	summary, err := document.Find(".abstract").First().Html()
	re := regexp.MustCompile("Bestätigte Fälle: ([0-9]+)")
	re2 := regexp.MustCompile("Bisher durchgeführte Testungen: ([0-9]+)")

	tests := atoi(re2.FindStringSubmatch(summary)[1])
	confirmed := atoi(re.FindStringSubmatch(summary)[1])
	return Stats{tests: tests, confirmed: confirmed}
}

func getDetails() []Stat {
	response, err := http.Get("https://www.sozialministerium.at/Themen/Gesundheit/Uebertragbare-Krankheiten/Infektionskrankheiten-A-Z/Neuartiges-Coronavirus.html")
	if err != nil {
		fmt.Println("Error get request")
		return []Stat{}
	}
	defer response.Body.Close()

	document, err := goquery.NewDocumentFromReader(response.Body)
	summary, err := document.Find("#content").Html()
	re := regexp.MustCompile(`(?P<location>\S+) \((?P<number>\d+)\)`)
	matches := re.FindAllStringSubmatch(summary, -1)

	result := make([]Stat, len(matches))
	for i, match := range matches {
		number := atoi(match[2])
		result[i] = Stat{match[1], number, 0}
	}

	return result
}

func getWorldStats() []WorldStat {
	response, err := http.Get("https://www.ecdc.europa.eu/en/geographical-distribution-2019-ncov-cases")
	if err != nil {
		fmt.Println("Error get request")

	}
	document, err := goquery.NewDocumentFromReader(response.Body)
	table := document.Find("table").Find("tbody")
	if table == nil {
		fmt.Println("Error getting world stats")
	}

	rows := table.Find("tr")
	result := make([]WorldStat, rows.Size()-2)

	rows.Each(func(i int, s *goquery.Selection) {
		if i > 0 && i < rows.Size()-1 {
			rowStart := s.Find("td").First()
			result[i-1] = WorldStat{
				continent: rowStart.Text(),
				country:   rowStart.Next().Find("p").Text(),
				infected:  atoi(rowStart.Next().Next().Find("p").Text()),
				deaths:    atoi(rowStart.Next().Next().Next().Find("p").Text()),
			}
		}
	})
	return result
}

func atoi(s string) int {
	result, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return result
}

var provinces = [9]string{"Wien", "Niederösterreich", "Oberösterreich", "Salzburg", "Tirol", "Vorarlberg", "Steiermark", "Burgenland", "Kärnten"}

func isAustria(location string) bool {
	for _, province := range provinces {
		if province == location {
			return true
		}
	}
	return false
}

func handleMetrics(w http.ResponseWriter, r *http.Request) {
	stats := getStats()
	fmt.Println("Summary: ", stats)
	fmt.Fprintf(w, "cov19_tests %d\n", stats.tests)
	fmt.Fprintf(w, "cov19_confirmed %d\n", stats.confirmed)

	details := getDetails()
	fmt.Println("Details: ", details)
	for _, detail := range details {
		if isAustria(detail.name) {
			fmt.Fprintf(w, "cov19_detail{country=\"Austria\",province=\"%s\"} %d\n", detail.name, detail.count)
		} else {
			fmt.Fprintf(w, "cov19_detail{country=\"%s\"} %d\n", detail.name, detail.count)
		}
	}

	for _, s := range getWorldStats() {
		fmt.Fprintf(w, "cov19_world_death{continent=\"%s\",country=\"%s\"} %d\n", s.continent, s.country, s.deaths)
		fmt.Fprintf(w, "cov19_world_infected{continent=\"%s\",country=\"%s\"} %d\n", s.continent, s.country, s.infected)
	}
}

func main() {
	http.HandleFunc("/metrics", handleMetrics)
	http.ListenAndServe(":8282", nil)
}
