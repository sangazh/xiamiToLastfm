package lastfm

import (
	"io/ioutil"
	"net/url"
	"os"
	"testing"

	"xiamiToLastfm/xiami"

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

func TestSignature(t *testing.T) {
	v := url.Values{}
	v.Set("method", "auth.getSession")
	v.Set("api_key", "4778db9e5d5b2dd00fb34792ac28c1c1")
	v.Set("token", "9V6bP2X4OZJcMi7IRz2M50w_IAWxZ1TC")
	sharedSecret = "bb471918f14e2b29e219185d4591baa6"
	result := signature(&v)
	if len(result) != 32 {
		t.Error(`signature() length not equal to 32`)
	}
	expected := `cec68f75b328c48f2749cf6cc6c7e5cb`

	assert.Equal(t, expected, result)
}

func TestSignature2(t *testing.T) {
	v := url.Values{}
	v.Set("method", "auth.getSession")
	v.Set("api_key", "4778db9e5d5b2dd00fb34792ac28c1c1")
	v.Set("token", "9V6bP2X4OZJcMi7IRz2M50w_IAWxZ1TC")
	v.Set("album[0]", "TVアニメ「血界戦線&BEYOND」オリジナルサウンドトラック")
	sharedSecret = "bb471918f14e2b29e219185d4591baa6"
	result := signature(&v)

	if len(result) != 32 {
		t.Error(`signature() length not equal to 32`)
	}
	expected := `c50b68fc0415def12fc048b9efaccdf1`
	assert.Equal(t, expected, result)
}

func TestScrobbleSignature(t *testing.T) {
	v := url.Values{}

	v.Set("artist[0]", "服部隆之")
	v.Set("artist[1]", "服部隆之")
	v.Set("track[0]", "自己顕示欲")
	v.Set("track[1]", "仕事")
	v.Set("album[0]", "TVアニメ『ID-0』オリジナルサウンドトラック")
	v.Set("album[1]", "TVアニメ『ID-0』オリジナルサウンドトラック")
	v.Set("timestamp[0]", "1523328000")
	v.Set("timestamp[1]", "1523327819")
	v.Set("api_key", "4778db9e5d5b2dd00fb34792ac28c1c1")
	v.Set("sk", "rIr2HM8h5s-_t-5nM0PKzPL8m7tjGxgH")
	v.Set("method", "track.scrobble")
	sharedSecret = "bb471918f14e2b29e219185d4591baa6"
	result := signature(&v)

	assert.Equal(t, 32, len(result))

	expected := `257e6be2dbc096e5a89a63ce7555bb09`
	assert.Equal(t, expected, result)
}

func TestSignatureWithSpecial(t *testing.T) {
	album := "TVアニメ「血界戦線&BEYOND」オリジナルサウンドトラック"
	v := url.Values{}
	v.Set("artist[0]", "岩崎太整")
	v.Set("track[0]", "BLOCK SCHOLARS feat. Hybrid Thoughts")
	v.Set("album[0]", album)
	v.Set("timestamp[0]", "1523608740")
	v.Set("api_key", apiKey)
	v.Set("sk", "1hWanAAsPJgERBmM-1jF04GFEjtz4lJG")
	v.Set("method", "track.scrobble")
	sharedSecret = "bb471918f14e2b29e219185d4591baa6"
	sig := signature(&v)
	expected := `7bb79d9a17dea0a205725c453295e8d6`
	assert.Equal(t, expected, sig)

	v.Set("api_sig", sig)
	v.Set("format", "json")

	s := v.Encode()
	v2, _ := url.ParseQuery(s)

	assert.Equal(t, album, v2.Get("album[0]"))
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
	apiUrl, _ = url.Parse("http://ws.audioscrobbler.com/2.0")
	apiKey = "4778db9e5d5b2dd00fb34792ac28c1c1"
	token = "9V6bP2X4OZJcMi7IRz2M50w_IAWxZ1TC"

	playedChan := make(chan xiami.Track, 10)
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
	nowPlayingChan := make(chan xiami.Track, 10)
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
	accepted, ignored := scrobbleResponse(data)

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
