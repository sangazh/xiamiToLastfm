package lastfm

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"sort"
	"time"

	"github.com/spf13/viper"
)

var (
	token, sk    string
	tokenExpired int64
)

// Authentication Guide:
// https://www.last.fm/api/desktopauth
// Step 1. GetToken
// Step 2. Authentication by user, direct to user.
// Step 3. GetSession, with signature
func Auth() {
	tokenOk, skOk := checkAuth()

	if skOk {
		return
	}

	if !tokenOk {
		fmt.Println("Fetching last.fm token...")
		if err := getToken(); err != nil {
			log.Fatal("last.fm: token fetch failed, err: ", err)
		}
	}

	fmt.Println("Please open the link below, and grant the permission.")
	fmt.Println(authPage())

	fmt.Println("Press 'Enter' to continue...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')

	//get session and save to the config file
	if err := session(); err != nil {
		fmt.Println("last.fm session fetch failed, Please contact the author.")
		log.Fatal("last.fm: session err: ", err)
	}
	return
}

func checkAuth() (tokenOk, skOk bool) {
	domain = viper.GetString("lastfm.domain")
	apiUrl = domain + "/2.0"
	token = viper.GetString("lastfm.auth.token")
	apiKey = viper.GetString("lastfm.auth.api_key")
	tokenExpired = viper.GetInt64("lastfm.auth.token_expired")
	sharedSecret = viper.GetString("lastfm.auth.shared_secret")
	sk = viper.GetString("lastfm.auth.sk")

	if token != "" && tokenExpired > time.Now().Unix() {
		tokenOk = true
		log.Println("last.fm token found.")
	}

	if sk != "" {
		skOk = true
		log.Println("last.fm session key found")
	}

	return
}

// https://www.last.fm/api/show/auth.getToken
// token valid for 60 minutes
func getToken() error {
	requestUrl := fmt.Sprintf("%s/?method=auth.gettoken&api_key=%s&format=json", apiUrl, apiKey)
	resp, err := getRequest(requestUrl)

	if err != nil {
		return err
	}

	result := toMap(resp)

	ok := false
	token, ok = result["token"]
	if !ok {
		return fmt.Errorf("parseToken failed")
	}

	viper.Set("lastfm.auth.token", token)
	viper.Set("lastfm.auth.token_expired", time.Now().Add(60 * time.Minute).Unix())
	viper.WriteConfig()
	return nil
}

// generate signature
func signature(v *url.Values) (sig string) {
	ordered := prepareSigText(*v)
	text := ordered + sharedSecret
	log.Println("last.fm: signature - before md5 ", text)

	data := []byte(text)
	hashed := md5.Sum(data)
	return hex.EncodeToString(hashed[:])
}

// sort query first, then return string with format of <key><value>
func prepareSigText(v url.Values) (text string) {
	if v == nil {
		return ""
	}
	var buf bytes.Buffer
	keys := make([]string, 0, len(v))
	for k := range v {
		keys = append(keys, k)
	}

	sort.Strings(keys)
	for _, k := range keys {
		vs := v[k]
		for _, v := range vs {
			buf.WriteString(k)
			buf.WriteString(v)
		}
	}
	return buf.String()
}

func authPage() string {
	return fmt.Sprintf("http://www.last.fm/api/auth/?api_key=%s&token=%s", apiKey, token)
}

// https://www.last.fm/api/show/auth.getSession
// Session keys have an infinite lifetime by default
func session() error {
	v := url.Values{}
	v.Set("method", "auth.session")
	v.Set("api_key", apiKey)
	v.Set("token", token)
	sig := signature(&v)
	v.Set("api_sig", sig)
	v.Set("format", "json")

	requestUrl, _ := url.Parse(apiUrl)
	requestUrl.RawQuery = v.Encode()
	resp, err := getRequest(requestUrl.String())

	if err != nil {
		return err
	}

	key, ok := parseKey(resp)
	if !ok {
		return fmt.Errorf("parse session key failed")
	}

	log.Println("last.fm session Key:", key)

	viper.Set("lastfm.auth.token", "")
	viper.Set("lastfm.auth.token_expired", 0)
	viper.Set("lastfm.auth.sk", key)
	viper.WriteConfig()
	return nil
}

func parseKey(b []byte) (string, bool) {
	type sessionResponse struct {
		Session struct {
			Name string `json:"name"`
			Key  string `json:"key"`
		} `json:"session"`
	}
	var s sessionResponse

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
