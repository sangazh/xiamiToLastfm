package xiami

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/spf13/viper"
)

type Track struct {
	Title     string `json:"songName"`
	Artist    string `json:"singers"`
	Album     string `json:"albumName"`
	Alias     string `json:"translation"`
	Timestamp int64 `json:"-"`
}

type SongData struct {
	Data []*Track `json:"data"`
}

func ReadFile() (*SongData, error) {
	r, err := os.Open(viper.GetString("xiami.file"))

	if err != nil {
		return nil, err
	}
	defer r.Close()
	b, _ := ioutil.ReadAll(r)

	data := new(SongData)
	if err := json.Unmarshal(b, data); err != nil {
		return nil, err
	}

	return data, nil
}
