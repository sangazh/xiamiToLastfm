package xiami

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/theherk/viper"
)

const (
	typeNowPlaying = iota + 1
	typePlayed
)

var (
	domain, uri, spm string
	userId           int
)

type Track struct {
	Title, Artist, Album string
	Timestamp            int64
}

func Init() {
	if checkConfig() {
		return
	}
	fmt.Println("Press Enter xiami profile page url")
	profileUrl, _ := bufio.NewReader(os.Stdin).ReadString('\n')

	parseUrl(profileUrl)
	return
}

func checkConfig() bool {
	domain = viper.GetString("xiami.domain")
	uri = domain + viper.GetString("xiami.url.recent")
	userId = viper.GetInt("xiami.user_id")
	spm = viper.GetString("xiami.spm")

	if userId > 0 && spm != "" {
		return true
	}
	return false
}

//parseUrl and save to config
func parseUrl(rawUrl string) (userId, spm string) {
	u, _ := url.Parse(rawUrl)
	spm = u.Query().Get("spm")

	s := strings.Split(u.Path, "/")
	userId = s[2]

	viper.Set("xiami.userId", userId)
	viper.Set("xiami.spm", spm)
	viper.WriteConfig()

	return userId, spm
}

func GetTracks(playingChan, playedChan chan interface{}) {
	lastCheckAt := viper.GetInt64("xiami.checked_at")
	requestUrl := fmt.Sprintf("%s%d", uri, userId)

	doc, err := getDoc(requestUrl)
	if err != nil || doc == nil {
		return
	}

	// Find the review items
	doc.Find(".track_list tr").Each(func(i int, s *goquery.Selection) {
		trackTime := s.Find(".track_time").Text()
		timeStamp, scrobbleType, ok := parseTime(trackTime)

		if !ok || timeStamp < lastCheckAt {
			return
		}

		title, _ := s.Find(".song_name a").Attr("title")
		trackUrl, _ := s.Find(".song_name a").Attr("href")
		artist, album, ok := getAlbum(trackUrl)
		if !ok {
			return
		}

		t := Track{Title: title, Artist: artist, Album: album, Timestamp: timeStamp}

		switch scrobbleType {
		case typeNowPlaying:
			playingChan <- t
			fmt.Printf("nowPlaying: %s - %s 《%s》 \n", title, artist, album)
			log.Println("xiami: GetTrack - playingChan <- t ", t)
		case typePlayed:
			playedChan <- t
			fmt.Printf("Listened: %d: %s - %s 《%s》 \n", i, title, artist, album)
			log.Println("xiami: GetTrack - playedChan <- t ", t)
		default:
			log.Println("xiami: GetTrack - switch default")
		}
	})
	log.Println("xiami: GetTrack returned.")
	return
}

// if time before 1 hour, then exact time cannot calculated, abort
func parseTime(s string) (t int64, srbType int, ok bool) {
	if strings.HasSuffix(s, "分钟前") { //播放完毕可以同步
		minutes, _ := strconv.Atoi(strings.TrimSuffix(s, "分钟前"))
		duration := - time.Minute * time.Duration(minutes)
		t := time.Now().Add(duration)
		return t.Truncate(time.Minute).Unix(), typePlayed, true
	}
	if strings.HasSuffix(s, "刚刚") || strings.HasSuffix(s, "秒前") { //正在播放
		return time.Now().Unix(), typeNowPlaying, true
	}
	return 0, 0, false
}

func getAlbum(url string) (artist, album string, ok bool) {
	doc, err := getDoc(fmt.Sprintf("%s%s", domain, url))
	if err != nil || doc == nil {
		return "", "", false
	}

	var info []string
	doc.Find("#albums_info tr").Each(func(i int, s *goquery.Selection) {
		// For each item found, get the band and title
		title, _ := s.Find("a").Attr("title")
		info = append(info, title)
	})

	return info[1], info[0], true
}

func getDoc(url string) (*goquery.Document, error) {
	log.Println("xiami: getDoc url: ", url)
	res, err := http.Get(url)

	if err != nil {
		log.Println("Fatal: ", err)
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Printf("xiami: Fatal: status code error: %d %s on %s\n", res.StatusCode, res.Status, url)
		return nil, err
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Println("xiami: Fatal: ", err)
		return nil, err
	}
	return doc, nil
}
