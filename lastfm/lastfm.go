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
	"xiamiToLastfm/musicbrainz"

	"github.com/theherk/viper"
)

var (
	domain, apiUrl, sharedSecret, apiKey string
)

type ScrobbleParams struct {
	Artist            []string `json:"artist,omitempty"`
	Track             []string `json:"track,omitempty"`
	Timestamp         []string `json:"timestamp,omitempty"`
	Album             []string `json:"album,omitempty"`
	TrackNumber       []string `json:"trackNumber,omitempty"`
	Mbid              []string `json:"mbid,omitempty"` //The MusicBrainz Track ID
	AlbumArtist       []string `json:"albumArtist,omitempty"`
	DurationInSeconds []string `json:"duration,omitempty"`
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

	if mbid, ok := musicbrainz.MbID(xm.Title, xm.Artist, xm.Album); ok {
		v.Set("mbid[0]", string(mbid))
	}

	v.Set("timestamp[0]", fmt.Sprint(xm.Timestamp))
	v.Set("method", "track.scrobble")
	v.Set("sk", sk)
	v.Set("api_key", apiKey)
	sig := signature(&v)
	v.Set("api_sig", sig)
	v.Set("format", "json")

	respData, err := postRequest(v.Encode(), quitChan)
	if err != nil {
		log.Println("last.fm: ", err)
		fmt.Println("last.fm: scrobble sent failed. Try later.")
		//if failed, insert back to channel
		playedChan <- xm
		return false
	}

	accepted, ignored := scrobbleResponse(respData)
	log.Printf("last.fm: Scrobbled succese - accepted: %d, ignored: %d\n", accepted, ignored)
	fmt.Printf("last.fm: Scrobbled succese. %s - %s \n", xm.Title, xm.Artist)

	//写下执行时间
	if len(playedChan) < 1 {
		viper.Set("xiami.checked_at", time.Now().Truncate(time.Minute).Unix())
		viper.WriteConfig()
	}
	return true
}

func scrobbleResponse(data []byte) (accepted, ignored int) {
	type response struct {
		Data struct {
			Msg struct {
				Accepted int `json:"accepted"`
				Ignored  int `json:"ignored"`
			} `json:"@attr"`
		} `json:"scrobbles"`
	}

	var resp response
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

	_, err := postRequest(v.Encode(), quitChan)
	if err != nil {
		fmt.Println("last.fm: UpdateNowPlaying sent failed.")
		log.Println("last.fm: ", err)
		//if failed, as discard.
		return false
	}

	fmt.Printf("last.fm: UpdateNowPlaying success. %s - %s \n", xm.Title, xm.Artist)
	return true
}

func getRequest(url string) ([]byte, error) {
	log.Println("last.fm: getRequest url: ", url)
	res, err := http.Get(url)

	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	resData, _ := ioutil.ReadAll(res.Body)
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("status code error: '%s' on %s body: %s", res.Status, url, string(resData))
	}

	log.Println("last.fm: getRequest response: ", string(resData))
	return resData, nil
}

func postRequest(query string, quitChan chan struct{}) ([]byte, error) {
	r := bytes.NewReader([]byte(query))
	contentType := "application/x-www-form-urlencoded"

	log.Println("last.fm: postRequest params: ", query)
	res, err := http.Post(apiUrl, contentType, r)

	if err != nil {
		return nil, fmt.Errorf("postRequest has error: %s", err)
	}

	resData, _ := ioutil.ReadAll(res.Body)
	defer res.Body.Close()

	if res.StatusCode != 200 {
		errCode, errMsg := handleError(resData)
		if errCode == 9 {
			fmt.Println(errMsg)
			resetAuth()
			fmt.Println("Config reset. Please re-start the program.")
			close(quitChan)
			os.Exit(1)
		}
		return nil, fmt.Errorf("status code error: '%s' on %s body: %s", res.Status, apiUrl, string(resData))

	}
	log.Println("last.fm: postRequest response: ", string(resData))
	return resData, nil
}

func toMap(byteData []byte) (result map[string]string) {
	r := bytes.NewReader(byteData)
	json.NewDecoder(r).Decode(&result)
	return result
}

func handleError(errData []byte) (code int, msg string) {
	type ErrResponse struct {
		Code int    `json:"error"`
		Msg  string `json:"message"`
	}

	var resp ErrResponse
	json.Unmarshal(errData, &resp)
	return resp.Code, resp.Msg
}
