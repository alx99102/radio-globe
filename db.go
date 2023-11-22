package main

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func initDB() *sql.DB {
	db, err := sql.Open("sqlite3", "./radio.db")
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

	return db
}

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

func ingestDB(db *sql.DB, stations []Station) {
    stmt, err := db.Prepare(`INSERT INTO Radios (name, url, url_resolved, homepage, favicon, country, language, codec, geo_lat, geo_long) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
    if err != nil {
        log.Fatal(err)
    }
    defer stmt.Close()

    for _, station := range stations {
        _, err := stmt.Exec(station.Name, station.URL, station.URLResolved, station.Homepage, station.Favicon, station.Country, station.Language, station.Codec, station.GeoLat, station.GeoLong)
        if err != nil {
            log.Println("Failed to insert station:", station.Name, "Error:", err)
        }
    }
}
