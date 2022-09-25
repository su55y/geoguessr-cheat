package location

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
)

const (
	locUrl = "https://nominatim.openstreetmap.org/reverse?lat=%s&lon=%s&format=json&accept-language=en"
)

var (
	rCoords = regexp.MustCompile("\\[null,null,([\\d\\.\\-]+),([\\d\\.\\-]+)\\],")
)

func ProceedUrl(u string, w *http.ResponseWriter, l *log.Logger) error {
	if r, ok := getGeoMeta(u, l); ok {
		if coords, ok := parseCoords(r, l); ok {
			if loc, ok := getLocation(coords, l); ok {
				return json.NewEncoder(*w).Encode(loc)
			}

			l.Println("can't get location from open street maps")
			return errors.New("open street maps request failed")
		}

		l.Println("can't parse coords")
		return errors.New("parse coords error")
	}

	l.Println("can't get geo meta")
	return errors.New("geo meta request failed")
}

func getLocation(c Coords, l *log.Logger) (Location, bool) {
	resp, err := http.Get(fmt.Sprintf(locUrl, c.Lat, c.Lng))
	if err != nil {
		log.Printf("location request error: %v\n", err.Error())
		return Location{}, false
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var loc Location
		err := json.NewDecoder(resp.Body).Decode(&loc)
		if err != nil {
			log.Printf("location response decoding error: %v", err.Error())
			return Location{}, false
		}

		return loc, true
	}
	log.Println("location request failed")
	return Location{}, false
}

func parseCoords(s string, l *log.Logger) (Coords, bool) {
	if len(s) > 0 {
		if rCoords.MatchString(s) {
			if r := rCoords.FindAllStringSubmatch(s, -1); r != nil && len(r) > 0 {
				for _, v := range r {
					if len(v) == 3 {
						return Coords{Lat: v[1], Lng: v[2]}, true
					}
				}
			}
		}
	}
	return Coords{}, false
}

func getGeoMeta(url string, l *log.Logger) (string, bool) {
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("request error: %v\n", err.Error())
		return "", false
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		if bodyBytes, err := io.ReadAll(resp.Body); err == nil {
			return string(bodyBytes), true
		} else {
			log.Printf("request read error: %v\n", err.Error())
		}
	}
	log.Println("request failed")

	return "", false
}
