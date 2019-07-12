package proxy

import (
	"fmt"
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
