package proxy

import (
	"fmt"
	"regexp"
	"strings"
)

type chekerPath func(requestPath string) (isFullfiled bool, err error)

//LazyCacheDefinition defintion for lazy request cache
type LazyCacheDefinition struct {
	//PathRegexPattern priority 1 checked on regex
	PathRegexPattern string `json:"pathRegexPattern"`
	//PathStartWithPattern priority 2 , path start with pattern
	PathStartWithPattern string `json:"pathStartWithPattern"`
	//PathEndWithPattern priority 3, if start with not zero length can be combined with this
	PathEndWithPattern string `json:"pathEndWithPattern"`

	//pattern: 1. start with , 2. start with, and end with, 3. regex
	//CachedRequest cached request
	CachedRequest map[string]RequestResult `json:"-"`
	//initialized flag struct initialized
	initialized bool
	//RequestPathChecker checker is path is match with rule
	RequestPathChecker chekerPath `json:"-"`
}

//checkIsPathMatched check is rule match
func (p *LazyCacheDefinition) checkIsPathMatched(requestPath string) (isFullfiled bool, err error) {
	if !p.initialized {
		err = p.initialize()
		if err != nil {
			fmt.Printf("Unable to initialize checker, error: %s", err.Error())
		}
	}
	return p.RequestPathChecker(requestPath)
}

//initialize initialize cache definition.
func (p *LazyCacheDefinition) initialize() (err error) {
	if len(p.PathRegexPattern) > 0 {
		var regexPathChecker *regexp.Regexp
		regexPathChecker, err = regexp.Compile(p.PathRegexPattern)
		p.RequestPathChecker = func(requestPath string) (isFullfiled bool, err error) {
			isFullfiled = regexPathChecker.Match([]byte(requestPath))
			return
		}
	} else if len(p.PathStartWithPattern) > 0 {
		if len(p.PathEndWithPattern) > 0 {
			p.RequestPathChecker = func(requestPath string) (isFullfiled bool, err error) {
				isFullfiled = strings.HasPrefix(requestPath, p.PathStartWithPattern) && strings.HasSuffix(requestPath, p.PathEndWithPattern)
				return
			}
		} else {
			p.RequestPathChecker = func(requestPath string) (isFullfiled bool, err error) {
				isFullfiled = strings.HasPrefix(requestPath, p.PathStartWithPattern)
				return
			}
		}

	} else if len(p.PathEndWithPattern) > 0 {
		p.RequestPathChecker = func(requestPath string) (isFullfiled bool, err error) {
			isFullfiled = strings.HasSuffix(requestPath, p.PathEndWithPattern)
			return
		}
	} else {
		err = fmt.Errorf("At least 1  of 3 parametrers (PathRegexPattern, PathStartWithPattern ,PathEndWithPattern ) should be filled")
		return
	}
	p.initialized = true
	return
}
