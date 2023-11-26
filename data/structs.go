package data

type Station struct {
	Name string `json:"name"`
	URL string `json:"url"`
	URLResolved string `json:"url_resolved"`
	Homepage string `json:"homepage"`
	Favicon string `json:"favicon"`
	Country string `json:"country"`
	Language string `json:"language"`
	Codec string `json:"codec"`
	GeoLat float64 `json:"geo_lat"`
	GeoLong float64 `json:"geo_long"`
}

type RadioContent struct {
	Station Station
	ContentType string
}

type MainContent struct {
	Radios []RadioContent
	GoogleMapsApiKey string
	Location string
}

type FeatureCollection struct {
	Features []Feature `json:"features"`
}

type Feature struct {
	Properties Properties `json:"properties"`
}

type Properties struct {
	Name         string `json:"name"`
	Country      string `json:"country"`
	Region       string `json:"region"`
	State        string `json:"state"`
	City         string `json:"city"`
	Formatted    string `json:"formatted"`
	AddressLine1 string `json:"address_line1"`
	AddressLine2 string `json:"address_line2"`
}

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