package proxy

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

//GetAPIProxy proxy API untuk get request
func GetAPIProxy(w http.ResponseWriter, r *http.Request) {
	url := GenerateAPIUrl(r)
	req, _ := http.NewRequest("GET", url, nil)
	CopyToHeader(r.Header, req.Header)
	fmt.Printf("Sending [GET]request to url: %s\n", url)
	invokeAPIRequest(req, &w)
}

//sendGetRequest sender get request. this is designed to be work with cache. so data is returned raw, to be cache. after cache complete then result send to client
func sendGetRequest(url string, header http.Header) (statusCode int, responseHeader http.Header, body []byte, err error) {
	var req *http.Request
	req, err = http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}
	CopyToHeader(header, req.Header)
	res, err2 := http.DefaultClient.Do(req)
	if err2 != nil {
		err = err2
		return
	}
	responseHeader = res.Header
	statusCode = res.StatusCode
	body, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}
	defer res.Body.Close()
	return
}
