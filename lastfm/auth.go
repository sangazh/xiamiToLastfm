package lastfm

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"time"

	"github.com/theherk/viper"
)

var (
	token, sk    string
	tokenExpired int64
)

func Auth() {
	tokenOk, skOk := checkAuth()

	if skOk {
		return
	}

	if !tokenOk {
		fmt.Println("Fetching last.fm token...")
		if !getToken() {
			log.Fatal("last.fm token fetch failed")
		}
	}

	fmt.Println("Please open the link below, and grant the permission.")
	fmt.Println(authPage())

	fmt.Println("Press 'Enter' to continue...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')

	//get session and save to the config file
	getSession()
	return
}

func checkAuth() (tokenOk, skOk bool) {
	domain = viper.GetString("lastfm.domain")
	apiKey = viper.GetString("lastfm.api_key")
	sharedSecret = viper.GetString("lastfm.shared_secret")
	apiUrl = domain + "/2.0"

	token = viper.GetString("lastfm.auth.token")
	tokenExpired = viper.GetInt64("lastfm.auth.token_expired")
	sk = viper.GetString("lastfm.auth.sk")

	if token != "" && tokenExpired > time.Now().Unix() {
		tokenOk = true
		log.Println("last.fm: valid token found.")
	}

	if sk != "" {
		skOk = true
		log.Println("last.fm: session key found")
	}

	return
}

// token valid for 60 minutes
func getToken() (ok bool) {
	url := fmt.Sprintf("%s/?method=auth.gettoken&api_key=%s&format=json", apiUrl, apiKey)

	resp, ok := getRequest(url)
	if !ok {
		return false
	}

	result := toMap(resp)

	token, ok = result["token"]
	if !ok {
		return false
	}

	viper.Set("lastfm.auth.token", token)
	viper.Set("lastfm.auth.token_expired", time.Now().Add(60 * time.Minute).Unix())
	viper.WriteConfig()

	return true
}

//generate signature
func signature(query map[string]interface{}) (sig string) {
	query["api_key"] = apiKey
	ordered := prepareSigText(query)
	log.Println("signature - ordered query string ", ordered)
	text := fmt.Sprintf("%s%s", ordered, sharedSecret)
	log.Println("signature - before md5 ", text)
	data := []byte(text)
	hashed := md5.Sum(data)
	return hex.EncodeToString(hashed[:])
}

//sort query first, then return string with format of <key><value>
func prepareSigText(query map[string]interface{}) (text string) {
	// sort query
	var mapKey []string
	for key := range query {
		mapKey = append(mapKey, key)
	}

	sort.Strings(mapKey)
	for _, key := range mapKey {
		text += fmt.Sprintf("%s%s", key, query[key])
	}

	return text
}

func authPage() string {
	return fmt.Sprintf("http://www.last.fm/api/auth/?api_key=%s&token=%s", apiKey, token)
}

// Session keys have an infinite lifetime by default
func getSession() {
	query := map[string]interface{}{
		"method":  "auth.getSession",
		"api_key": apiKey,
		"token":   token,
	}

	query["api_sig"] = signature(query)
	url := fmt.Sprintf("%s/?%s&format=json", apiUrl, queryString(query))
	resp, ok := getRequest(url)
	if !ok {
		log.Fatal("last.fm getSession Failed")
	}

	key, ok := parseKey(resp)
	if !ok {
		log.Fatal("last.fm: parse session key Failed")
	}

	log.Println("last.fm: session Key: ", key)

	viper.Set("lastfm.auth.token", "")
	viper.Set("lastfm.auth.token_expired", 0)
	viper.Set("lastfm.auth.sk", key)
	viper.WriteConfig()
	return
}

type SessionResp struct {
	Session struct {
		Name string `json:"name"`
		Key  string `json:"key"`
	} `json:"session"`
}

func parseKey(b []byte) (string, bool) {
	var s SessionResp
	json.Unmarshal(b, &s)
	if s.Session.Key == "" {
		return "", false
	}
	return s.Session.Key, true
}

func resetAuth() {
	viper.Set("lastfm.auth.token", "")
	viper.Set("lastfm.auth.token_expired", 0)
	viper.Set("lastfm.auth.sk", "")
	viper.WriteConfig()
}
