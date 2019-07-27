package proxy

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
)

//RequestResult request result
type RequestResult struct {
	//ResponseHeader response header for response to request
	ResponseHeader http.Header
	//ResponseBody response from original. this will be used as response to request on proxy
	ResponseBody []byte
}

//GatewayHandler struct handler gateway
type GatewayHandler struct {
	//cachedRequest cached request
	cachedRequest map[string]RequestResult
	//cachedPostRequest post request that cached. if same path is requested for post, then request will not continued to proxy
	cachedPostRequest map[string]RequestResult
	//LazyCacheDefinitions lazy cache definitions
	LazyCacheDefinitions []LazyCacheDefinition `json:"lazyCacheDefinitions"`
	//PreFetchPathDefinitions prefetched cache definitions. load on startup
	PreFetchPathDefinitions []PreFetchPathDefinition `json:"preFetchPathDefinitions"`
	//CachedPostRequestPaths post request that will be cached.
	CachedPostRequestPaths []string `json:"cachedPostRequestPaths"`
	//serveFromFileExactMatch handler from file with exact match rule
	serveFromFileExactMatch []ServeWithFileDefinition
	//serveFromFileStartWithPattern path start for handle from file
	serveFromFileStartWithPattern []ServeWithFileDefinition
}

//RegisterServeFromFileConfig register handler from file. also split the dta
func (p *GatewayHandler) RegisterServeFromFileConfig(param []ServeWithFileDefinitionPlain) {
	if len(param) == 0 {
		return
	}
	for _, p2 := range param {
		if len(p2.ExactPattern) > 0 {
			p.serveFromFileExactMatch = append(p.serveFromFileExactMatch, ServeWithFileDefinition{ContentType: p2.ContentType,
				ExactPattern: p2.ExactPattern,
				FilePath:     p2.FilePath})
		} else if len(p2.StartPatterns) > 0 {
			p.serveFromFileStartWithPattern = append(p.serveFromFileStartWithPattern, ServeWithFileDefinition{ContentType: p2.ContentType,
				StartPatterns: p2.StartPatterns,
				FilePath:      p2.FilePath})
		}
	}
}

//HandleHTTPRequest handler http request
func (p *GatewayHandler) HandleHTTPRequest(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case "DELETE":
		DelAPIProxy(w, r)
	case "PUT":
		PutAPIProxy(w, r)
	case "POST":
		p.handlePostRequest(w, r)
	case "GET":
		p.handleGetRequest(w, r)
	default:
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte(fmt.Sprintf("Unhandled method: %s", r.Method)))
	}
}

//handlePostRequest handle post request. if request is cache able, then request will be requested once and next request will be use prev result data
func (p *GatewayHandler) handlePostRequest(w http.ResponseWriter, r *http.Request) {
	reqPath := r.URL.Path
	for _, pth := range p.CachedPostRequestPaths {
		if pth == reqPath {
			if cacheData, ok := p.cachedPostRequest[reqPath]; ok {
				p.responseWithCache(w, cacheData)
				return
			}
			statusCode, respHeader, body, errReq := sendPostRequest(r)
			if errReq != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(errReq.Error()))
				return
			}
			if statusCode == 200 {
				p.cachedPostRequest[reqPath] = RequestResult{ResponseHeader: respHeader, ResponseBody: body}
			}
			CopyToHeader(respHeader, w.Header())
			w.WriteHeader(statusCode)
			w.Write(body)

			return
		}
	}
	PostAPIProxy(w, r)
}

//responseWithCache writer from cache to current request response
func (p *GatewayHandler) responseWithCache(w http.ResponseWriter, reqResult RequestResult) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in f", r)
		}
	}()
	h := w.Header()
	for key1, v1 := range reqResult.ResponseHeader {
		for _, v := range v1 {
			h.Add(key1, v)
		}
	}
	w.Write(reqResult.ResponseBody)

}

//handleGetRequest handler get request. this is proxy able
func (p *GatewayHandler) handleGetRequest(w http.ResponseWriter, r *http.Request) {
	reqPath := r.URL.Path
	if rslt, ok := p.cachedRequest[reqPath]; ok { // already cached
		fmt.Printf("[serve-with-cache] %s", reqPath)
		p.responseWithCache(w, rslt)
		return
	}

	for _, cacheDef := range p.LazyCacheDefinitions {
		fullFilled, _ := cacheDef.checkIsPathMatched(reqPath)
		if fullFilled {
			statusCode, respHeader, body, errReq := sendGetRequest(GenerateAPIUrl(r), r.Header)
			if errReq != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(errReq.Error()))
				return
			}
			if statusCode == 200 {
				p.cachedRequest[r.URL.Path] = RequestResult{ResponseHeader: respHeader, ResponseBody: body}
			}
			CopyToHeader(respHeader, w.Header())
			w.WriteHeader(statusCode)
			w.Write(body)
		}
	}
	for _, hExact := range p.serveFromFileExactMatch {
		if hExact.CheckIsPathServed(reqPath) {
			hExact.handleRequest(w, r)
			return
		}
	}
	for _, hStartwith := range p.serveFromFileStartWithPattern {
		if hStartwith.CheckIsPathServed(reqPath) {
			hStartwith.handleRequest(w, r)
			return
		}
	}
	// ok then send no proxy request
	GetAPIProxy(w, r)
}

//Initialize initialize manager
func (p *GatewayHandler) Initialize() (err error) {
	p.cachedRequest = make(map[string]RequestResult)
	p.cachedPostRequest = make(map[string]RequestResult)
	if len(p.PreFetchPathDefinitions) > 0 {
		for _, def := range p.PreFetchPathDefinitions {
			errCurrent := p.requestPredefinedCache(def)
			if errCurrent != nil {
				err = errCurrent
			}
		}
		return
	}
	return
}

//RequestPredefinedCache request predefined cache. run on startup
func (p *GatewayHandler) requestPredefinedCaches(logEntry *logrus.Entry, definitions ...PreFetchPathDefinition) (err error) {
	for _, def := range definitions {
		errCurrent := p.requestPredefinedCache(def)
		if errCurrent != nil {
			logEntry.WithError(errCurrent).WithField("cacheDefintion", def).Error(errCurrent.Error())
			err = errCurrent
		}
	}
	return
}

//requestPredefinedCache request predefined cache
func (p *GatewayHandler) requestPredefinedCache(definition PreFetchPathDefinition) (err error) {
	fmt.Printf("[requesting-cached] %s\n", definition.RequestPath)
	url := definition.SourcePath
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = GenerateFullURLByPath(definition.SourcePath)
	}
	statusCode, responseHeader, body, err1 := sendGetRequest(url, http.Header{})
	if err1 != nil {
		err = err1
		return
	}
	if statusCode == 200 {
		p.cachedRequest[definition.RequestPath] = RequestResult{ResponseHeader: responseHeader, ResponseBody: body}
		fmt.Printf("[cached] %s\n", definition.RequestPath)
	}
	return
}
