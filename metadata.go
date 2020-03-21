package main

import (
	"encoding/csv"
	"log"
	"os"
	"regexp"
	"strings"
)

type metadataProvider struct {
	data map[string]metaData
}

type location struct {
	lat  float64
	long float64
}

type metaData struct {
	location   location
	country    string
	population uint64
}

func normalizeName(name string) string {
	space := regexp.MustCompile(`[^A-Za-z]+`)
	result := strings.ToUpper(space.ReplaceAllString(name, ""))
	return result
}

func newMetadataProvider() *metadataProvider {
	return newMetadataProviderWithFilename("metadata.csv")
}

func newMetadataProviderWithFilename(filename string) *metadataProvider {
	csvFile, err := os.Open(filename)
	if err != nil {
		log.Print(err)
		return nil
	}
	r := csv.NewReader(csvFile)
	records, err := r.ReadAll()
	if err != nil {
		log.Print(err)
		return nil
	}
	data := make(map[string]metaData, len(records))

	for _, row := range records {
		data[normalizeName(row[0])] = metaData{location{atof(row[2]), atof(row[3])}, row[0], atoi(row[1])}
	}
	return &metadataProvider{data: data}
}

func (l *metadataProvider) getMetadata(location string) *metaData {
	if l, ok := l.data[normalizeName(location)]; ok {
		return &l
	}
	return nil
}

//getLocation returns lat/long for a location name
func (l *metadataProvider) getLocation(location string) *location {
	if l, ok := l.data[normalizeName(location)]; ok {
		return &l.location
	}
	return nil
}

//getPopulation for a given location by name
func (l *metadataProvider) getPopulation(location string) uint64 {
	if l, ok := l.data[normalizeName(location)]; ok {
		return l.population
	}
	return 0
}
