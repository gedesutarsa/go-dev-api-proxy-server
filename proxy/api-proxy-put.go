package proxy

import (
	"fmt"
	"net/http"
)

//PutAPIProxy proxy for PUT request
func PutAPIProxy(w http.ResponseWriter, r *http.Request) {
	url := GenerateAPIUrl(r)
	req, _ := http.NewRequest("PUT", url, r.Body)
	CopyToHeader(r.Header, req.Header)
	fmt.Printf("Sending [PUT]request to url: %s", url)
	invokeAPIRequest(req, &w)
}
