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
	name  string
	count int
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

	tests, _ := strconv.Atoi(re2.FindStringSubmatch(summary)[1])
	confirmed, _ := strconv.Atoi(re.FindStringSubmatch(summary)[1])
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
		number, _ := strconv.Atoi(match[2])
		result[i] = Stat{match[1], number}
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

}

func main() {
	http.HandleFunc("/metrics", handleMetrics)
	http.ListenAndServe(":8282", nil)
}
