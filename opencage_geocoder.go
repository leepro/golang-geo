package geo

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

// This struct contains all the funcitonality
// of interacting with the OpenCage Geocoding Service
type OpenCageGeocoder struct{}

// This struct contains selected fields from OpenCage's Geocoding Service response
type opencageGeocodeResponse struct {
	Results []struct {
		Formatted string `json:"formatted"`
		Geometry  struct {
			Lat float64
			Lng float64
		}
	}
}

// This is the error that consumers receive when there
// are no results from the geocoding request.
var opencageZeroResultsError = errors.New("ZERO_RESULTS")

// This contains the base URL for the Mapquest Geocoder API.
var opencageGeocodeURL = "http://api.opencagedata.com/geocode/v1/json"

// Note:  In the next major revision (1.0.0), it is planned
//        That Geocoders should adhere to the `geo.Geocoder`
//        interface and provide versioning of APIs accordingly.
// Sets the base URL for the OpenCage Geocoding API.
func SetOpenCageGeocodeURL(newGeocodeURL string) {
	opencageGeocodeURL = newGeocodeURL
}

// Issues a request to the open OpenCage API geocoding services using the passed in url query.
// Returns an array of bytes as the result of the api call or an error if one occurs during the process.
func (g *OpenCageGeocoder) Request(url string) ([]byte, error) {
	client := &http.Client{}
	fullUrl := fmt.Sprintf("%s/%s", opencageGeocodeURL, url)

	// TODO Refactor into an api driver of some sort
	//      It seems odd that golang-geo should be responsible of versioning of APIs, etc.
	req, _ := http.NewRequest("GET", fullUrl, nil)
	resp, requestErr := client.Do(req)

	if requestErr != nil {
		panic(requestErr)
	}

	// TODO figure out a better typing for response
	data, dataReadErr := ioutil.ReadAll(resp.Body)

	if dataReadErr != nil {
		return nil, dataReadErr
	}

	return data, nil
}

// Returns the first point returned by OpenCage's geocoding service or an error
// if one occurs during the geocoding request.
func (g *OpenCageGeocoder) Geocode(query string) (*Point, error) {
	url_safe_query := url.QueryEscape(query)
	data, err := g.Request(fmt.Sprintf("?q=%s&pretty=1", url_safe_query))
	if err != nil {
		return nil, err
	}

	res := &opencageGeocodeResponse{}
	err = json.Unmarshal(data, &res)
	if err != nil {
		return nil, err
	}

	if len(res.Results) == 0 {
		return nil, opencageZeroResultsError
	}

	p := &Point{
		lat: res.Results[0].Geometry.Lat,
		lng: res.Results[0].Geometry.Lng,
	}

	return p, nil
}

// Returns the first most available address that corresponds to the passed in point.
// It may also return an error if one occurs during execution.
func (g *OpenCageGeocoder) ReverseGeocode(p *Point) (string, error) {
	data, err := g.Request(fmt.Sprintf("?q=lat=%f,lon=%f&pretty=1", p.lat, p.lng))
	if err != nil {
		return "", err
	}

	res := &opencageGeocodeResponse{}
	err = json.Unmarshal(data, &res)
	if err != nil {
		return "", err
	}

	resStr := res.Results[0].Formatted

	return resStr, nil
}