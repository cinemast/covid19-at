package main

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

type mapsResponse struct {
	Candidates []struct {
		Geometry struct {
			Location struct {
				Lat float64
				Lng float64
			}
		}
	}
}

type mapLocation struct {
	latitude  float64
	longitude float64
}

func getLocation(location string, key string) (*mapLocation, error) {
	response, err := http.Get(fmt.Sprintf("https://maps.googleapis.com/maps/api/place/findplacefromtext/json?input=%s&inputtype=textquery&&fields=geometry&key=%s", url.QueryEscape(location), key))
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	bytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	result := mapsResponse{}
	err = json.Unmarshal(bytes, &result)
	if err != nil {
		return nil, err
	}

	if len(result.Candidates) == 0 {
		return nil, errors.New("Not enough candidates for " + location)
	}

	return &mapLocation{latitude: result.Candidates[0].Geometry.Location.Lat, longitude: result.Candidates[0].Geometry.Location.Lng}, nil
}

func ftos(f float64) string {
	return strconv.FormatFloat(f, 'f', 6, 64)
}

func main() {

	args := os.Args[1:]
	if len(args) < 1 {
		panic("Commandline argument for google maps apikey required")
	}

	apiKey := args[0]

	csvFile, _ := os.Open("bezirke.csv")
	r := csv.NewReader(csvFile)
	records, _ := r.ReadAll()

	for _, r := range records {
		location := r[0]
		loc, err := getLocation(location, apiKey)
		if err != nil {
			fmt.Println(err.Error())
		}
		fmt.Println(fmt.Sprintf("%s,%s,%f,%f", location, r[1], loc.latitude, loc.longitude))
	}

}
