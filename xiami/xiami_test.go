package xiami

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseUrl(t *testing.T) {
	url := "http://www.xiami.com/u/37185?spm=a1z1s.6928797.1561534497.2.dETWvH"
	userId, spm := parseUrl(url)

	assert.Equal(t, "37185", userId)
	assert.Equal(t, "a1z1s.6928797.1561534497.2.dETWvH", spm)
}

func TestParseTime(t *testing.T) {
	test := "1分钟前"
	timestamp, _, _ := parseTime(test)
	now := time.Now().Truncate(time.Minute).Unix()
	var expected int64 = 60
	assert.Equal(t, expected, now-timestamp)
}

func TestGetAlbum(t *testing.T) {
	url := "https://www.xiami.com/song/mSJtnV7aa77?spm=a1z1s.6626017.0.0.nGwy1E"
	artist, album, ok := album(url)
	assert.True(t, ok)
	assert.Equal(t, "川井憲次" ,artist)
	assert.Equal(t, "NHKスペシャル「人体 神秘の巨大ネットワーク」オリジナル・サウンドトラック" ,album)
}
