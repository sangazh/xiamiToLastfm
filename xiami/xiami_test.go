package xiami

import (
	"testing"
	"fmt"
)

func TestParseUrl(t *testing.T) {
	url := "http://www.xiami.com/u/37185?spm=a1z1s.6928797.1561534497.2.dETWvH"
	userId, spm := parseUrl(url)
	if userId != "37185" || spm !=  "a1z1s.6928797.1561534497.2.dETWvH" {
		t.Error("parseUrl() error")
	}
}

func TestParseTime(t *testing.T) {
	test := "1分钟前"
	timestamp, _, _ := parseTime(test)
	fmt.Println(test, timestamp)

}
