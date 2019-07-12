package proxy

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

//APIEndpointURL endpoint url actual
var APIEndpointURL string

//CopyToHeader copy header to header
func CopyToHeader(source http.Header, destination http.Header) {
	if source == nil || destination == nil {
		return
	}
	for key, val := range source {
		if val == nil || len(val) == 0 {
			continue
		}
		for _, v := range val {
			destination.Add(key, v)
		}
	}
}

//GenerateAPIUrl generate URL API
func GenerateAPIUrl(r *http.Request) (apiURL string) {
	url := APIEndpointURL + r.URL.Path
	qs := r.URL.RawQuery
	if len(qs) > 0 {
		url = url + "?" + qs
	}
	return url
}

//GenerateFullURLByPath generate full url by path
func GenerateFullURLByPath(path string) (fullURL string) {
	rtvl := APIEndpointURL
	if (!strings.HasSuffix(rtvl, "/") && strings.HasPrefix(path, "/")) || (strings.HasPrefix(rtvl, "/") && !strings.HasPrefix(path, "/")) {
		return fmt.Sprintf("%s/%s", rtvl, path)
	}

	return rtvl
}

//invokeAPIRequest send API request and send result to client
func invokeAPIRequest(req *http.Request, targetResponse *http.ResponseWriter) {
	w := (*targetResponse)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	if res.StatusCode != 200 {
		w.WriteHeader(res.StatusCode)
	}
	CopyToHeader(res.Header, w.Header())
	body, _ := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	w.Write(body)
}

//PreFetchPathDefinition path that pre fetch. fetch when app Up
type PreFetchPathDefinition struct {
	//RequestPath request path of file request(no url)
	RequestPath string `json:"requestPath"`
	//SourcePath url for source file for download. if start with http:// or https:// means url is absolute, else will be appended with base path
	SourcePath string `json:"sourcePath"`
	//ResponseHeader response header for response to request
	ResponseHeader http.Header `json:"-"`
	//ResponseBody response from original. this will be used as response to request on proxy
	ResponseBody []byte `json:"-"`
}

//ServeWithFileDefinitionPlain only data
type ServeWithFileDefinitionPlain struct {
	//ContentType file content type
	ContentType string `json:"contentType"`
	//FilePath file to handle the request
	FilePath string `json:"filePath"`
	//StartPatterns pattern awal dari path di handle dengan file handler
	StartPatterns []string `json:"startPatterns"`
	//ExactPattern exact match pattern
	ExactPattern []string `json:"exactPattern"`
}

//ServeWithFileDefinition definition for path handler by local file. this to server all path with index.html for example on client side routing
type ServeWithFileDefinition struct {
	//ContentType file content type
	ContentType string `json:"contentType"`
	//FilePath file to handle the request
	FilePath string `json:"filePath"`
	//StartPatterns pattern awal dari path di handle dengan file handler
	StartPatterns []string `json:"startPatterns"`
	//ExactPattern exact match pattern
	ExactPattern []string `json:"exactPattern"`
}

//CheckIsPathServed check is path served by current definition
func (p *ServeWithFileDefinition) CheckIsPathServed(path string) (served bool) {
	if len(p.ExactPattern) > 0 {
		for _, pExct := range p.ExactPattern {
			if pExct == path {
				return true
			}
		}
	} else if len(p.StartPatterns) > 0 {
		for _, pStart := range p.StartPatterns {
			if strings.HasPrefix(path, pStart) {
				return true
			}
		}
	}
	return false
}

//handleRequest handle request with file from path
func (p *ServeWithFileDefinition) handleRequest(w http.ResponseWriter, r *http.Request) {
	reqPath := r.URL.Path
	fmt.Printf("[handle-with-file] path : %s  >> file %s\n", reqPath, p.FilePath)
	content, err := ioutil.ReadFile(p.FilePath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.Header().Add("Content-Type", p.ContentType)
	w.Write(content)
}
