package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/gedesutarsa/go-dev-api-proxy-server/proxy"
	"github.com/gorilla/mux"
)

func main() {

	r := mux.NewRouter()
	jsonPAth := os.Getenv("jsonFilePath")
	dat, errPath := ioutil.ReadFile(jsonPAth)
	if errPath != nil {
		panic(errPath.Error())
	}
	var cnf proxy.Config
	err := json.Unmarshal(dat, &cnf)
	if err != nil {
		panic(err.Error())
	}
	proxy.APIEndpointURL = cnf.APIURL
	fmt.Printf("Url def: %s", proxy.APIEndpointURL)
	mgr := proxy.GatewayHandler{LazyCacheDefinitions: cnf.LazyCacheDefinitions, PreFetchPathDefinitions: cnf.PreFetchPathDefinitions, CachedPostRequestPaths: cnf.CachedPostRequestPaths}

	mgr.RegisterServeFromFileConfig(cnf.ServeWithFiles)
	mgr.Initialize()
	r.PathPrefix("/").HandlerFunc(mgr.HandleHTTPRequest)
	port := os.Getenv("port")
	if len(port) == 0 {
		port = "8080"
	}
	http.ListenAndServe(fmt.Sprintf(":%s", port), r)

}

/*
func allInHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(fmt.Sprintf("OK, method: %s", r.Method)))
}

func delHandler(w http.ResponseWriter, r *http.Request) {

	w.Write([]byte("DEL ok"))
}
func putHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("PUT ok"))
}
func postHandler(w http.ResponseWriter, r *http.Request) {

	w.Write([]byte("POST ok"))
}
func getHandler(w http.ResponseWriter, r *http.Request) {
	reqPath := r.URL.Path
	if strings.HasPrefix(reqPath, "/backoffice/api/") {
		proxy.GetAPIProxy(w, r)
		return
	}
	fmt.Printf("Handling path: %s, qs :%s", r.URL.Path, r.URL.RawQuery)
	w.Write([]byte("GET ok"))

}

type permanentCacheDefinition struct {
}

func postHandlerExplorer(w http.ResponseWriter, r *http.Request) {
	bodyString, _ := ioutil.ReadAll(r.Body)
	w.Write([]byte(fmt.Sprintf("POST ok, body : %s", bodyString)))
}

*/
