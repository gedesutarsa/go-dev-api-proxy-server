package proxy

//Config json conveter for config
type Config struct {
	//APIURL url handle API
	APIURL string `json:"apiUrl"`
	//LazyCacheDefinitions lazy cache definitions
	LazyCacheDefinitions []LazyCacheDefinition `json:"lazyCacheDefinitions"`
	//PreFetchPathDefinitions preferched config
	PreFetchPathDefinitions []PreFetchPathDefinition `json:"preFetchPathDefinitions"`
	//ServeWithFiles definition for path that will handle with local file
	ServeWithFiles []ServeWithFileDefinitionPlain `json:"serveWithFiles"`
	//CachedPostRequestPaths post request that will be cached.
	CachedPostRequestPaths []string `json:"cachedPostRequestPaths"`
}
