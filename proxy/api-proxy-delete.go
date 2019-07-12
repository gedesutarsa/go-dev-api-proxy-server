package proxy

import (
	"fmt"
	"net/http"
)

//DelAPIProxy proxy API for del
func DelAPIProxy(w http.ResponseWriter, r *http.Request) {
	url := GenerateAPIUrl(r)
	req, _ := http.NewRequest("DELETE", url, nil)
	CopyToHeader(r.Header, req.Header)
	fmt.Printf("Sending [DELETE]request to url: %s", url)
	invokeAPIRequest(req, &w)

}
