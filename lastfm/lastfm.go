package lastfm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

// QuitChan: an empty channel used to signal main channel to stop.
var (
	domain, apiUrl, sharedSecret, apiKey string
	QuitChan                             chan struct{}
)

func getRequest(url string) ([]byte, error) {
	log.Println("last.fm: getRequest url: ", url)
	res, err := http.Get(url)

	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	resData, _ := ioutil.ReadAll(res.Body)
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("status code error: '%s' on %s body: %s", res.Status, url, string(resData))
	}

	log.Println("last.fm: getRequest response: ", string(resData))
	return resData, nil
}

func postRequest(query string) ([]byte, error) {
	r := bytes.NewReader([]byte(query))
	contentType := "application/x-www-form-urlencoded"

	log.Println("last.fm: postRequest params: ", query)
	res, err := http.Post(apiUrl, contentType, r)

	if err != nil {
		return nil, fmt.Errorf("postRequest has error: %s", err)
	}

	resData, _ := ioutil.ReadAll(res.Body)
	defer res.Body.Close()

	if res.StatusCode != 200 {
		errCode, errMsg := handleError(resData)
		if errCode == 9 {
			fmt.Println(errMsg)
			resetAuth()
			fmt.Println("Config reset. Please re-start the program.")
			close(QuitChan)
			os.Exit(1)
		}
		return nil, fmt.Errorf("status code error: '%s' on %s body: %s", res.Status, apiUrl, string(resData))

	}
	log.Println("last.fm: postRequest response: ", string(resData))
	return resData, nil
}

func toMap(byteData []byte) (result map[string]string) {
	r := bytes.NewReader(byteData)
	json.NewDecoder(r).Decode(&result)
	return result
}

func handleError(errData []byte) (code int, msg string) {
	type ErrResponse struct {
		Code int    `json:"error"`
		Msg  string `json:"message"`
	}

	var resp ErrResponse
	json.Unmarshal(errData, &resp)
	return resp.Code, resp.Msg
}
