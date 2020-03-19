package main

import (
	"encoding/csv"
	"os"
	"regexp"
	"strings"
)

//MetadataProvider to lookup locations based on names
type MetadataProvider struct {
	data map[string]metaData
}

//Location describes Lat/Long values
type Location struct {
	lat  float64
	long float64
}

type metaData struct {
	location   Location
	country    string
	population uint64
}

func normalizeName(name string) string {
	space := regexp.MustCompile(`[^A-Za-z]+`)
	result := strings.ToUpper(space.ReplaceAllString(name, ""))
	return result
}

func NewMetadataProvider() *MetadataProvider {
	return NewMetadataProviderWithFilename("metadata.csv")
}

//NewMetadataProvider creates a new locationProvider
func NewMetadataProviderWithFilename(filename string) *MetadataProvider {
	csvFile, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	r := csv.NewReader(csvFile)
	records, err := r.ReadAll()
	if err != nil {
		panic(err)
	}
	data := make(map[string]metaData, len(records))

	for _, row := range records {
		data[normalizeName(row[0])] = metaData{Location{atof(row[2]), atof(row[3])}, row[0], atoi(row[1])}
	}
	return &MetadataProvider{data: data}
}

func (l *MetadataProvider) GetMetadata(location string) *metaData {
	if l, ok := l.data[normalizeName(location)]; ok {
		return &l
	}
	return nil
}

//GetLocation returns lat/long for a location name
func (l *MetadataProvider) GetLocation(location string) *Location {
	if l, ok := l.data[normalizeName(location)]; ok {
		return &l.location
	}
	return nil
}

//GetPopulation for a given location by name
func (l *MetadataProvider) GetPopulation(location string) uint64 {
	if l, ok := l.data[normalizeName(location)]; ok {
		return l.population
	}
	return 0
}
