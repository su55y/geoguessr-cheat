package location

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"
)

const (
	locUrl = "https://nominatim.openstreetmap.org/reverse?lat=%s&lon=%s&format=json&accept-language=en"
)

var (
	rxCoords  = regexp.MustCompile("\\[null,null,([\\d\\.\\-]+),([\\d\\.\\-]+)\\],")
	geoMetaRX = regexp.MustCompile(
		"^\\S+req\\?u=(https:\\/\\/maps\\.googleapis\\.com\\/maps\\/api\\/js\\/GeoPhotoService\\.GetMetadata\\S+&callback=\\S+)$",
	)
	errInvalidBody = func(s string) error { return fmt.Errorf("invalid geoMetaBody format: %+s", s) }
)

func WriteLocation(requestUrl string, w *http.ResponseWriter) error {
	metaUrl, parseErr := parseUrl(requestUrl)
	if parseErr != nil {
		return parseErr
	}

	geoMetaBody, err := fetchGeoMetaBody(metaUrl)
	if err != nil {
		return err
	}

	coords := parseGeoMetaBody(geoMetaBody)
	if coords == nil {
		return errInvalidBody(geoMetaBody)
	}

	loc, err := getLocation(*coords)
	if err != nil {
		return err
	}

	return json.NewEncoder(*w).Encode(loc)
}

func getLocation(c Coords) (*Location, error) {
	u := fmt.Sprintf(locUrl, c.Lat, c.Lng)
	resp, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s %s", resp.Status, u)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var loc Location
	if err := json.Unmarshal(body, &loc); err != nil {
		return nil, err
	}

	return &loc, nil
}

func parseGeoMetaBody(body string) *Coords {
	if !rxCoords.MatchString(body) {
		return nil
	}

	matches := rxCoords.FindAllStringSubmatch(body, -1)
	if matches == nil || len(matches) == 0 {
		return nil
	}

	for _, group := range matches {
		if len(group) == 3 {
			return &Coords{Lat: group[1], Lng: group[2]}
		}
	}

	return nil
}

func fetchGeoMetaBody(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%s %s", resp.Status, url)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func parseUrl(reqUrl string) (string, error) {
	if !geoMetaRX.MatchString(reqUrl) {
		return "", fmt.Errorf("invalid url %s", reqUrl)
	}
	matches := geoMetaRX.FindStringSubmatch(reqUrl)
	if matches == nil || len(matches) != 2 {
		return "", fmt.Errorf("cannot parse url %s", reqUrl)
	}
	return matches[1], nil
}
