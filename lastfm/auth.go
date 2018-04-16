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
	apiUrl = domain + "/2.0"

	token = viper.GetString("lastfm.auth.token")
	tokenExpired = viper.GetInt64("lastfm.auth.token_expired")
	sk = viper.GetString("lastfm.auth.sk")
	sharedSecret = viper.GetString("lastfm.auth.shared_secret")
	apiKey = viper.GetString("lastfm.auth.api_key")

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
	requestUrl := fmt.Sprintf("%s/?method=auth.gettoken&api_key=%s&format=json", apiUrl, apiKey)
	resp, ok := getRequest(requestUrl)
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
func signature(v *url.Values) (sig string) {
	ordered := prepareSigText(*v)
	log.Println("signature - ordered query string ", ordered)
	text := ordered + sharedSecret
	log.Println("signature - before md5 ", text)
	data := []byte(text)
	hashed := md5.Sum(data)
	return hex.EncodeToString(hashed[:])
}

//sort query first, then return string with format of <key><value>
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
		//prefix := url.QueryEscape(k)
		for _, v := range vs {
			//buf.WriteString(prefix)
			//buf.WriteString(url.QueryEscape(v))
			buf.WriteString(k)
			buf.WriteString(v)
		}
	}
	return buf.String()
}

func authPage() string {
	return fmt.Sprintf("http://www.last.fm/api/auth/?api_key=%s&token=%s", apiKey, token)
}

// Session keys have an infinite lifetime by default
func getSession() {
	v := url.Values{}
	v.Set("method", "auth.getSession")
	v.Set("api_key", apiKey)
	v.Set("token", token)
	sig := signature(&v)
	v.Set("api_sig", sig)
	v.Set("format", "json")

	query, _ := url.QueryUnescape(v.Encode())
	requestUrl := fmt.Sprintf("%s/?%s", apiUrl, query)
	resp, ok := getRequest(requestUrl)
	if !ok {
		log.Fatal("last.fm getSession Failed")
		fmt.Println("last.fm getSession Failed, Please contact the author.")
	}

	key, ok := parseKey(resp)
	if !ok {
		log.Fatal("last.fm: parse session key Failed")
		fmt.Println("last.fm getSession Failed, Please contact the author.")
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
