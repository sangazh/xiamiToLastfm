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
	domain, spm string
	userId      int
)

type Track struct {
	Title, Artist, Album string
	Timestamp            int64
}

// Get xiami user id from user's profile page url
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
	userId = viper.GetInt("xiami.user_id")
	spm = viper.GetString("xiami.spm")
	if userId > 0 && spm != "" {
		return true
	}
	return false
}

// parseUrl and save to config
func parseUrl(rawUrl string) (userId, spm string) {
	u, _ := url.Parse(rawUrl)
	spm = u.Query().Get("spm")
	spm = strings.TrimRight(spm, "\n")

	s := strings.Split(u.Path, "/")
	userId = s[2]

	viper.Set("xiami.user_id", userId)
	viper.Set("xiami.spm", spm)
	viper.WriteConfig()

	return userId, spm
}

// Access user's recent track page.
func Tracks(playingChan, playedChan chan Track) error {
	recentUri := viper.GetString("xiami.url.recent")
	lastCheckAt := viper.GetInt64("xiami.checked_at")

	recentUrl, _ := url.Parse(domain)
	recentUrl.Path += recentUri + fmt.Sprint(userId)

	doc, err := document(recentUrl)
	if err != nil || doc == nil {
		log.Println("xiami.Tracks:", err)
		return err
	}

	// Find the track list
	doc.Find(".track_list tr").Each(func(i int, s *goquery.Selection) {
		trackTime := s.Find(".track_time").Text()
		timeStamp, scrobbleType, ok := parseTime(trackTime)

		// compare with last record check time, if before, means scrobbles are up to date.
		if !ok || timeStamp < lastCheckAt {
			return
		}

		title, _ := s.Find(".song_name a").Attr("title")
		trackUrl, _ := s.Find(".song_name a").Attr("href")

		//find record's artist and album from its' detail page, as artist and album are required.
		artist, album, ok := album(trackUrl)
		time.Sleep(2 * time.Second)

		if !ok {
			log.Println("xiami.album fetch failed")
			return
		}

		t := Track{Title: title, Artist: artist, Album: album, Timestamp: timeStamp}

		switch scrobbleType {
		case typeNowPlaying:
			playingChan <- t
			fmt.Printf("nowPlaying: %s - %s 《%s》 \n", title, artist, album)
			log.Println("xiami.Tracks: playingChan <- t ", t)
		case typePlayed:
			playedChan <- t
			fmt.Printf("Listened: %d: %s - %s 《%s》 \n", i, title, artist, album)
			log.Println("xiami.Tracks: playedChan <- t ", t)
		default:
			log.Println("xiami.Tracks: switch default")
		}
	})
	log.Println("xiami.Tracks returned.")
	return nil
}

// if time before 1 hour, then exact time cannot be calculated, abort
func parseTime(s string) (t int64, srbType int, ok bool) {
	//播放完毕可以同步
	if strings.HasSuffix(s, "分钟前") {
		minutes, _ := strconv.Atoi(strings.TrimSuffix(s, "分钟前"))
		duration := - time.Minute * time.Duration(minutes)
		t := time.Now().Add(duration)
		return t.Truncate(time.Minute).Unix(), typePlayed, true
	}

	//正在播放
	if strings.HasSuffix(s, "刚刚") || strings.HasSuffix(s, "秒前") {
		return time.Now().Unix(), typeNowPlaying, true
	}
	return 0, 0, false
}

func album(uri string) (artist, album string, ok bool) {
	songUrl, _ := url.Parse(domain)
	songUrl.Path += uri

	doc, err := document(songUrl)
	if err != nil {
		log.Println("xiami.album:", err)
		return "", "", false
	}
	if doc == nil {
		return "", "", false
	}

	var info []string
	doc.Find("#albums_info tr").Each(func(i int, s *goquery.Selection) {
		// For each item found, get the artist and title
		title, _ := s.Find("a").Attr("title")
		info = append(info, title)
	})

	if len(info) < 2 {
		return "", "", false
	}

	return info[1], info[0], true
}

func document(u *url.URL) (*goquery.Document, error) {
	v := url.Values{}
	v.Set("spm", spm)
	u.RawQuery = v.Encode()

	log.Println("xiami.document url:", u)
	res, err := http.Get(u.String())

	if err != nil {
		log.Println("xiami.document Fatal: ", err)
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("status code error: '%s' on %s", res.Status, u)
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}
	return doc, nil
}
