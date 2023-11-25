package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

/*
{
  "items": [
    {
      "title": "Montréal, QC, Canada",
      "id": "here:cm:namedplace:21155524",
      "resultType": "locality",
      "localityType": "city",
      "address": {
        "label": "Montréal, QC, Canada",
        "countryCode": "CAN",
        "countryName": "Canada",
        "stateCode": "QC",
        "state": "Québec",
        "county": "Montréal",
        "city": "Montréal",
        "postalCode": "H2L"
      },
      "position": { "lat": 45.5124, "lng": -73.55469 },
      "mapView": {
        "west": -73.97661,
        "south": 45.41119,
        "east": -73.47194,
        "north": 45.70479
      },
      "scoring": {
        "queryScore": 1.0,
        "fieldScore": { "country": 1.0, "state": 1.0, "city": 1.0 }
      }
    }
  ]
}
*/
// Structs for parsing JSON response
type HereResponse struct {
	Items []Item `json:"items"`
}

type Item struct {
	Title string `json:"title"`
	Id string `json:"id"`
	ResultType string `json:"resultType"`
	LocalityType string `json:"localityType"`
	Address Address `json:"address"`
	Position Position `json:"position"`
	MapView MapView `json:"mapView"`
	Scoring Scoring `json:"scoring"`
}

type Address struct {
	Label string `json:"label"`
	CountryCode string `json:"countryCode"`
	CountryName string `json:"countryName"`
	StateCode string `json:"stateCode"`
	State string `json:"state"`
	County string `json:"county"`
	City string `json:"city"`
	PostalCode string `json:"postalCode"`
}

type Position struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type MapView struct {
	West float64 `json:"west"`
	South float64 `json:"south"`
	East float64 `json:"east"`
	North float64 `json:"north"`
}

type Scoring struct {
	QueryScore float64 `json:"queryScore"`
	FieldScore FieldScore `json:"fieldScore"`
}

type FieldScore struct {
	Country float64 `json:"country"`
	State float64 `json:"state"`
	City float64 `json:"city"`
}

// get geocoding API key from creds.json
func getHereAPIKey() string {
	var credentials Credentials
	data, err := ioutil.ReadFile("creds.json")
	if err != nil {
		fmt.Println(err)
		return ""
	}
	err = json.Unmarshal(data, &credentials)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	return credentials.HereApiKey
}


func geocode(address string) (float64, float64, error) {
	// get API key
	apiKey := getHereAPIKey()
	// build URL

	// URL encode address
	address = url.QueryEscape(address)
	
	url := "https://geocode.search.hereapi.com/v1/geocode?q=" + address + "&apiKey=" + apiKey
	
	// GET request
	res, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
		return 0, 0, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		fmt.Printf("API request failed with status code: %d\n", res.StatusCode)
		return 0, 0, fmt.Errorf("API request failed with status code: %d", res.StatusCode)
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return 0, 0, err
	}

	// parse JSON
	var hereResponse HereResponse
	err = json.Unmarshal(body, &hereResponse)
	if err != nil {
		fmt.Println(err)
		return 0, 0, err
	}
	// return lat, lng
	if len(hereResponse.Items) > 0 {
		return hereResponse.Items[0].Position.Lat, hereResponse.Items[0].Position.Lng, nil
	}
	//fix this return
	return 0, 0, fmt.Errorf("no results found")
}
