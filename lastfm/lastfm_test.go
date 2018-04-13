package lastfm

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"testing"

	"xiami2LastFM/xiami"

	"github.com/stretchr/testify/assert"
)

func TestCheckAuth(t *testing.T) {
	checkAuth()
}

func TestPrepareSigText(t *testing.T) {
	v := url.Values{}
	v.Set("a", "123")
	v.Set("c", "345")
	v.Set("b", "567")

	result := prepareSigText(v)
	expected := "a123b567c345"
	assert.Equal(t, expected, result)
}

func TestQueryString(t *testing.T) {
	query := map[string]interface{}{
		"a": "123",
		"c": "345",
		"b": "567",
	}
	result := queryString(query)
	expected := "a=123&c=345&b=567"
	assert.Equal(t, expected, result)

	queryEmpty := make(map[string]interface{})
	result = queryString(queryEmpty)
	expected = ""
	assert.Equal(t, expected, result)

	query = map[string]interface{}{
		"album[0]": "TVアニメ「血界戦線&BEYOND」オリジナルサウンドトラック",
		//"c": "345",
		//"b": "567",
	}

	result = queryString(query)
	fmt.Println(result)

}

func TestSignature(t *testing.T) {
	v := url.Values{}
	v.Set("method", "auth.getSession")
	v.Set("api_key", "4778db9e5d5b2dd00fb34792ac28c1c1")
	v.Set("token", "9V6bP2X4OZJcMi7IRz2M50w_IAWxZ1TC")
	result := signature(&v)
	if len(result) != 32 {
		t.Error(`signature() length not equal to 32`)
	}
	expected := `fb441e79f2820ccd7d7bd71174ff34f9` //e05dd3b746c95c6d5d896cd7079757fe

	assert.Equal(t, expected, result)
}

func TestSignature2(t *testing.T) {
	v := url.Values{}
	v.Set("method", "auth.getSession")
	v.Set("api_key", "4778db9e5d5b2dd00fb34792ac28c1c1")
	v.Set("token", "9V6bP2X4OZJcMi7IRz2M50w_IAWxZ1TC")
	v.Set("album[0]", "TVアニメ「血界戦線&BEYOND」オリジナルサウンドトラック")

	result := signature(&v)
	fmt.Println(v.Encode())

	if len(result) != 32 {
		t.Error(`signature() length not equal to 32`)
	}
	expected := `43da2b8501daadcc39738998ed16dd26`
	assert.Equal(t, expected, result)
}

func TestScrobbleSignature(t *testing.T) {
	v := url.Values{}
	/*
	query := map[string]interface{}{
		"artist[0]":    "服部隆之",
		"artist[1]":    "服部隆之",
		"track[0]":     "自己顕示欲",
		"track[1]":     "仕事",
		"album[0]":     "TVアニメ『ID-0』オリジナルサウンドトラック",
		"album[1]":     "TVアニメ『ID-0』オリジナルサウンドトラック",
		"timestamp[0]": "1523328000",
		"timestamp[1]": "1523327819",
		"api_key":      "4778db9e5d5b2dd00fb34792ac28c1c1",
		"sk":           "rIr2HM8h5s-_t-5nM0PKzPL8m7tjGxgH",
		"method":       "track.scrobble",
	}*/

	result := signature(&v)

	assert.Equal(t, 32, len(result))

	expected := `257e6be2dbc096e5a89a63ce7555bb09`
	assert.Equal(t, expected, result)
}

func TestParseKey(t *testing.T) {
	r, err := os.Open("sample/sk.json")

	if err != nil {
		t.Error(err)
	}
	defer r.Close()
	bytes, _ := ioutil.ReadAll(r)
	token, ok := parseKey(bytes)
	assert.Equal(t, 32, len(token))
	assert.True(t, ok)
}

func TestStartScrobble(t *testing.T) {
	playedChan := make(chan interface{}, 10)
	quitChan := make(chan struct{})
	defer close(playedChan)
	defer close(quitChan)
	track := xiami.Track{
		Title:     "自己顕示欲",
		Album:     "TVアニメ『ID-0』オリジナルサウンドトラック",
		Artist:    "服部隆之",
		Timestamp: 1523328000,
	}
	playedChan <- track
	track = xiami.Track{
		Title:     "仕事",
		Album:     "TVアニメ『ID-0』オリジナルサウンドトラック",
		Artist:    "服部隆之",
		Timestamp: 1523327819,
	}
	playedChan <- track
	assert.True(t, StartScrobble(playedChan, quitChan))
}

func TestUpdateNowPlaying(t *testing.T) {
	nowPlayingChan := make(chan interface{}, 10)
	quitChan := make(chan struct{})
	defer close(nowPlayingChan)
	defer close(quitChan)

	track := xiami.Track{
		Title:     "自己顕示欲",
		Album:     "TVアニメ『ID-0』オリジナルサウンドトラック",
		Artist:    "服部隆之",
		Timestamp: 1523328000,
	}
	nowPlayingChan <- track
	assert.True(t, UpdateNowPlaying(nowPlayingChan, quitChan))
}

func TestRenderScrobbleResp(t *testing.T) {
	r, err := os.Open("sample/scrobble.json")

	if err != nil {
		t.Error(err)
	}
	defer r.Close()
	data, _ := ioutil.ReadAll(r)
	accepted, ignored := renderScrobbleResp(data)

	assert.Equal(t, 2, accepted)
	assert.Equal(t, 0, ignored)
}

func TestHandleError(t *testing.T) {
	r, err := os.Open("sample/error.json")

	if err != nil {
		t.Error(err)
	}
	defer r.Close()
	data, _ := ioutil.ReadAll(r)
	code, msg := handleError(data)
	assert.Equal(t, 9, code)
	assert.Equal(t, "Invalid session key - Please re-authenticate", msg)
}
