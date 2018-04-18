package lastfm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"xiamiToLastfm/xiami"

	"github.com/theherk/viper"
)

var (
	domain, apiUrl, sharedSecret,apiKey  string
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

func StartScrobble(playedChan chan xiami.Track, quitChan chan struct{}) bool {
	xm := <-playedChan

	log.Println("last.fm: playedChan track: ", xm)

	v := url.Values{}
	v.Set("artist[0]", xm.Artist)
	v.Set("album[0]", xm.Album)
	v.Set("track[0]", xm.Title)
	v.Set("timestamp[0]", fmt.Sprint(xm.Timestamp))
	v.Set("method", "track.scrobble")
	v.Set("sk", sk)
	v.Set("api_key", apiKey)
	sig := signature(&v)
	v.Set("api_sig", sig)
	v.Set("format", "json")

	respData, ok := postRequest(v.Encode(), quitChan)
	if !ok {
		fmt.Println("last.fm: scrobble sent failed. Try later.")

		//if failed, insert back to channel
		playedChan <- xm
		return false
	}

	accepted, ignored := renderScrobbleResp(respData)
	log.Printf("last.fm: Scrobbled succese - accepted: %d, ignored: %d\n", accepted, ignored)
	fmt.Printf("last.fm: Scrobbled succese. %s - %s \n", xm.Title, xm.Artist)

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

func UpdateNowPlaying(nowPlayingChan chan xiami.Track, quitChan chan struct{}) bool {
	xm := <-nowPlayingChan
	log.Println("last.fm: nowPlayingChan track: ", xm)

	v := url.Values{}
	v.Set("method", "track.updateNowPlaying")
	v.Set("sk", sk)
	v.Set("api_key", apiKey)
	v.Set("artist", xm.Artist)
	v.Set("album", xm.Album)
	v.Set("track", xm.Title)
	sig := signature(&v)
	v.Set("api_sig", sig)
	v.Set("format", "json")

	_, ok := postRequest(v.Encode(), quitChan)
	if !ok {
		fmt.Println("last.fm: UpdateNowPlaying sent failed.")
		//if failed, as discard.
		return false
	}

	fmt.Printf("last.fm: UpdateNowPlaying success. %s - %s \n", xm.Title, xm.Artist)
	return true
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

func postRequest(query string, quitChan chan struct{}) ([]byte, bool) {
	r := bytes.NewReader([]byte(query))
	contentType := "application/x-www-form-urlencoded"

	log.Println("last.fm: postRequest params: ", query)
	res, err := http.Post(apiUrl, contentType, r)

	if err != nil {
		log.Println("last.fm: postRequest has error: ", err)
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
			os.Exit(1)
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
