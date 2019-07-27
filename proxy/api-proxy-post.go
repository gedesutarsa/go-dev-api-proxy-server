package proxy

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

//PostAPIProxy forwader POST
func PostAPIProxy(w http.ResponseWriter, r *http.Request) {
	url := GenerateAPIUrl(r)
	req, _ := http.NewRequest("POST", url, r.Body)
	CopyToHeader(r.Header, req.Header)
	fmt.Printf("Sending [POST]request to url: %s", url)
	invokeAPIRequest(req, &w)
}

//sendPostRequest sender post request
func sendPostRequest(r *http.Request) (statusCode int, responseHeader http.Header, body []byte, err error) {
	url := GenerateAPIUrl(r)

	var req *http.Request
	req, err = http.NewRequest("POST", url, r.Body)
	CopyToHeader(r.Header, req.Header)
	if err != nil {
		return
	}
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
