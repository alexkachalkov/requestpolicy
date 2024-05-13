package requestpolicy

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
)

type PathConfig struct {
	PathRegex       string `yaml:"pathRegex"`
	QueryParamRegex string `yaml:"queryParamRegex"`
}

func CreateConfig() *Config {
	return &Config{}
}

type Config struct {
	WhitelistPaths []PathConfig `yaml:"whitelistPaths"`
	BlacklistPaths []PathConfig `yaml:"blacklistPaths"`
}

type Middleware struct {
	next   http.Handler
	name   string
	config *Config
}

func New(ctx context.Context, next http.Handler, conf *Config, name string) (http.Handler, error) {
	for _, pathConfig := range append(conf.WhitelistPaths, conf.BlacklistPaths...) {
		if _, err := regexp.Compile(pathConfig.PathRegex); err != nil {
			return nil, fmt.Errorf("invalid path regex: %w", err)
		}
		if _, err := regexp.Compile(pathConfig.QueryParamRegex); err != nil {
			return nil, fmt.Errorf("invalid query param regex: %w", err)
		}
	}

	return &Middleware{
		next:   next,
		name:   name,
		config: conf,
	}, nil
}

func (m *Middleware) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	path := req.URL.Path
	query := req.URL.RawQuery
	for _, pathConfig := range m.config.WhitelistPaths {
		matchPath, _ := regexp.MatchString(pathConfig.PathRegex, path)
		matchQuery := matchQuery(pathConfig, query)
		if matchPath && matchQuery {
			m.next.ServeHTTP(rw, req)
			return
		}
	}

	for _, pathConfig := range m.config.BlacklistPaths {
		matchPath, _ := regexp.MatchString(pathConfig.PathRegex, path)
		matchQuery := matchQuery(pathConfig, query)
		if matchPath && matchQuery {
			http.Error(rw, "Forbidden", http.StatusForbidden)
			return
		}
	}

	m.next.ServeHTTP(rw, req)
}

func matchQuery(pathConfig PathConfig, query string) bool {
	if pathConfig.QueryParamRegex == "" {
		return true
	} else {
		match, _ := regexp.MatchString(pathConfig.QueryParamRegex, query)
		return match
	}
}
