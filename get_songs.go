package main

import (
	"database/sql"
	"encoding/json"
	"encoding/xml"
	"github.com/patrickmn/go-cache"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"
)

type ITunesSearch struct {
	Origin string `json:"origin"`
	Data   struct {
		ResultCount int                  `json:"-"`
		Results     []ITunesResultSearch `json:"results"`
	} `json:"data"`
}

type ITunesResultSearch struct {
	TrackId         int     `json:"trackId"`
	TrackName       string  `json:"trackName"`
	ArtistName      string  `json:"artistName"`
	TrackTimeMillis int     `json:"trackTimeMillis"`
	CollectionName  string  `json:"collectionName"`
	ArtworkUrl30    string  `json:"artworkUrl30"`
	TrackPrice      float32 `json:"trackPrice"`
}

type SearchLyricResult struct {
	XMLName           xml.Name `xml:"ArrayOfSearchLyricResult"`
	Text              string   `xml:",chardata"`
	Xsd               string   `xml:"xsd,attr"`
	Xsi               string   `xml:"xsi,attr"`
	Xmlns             string   `xml:"xmlns,attr"`
	SearchLyricResult []struct {
		Text          string `xml:",chardata"`
		Nil           string `xml:"nil,attr"`
		TrackId       int    `xml:"TrackId"`
		LyricChecksum string `xml:"LyricChecksum"`
		LyricId       string `xml:"LyricId"`
		SongUrl       string `xml:"SongUrl"`
		ArtistUrl     string `xml:"ArtistUrl"`
		Artist        string `xml:"Artist"`
		Song          string `xml:"Song"`
		SongRank      string `xml:"SongRank"`
		TrackChecksum string `xml:"TrackChecksum"`
	} `xml:"SearchLyricResult"`
}

type Song struct {
	Id       int     `json:"id"`
	Name     string  `json:"name"`
	Artist   string  `json:"artist"`
	Duration int     `json:"duration"`
	Album    string  `json:"album"`
	Artwork  string  `json:"artwork"`
	Price    float32 `json:"price"`
	Origin   string  `json:"origin"`
}

func getAll(w http.ResponseWriter, r *http.Request) {

	db, err := setupDB()
	if err != nil {
		log.Printf("Error al configurar Base de datos %v", err)
	}

	name := r.URL.Query().Get("name")
	artist := r.URL.Query().Get("artist")
	album := r.URL.Query().Get("album")

	var results []Song

	dataCache := saveCache(name, artist, album, db)

	if dataCache != nil {
		results = dataCache
	} else {

		iTunes := getItunes(w, name, artist, album)
		chart := getChartLyrics(w, name, artist)

		results = append(results, saveITunes(iTunes, db)...)
		results = append(results, saveChart(chart, db)...)
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(results)
	if err != nil {
		log.Printf("Error al enviar respuesta: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func saveITunes(iTunes []ITunesResultSearch, db *sql.DB) []Song {
	var results []Song
	for i, obj := range iTunes {
		results = append(results, Song{
			Id:       obj.TrackId,
			Name:     obj.TrackName,
			Artist:   obj.ArtistName,
			Duration: obj.TrackTimeMillis,
			Album:    obj.CollectionName,
			Artwork:  obj.ArtworkUrl30,
			Price:    obj.TrackPrice,
			Origin:   "Itunes",
		})

		_, err := db.Exec("INSERT INTO songs(track_id, name, artist, duration, album, artwork, price, origin) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
			results[i].Id, results[i].Name, results[i].Artist, results[i].Duration, results[i].Album, results[i].Artwork, results[i].Price, results[i].Origin)
		if err != nil {
			log.Printf("Error al guardar en base de datos: %v", err)
		}
	}
	return results
}

func saveChart(chart SearchLyricResult, db *sql.DB) []Song {
	var results []Song
	for i, obj := range chart.SearchLyricResult {
		results = append(results, Song{
			Id:     obj.TrackId,
			Name:   obj.Song,
			Artist: obj.Artist,
			Origin: "ChartLyrics",
		})

		_, err := db.Exec("INSERT INTO songs(track_id, name, artist, duration, album, artwork, price, origin) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
			results[i].Id, results[i].Name, results[i].Artist, 0, "", "", 0, results[i].Origin)
		if err != nil {
			log.Printf("Error al guardar en base de datos: %v", err)
		}
	}
	return results
}

func getItunes(w http.ResponseWriter, name string, artist string, album string) []ITunesResultSearch {

	apiUrl := "http://itunes.apple.com/search"

	params := url.Values{}
	params.Add("term", name+" "+artist+" "+album)
	params.Add("entity", "musicTrack")

	apiUrlWithParams := apiUrl + "?" + params.Encode()

	response, err := http.Get(apiUrlWithParams)
	if err != nil {
		log.Printf("Error al consultar iTunes: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("Error al leer body: %v", err)
		}
	}(response.Body)

	body, err := io.ReadAll(response.Body)

	var iTunesResult ITunesSearch
	if err := json.Unmarshal(body, &iTunesResult.Data); err != nil {
		log.Printf("Error al decodificar respuesta de iTunes: %v", err)
		return nil
	}

	return iTunesResult.Data.Results

}

func getChartLyrics(w http.ResponseWriter, name string, artist string) SearchLyricResult {

	apiUrl := "http://api.chartlyrics.com/apiv1.asmx/SearchLyric"

	params := url.Values{}
	params.Add("artist", artist)
	params.Add("song", name)

	apiUrlWithParams := apiUrl + "?" + params.Encode()

	response, err := http.Get(apiUrlWithParams)
	if err != nil {
		log.Printf("Error al consultar Chart Lyrics: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("Error al leer body: %v", err)
		}
	}(response.Body)

	xmlData, err := io.ReadAll(response.Body)
	if err != nil {
		log.Printf("Error al leer cuerpo xml: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	var result SearchLyricResult

	if err := xml.Unmarshal(xmlData, &result); err != nil {
		log.Printf("Error al analizar cuerpo xml: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	return result

}

func saveCache(name string, artist string, album string, db *sql.DB) []Song {
	c := cache.New(5*time.Minute, 10*time.Minute)

	var data []Song

	cacheData, found := c.Get("cachedData")
	if found {
		data = cacheData.([]Song)
	} else {
		rows, err := db.Query("SELECT track_id, name, artist, duration, album, artwork, price, origin FROM songs where name like CONCAT('%', ?, '%') and artist like CONCAT('%', ?, '%') or album like CONCAT('%', ?, '%') ",
			name, artist, album)
		if err != nil {
			log.Printf("Error al obtener datos de la base de datos: %v", err)
			return nil
		}

		defer func(rows *sql.Rows) {
			err := rows.Close()
			if err != nil {
				log.Printf("Error al leer body: %v", err)
			}
		}(rows)

		for rows.Next() {
			var song Song
			err := rows.Scan(&song.Id, &song.Name, &song.Artist, &song.Duration, &song.Album, &song.Artwork, &song.Price, &song.Origin)
			if err != nil {
				log.Printf("Error al scanear datos: %v", err)
				return nil
			}

			data = append(data, song)
		}

		c.Set("cachedData", data, cache.DefaultExpiration)
	}
	return data
}
