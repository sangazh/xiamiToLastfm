package musicbrainz

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestSearch(t *testing.T) {
	title := "武士"
	artist := "吉田潔"
	album := "武士~もののふ"
	_, ok := MbID(title, artist, album)
	assert.True(t, ok)
}
