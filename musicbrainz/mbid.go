package musicbrainz

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	"github.com/spf13/viper"
)

// get record's MBID from its title, artist and album.
func MbID(title, artist, album string) (id MBID, ok bool) {
	requestUrl := prepareUrl(title, artist, album)
	res, err := SearchRecording(requestUrl)
	if err != nil {
		log.Println("musicbrainz err: ", err)
		return "", false
	}

	matched := res.ResultsWithScore(100)
	if len(matched) > 0 {
		return matched[0].ID, true
	}

	return "", false
}

func prepareUrl(title, artist, album string) string {
	domain := viper.GetString("musicbrainz.domain")
	recUrl := viper.GetString("musicbrainz.url.recording")

	query := fmt.Sprintf("artistname:%s AND release:%s AND recording:%s", artist, album, title)
	v := url.Values{}
	v.Set("query", query)
	v.Set("limit", "3")

	u, _ := url.Parse(domain)
	u.Path += recUrl
	u.RawQuery = v.Encode()
	return u.String()
}

func getRequest(url string, result interface{}) error {
	log.Println("musicbrainz: getRequest url: ", url)
	res, err := http.Get(url)

	if err != nil {
		return err
	}
	defer res.Body.Close()
	resData, _ := ioutil.ReadAll(res.Body)

	if res.StatusCode != 200 {
		return fmt.Errorf("status code error: '%s' on %s", res.Status, url)
	}

	err = xml.Unmarshal(resData, result)
	if err != nil {
		return fmt.Errorf("xml docode err: %s", err)
	}

	return nil

}
