package main

import (
	"encoding/csv"
	"os"
)

//LocationProvider to lookup locations based on names
type LocationProvider struct {
	locations map[string]Location
}

//Location describes Lat/Long values
type Location struct {
	lat  float64
	long float64
}

//NewLocationProvider creates a new locationProvider
func NewLocationProvider() *LocationProvider {
	csvFile, _ := os.Open("locations.csv")
	r := csv.NewReader(csvFile)
	records, _ := r.ReadAll()
	locations := make(map[string]Location, len(records))

	for _, row := range records {
		locations[row[2]] = Location{atof(row[0]), atof(row[1])}
	}
	return &LocationProvider{locations: locations}
}

//GetLocation returns lat/long for a location name
func (l *LocationProvider) GetLocation(location string) *Location {
	if l, ok := l.locations[location]; ok {
		return &l
	}
	return nil
}
