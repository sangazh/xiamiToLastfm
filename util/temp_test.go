package util

import (
	"testing"
	"xiamiToLastfm/xiami"
	"github.com/stretchr/testify/assert"
)

func TestTempStore(t *testing.T) {
	playedChan := make(chan xiami.Track, 10)
	defer close(playedChan)
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
	assert.True(t, TempStore(playedChan))
}

func TestTempRead(t *testing.T) {
	playedChan := make(chan xiami.Track, 10)
	defer close(playedChan)
	assert.True(t, TempRead(playedChan))
	assert.Equal(t, 2, len(playedChan))
}