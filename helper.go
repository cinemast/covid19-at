package main

import (
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func atoi(s string) uint64 {
	s = strings.ReplaceAll(s, ".", "")
	s = strings.ReplaceAll(s, ",", "")
	result, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0
	}
	return result
}

func atoif(s string) float64 {
	return float64(atoi(s))
}

func atof(s string) float64 {
	result, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return result
}

func ftos(f float64) string {
	return strconv.FormatFloat(f, 'f', 6, 64)
}

func fatalityRate(infections uint64, deaths uint64) float64 {
	return float64(deaths) / float64(infections)
}

func infectionRate(infections uint64, population uint64) float64 {
	return float64(infections) / float64(population)
}

func infection100k(infections uint64, population uint64) float64 {
	return infectionRate(infections, population) * float64(100000)
}

func readArrayFromGet(url string) (string, error) {
	client := http.Client{Timeout: 5 * time.Second}
	response, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	json, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	jsonString := string(json)
	arrayBegin := strings.Index(jsonString, "[")
	if arrayBegin == -1 {
		return "", errors.New("Could not find beginning of array")
	}

	arrayEnd := strings.LastIndex(jsonString, "]")
	if arrayEnd == -1 {
		return "", errors.New("Could not find end of array")
	}

	return jsonString[arrayBegin : arrayEnd+1], nil
}
