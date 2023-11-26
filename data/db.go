package data

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func InitDB() {
	db, err := sql.Open("sqlite3", "./data/radio.db")
	if err != nil {
		log.Fatal(err)
	}

	// create table
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS Radios ( id INTEGER PRIMARY KEY, name TEXT, url TEXT, url_resolved TEXT, homepage TEXT, favicon TEXT, country TEXT, language TEXT, codec TEXT, geo_lat REAL, geo_long REAL);`)
	if err != nil {
		log.Fatal(err)
	}

	// create an R-Tree index for the geo_lat and geo_long columns
	_, err = db.Exec("CREATE VIRTUAL TABLE IF NOT EXISTS geo_index USING rtree(id, minLat, maxLat, minLon, maxLon);")
	if err != nil {
    	log.Fatal(err)
	}

	DB = db
}

func SearchByCoordinates(db *sql.DB, lat, long float64) []Station {
	rows, err := db.Query("SELECT * FROM Radios WHERE geo_lat BETWEEN ? AND ? AND geo_long BETWEEN ? AND ?", lat - 0.5, lat + 0.5, long - 0.5, long + 0.5)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var stations []Station
	var id int
	for rows.Next() {
		var station Station
		if err := rows.Scan(&id, &station.Name, &station.URL, &station.URLResolved, &station.Homepage, &station.Favicon, &station.Country, &station.Language, &station.Codec, &station.GeoLat, &station.GeoLong); err != nil {
			log.Fatal(err)
		}
		stations = append(stations, station)
	}

	return stations
}
	
var timeoutClient = &http.Client{Timeout: 10 * time.Second}

func GetContentType(URL string) (string, error) {
    // Make a request to the URL
    resp, err := timeoutClient.Get(URL)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

	// Check the content-type header
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		return "", fmt.Errorf("can't determine file type")
	}

	return contentType, nil 
}

func CloseDB() {
	rows, err := DB.Query("SELECT id FROM Radios")
	if err != nil {
		log.Fatal(err)
	}

	rows.Close()
	DB.Close()
}