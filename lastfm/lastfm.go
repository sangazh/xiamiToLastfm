package lastfm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"xiami2LastFM/xiami"

	"github.com/theherk/viper"
)

var (
	domain, apiKey, sharedSecret, apiUrl string
)

type ScrobbleParams struct {
	Artist            []string `json:"artist,omitempty"`
	Track             []string `json:"track,omitempty"`
	Timestamp         []string `json:"timestamp,omitempty"`
	Album             []string `json:"album,omitempty"`
	TrackNumber       []string `json:"trackNumber,omitempty"`
	Mbid              []string `json:"mbid,omitempty"` //The MusicBrainz Track ID
	AlbumArtist       []string `json:"albumArtist,omitempty"`
	DurationInSeconds []string `jsonapikey:"duration,omitempty"`
	ApiKey            string   `json:"api_key"`
	ApiSig            string   `json:"api_sig"`
	Sk                string   `json:"sk"`
	Format            string   `json:"format"`
	Method            string   `json:"method"`
}

func StartScrobble(playedChan chan interface{}, quitChan chan struct{}) bool {
	t := <-playedChan
	if t == nil {
		return false
	}

	xm := t.(xiami.Track)
	log.Println("last.fm: playedChan track: ", xm)

	query := make(map[string]interface{}, 0)

	query["artist[0]"] = xm.Artist
	query["album[0]"] = xm.Album
	query["track[0]"] = xm.Title
	query["timestamp[0]"] = fmt.Sprint(xm.Timestamp)

	query["method"] = "track.scrobble"
	query["sk"] = sk
	query["api_key"] = apiKey
	query["api_sig"] = signature(query)
	query["format"] = "json"
	log.Println("last.fm: StartScrobble - request query:", query)

	respData, ok := postRequest(query, quitChan)
	if !ok {
		fmt.Println("last.fm: scrobble sent failed.")
		return false
	}

	accepted, ignored := renderScrobbleResp(respData)
	fmt.Printf("last.fm: Scrobbled succeseful - accepted: %d, ignored: %d\n", accepted, ignored)

	//写下执行时间
	if len(playedChan) < 1 {
		viper.Set("xiami.checked_at", time.Now().Truncate(time.Minute).Unix())
		viper.WriteConfig()
	}
	return true
}

type ScrobbleResponse struct {
	Data ScrobbleData `json:"scrobbles"`
}

type ScrobbleData struct {
	Msg ScrobbleMsg `json:"@attr"`
}

type ScrobbleMsg struct {
	Accepted int `json:"accepted"`
	Ignored  int `json:"ignored"`
}

func renderScrobbleResp(data []byte) (accepted, ignored int) {
	var resp ScrobbleResponse
	json.Unmarshal(data, &resp)
	return resp.Data.Msg.Accepted, resp.Data.Msg.Ignored
}

func UpdateNowPlaying(nowPlayingChan chan interface{}, quitChan chan struct{}) bool {
	t := <-nowPlayingChan
	if t == nil {
		return false
	}

	xm := t.(xiami.Track)
	log.Println("last.fm: nowPlayingChan track: ", xm)

	query := map[string]interface{}{
		"method":  "track.updateNowPlaying",
		"sk":      sk,
		"api_key": apiKey,
		"artist":  xm.Artist,
		"album":   xm.Album,
		"track":   xm.Title,
	}

	query["api_sig"] = signature(query)
	query["format"] = "json"

	_, ok := postRequest(query, quitChan)
	if !ok {
		fmt.Println("last.fm: UpdateNowPlaying sent failed.")
		return false
	}

	fmt.Println("nowPlaying", xm)
	return true
}

// query input as map format, expect output with format of key=value&key=value
func queryString(query map[string]interface{}) (text string) {
	for key, value := range query {
		text += fmt.Sprintf("%s=%s&", key, value)
	}

	return strings.TrimRight(text, "&")
}

func getRequest(url string) ([]byte, bool) {
	log.Println("last.fm: getRequest url: ", url)
	res, err := http.Get(url)

	if err != nil {
		log.Println(err)
		return nil, false
	}
	defer res.Body.Close()

	resData, _ := ioutil.ReadAll(res.Body)

	if res.StatusCode != 200 {
		log.Printf("status code error: %d %s on %s ", res.StatusCode, res.Status, url)
		log.Println("err body: ", string(resData))
		return nil, false
	}
	log.Println("last.fm: getRequest response: ", string(resData))

	return resData, true
}

func postRequest(query map[string]interface{}, quitChan chan struct{}) ([]byte, bool) {
	r := bytes.NewReader([]byte(queryString(query)))
	contentType := "application/x-www-form-urlencoded"

	log.Println("last.fm: postRequest params: ", query)
	res, err := http.Post(apiUrl, contentType, r)

	if err != nil {
		log.Println(err)
		return nil, false
	}

	resData, _ := ioutil.ReadAll(res.Body)
	defer res.Body.Close()

	if res.StatusCode != 200 {
		log.Printf("last.fm: postRequest status code error: %d %s on %s ", res.StatusCode, res.Status, apiUrl)
		log.Println("last.fm: postReques err body: ", string(resData))
		errCode, errMsg := handleError(resData)
		if errCode == 9 {
			fmt.Println(errMsg)
			resetAuth()
			fmt.Println("Config reset. Please re-start the program.")
			close(quitChan)
		}
		return nil, false
	}
	log.Println("last.fm: postReques response: ", string(resData))

	return resData, true
}

func toMap(byteData []byte) (result map[string]string) {
	r := bytes.NewReader(byteData)
	json.NewDecoder(r).Decode(&result)
	return result
}

type ErrResponse struct {
	Code int    `json:"error"`
	Msg  string `json:"message"`
}

func handleError(errData []byte) (code int, msg string) {
	var resp ErrResponse
	json.Unmarshal(errData, &resp)
	return resp.Code, resp.Msg
}
