package requestpolicy

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		desc           string
		whitelistPaths []PathConfig
		blacklistPaths []PathConfig
		expErr         bool
	}{
		{
			desc: "should return no error",
			whitelistPaths: []PathConfig{
				{
					PathRegex:       "^/$",
					QueryParamRegex: "(category)=oldest*",
				},
			},
			expErr: false,
		},
		{
			desc: "should return an error",
			blacklistPaths: []PathConfig{
				{
					PathRegex:       "*",
					QueryParamRegex: "*",
				},
			},
			expErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			conf := &Config{
				WhitelistPaths: test.whitelistPaths,
				BlacklistPaths: test.blacklistPaths,
			}

			_, err := New(context.Background(), nil, conf, "requestpolicy")
			if (err != nil) != test.expErr {
				t.Errorf("New() error = %v, wantErr %v", err, test.expErr)
			}
		})
	}
}

func TestServeHTTP(t *testing.T) {
	tests := []struct {
		desc           string
		whitelistPaths []PathConfig
		blacklistPaths []PathConfig
		reqPath        string
		reqQuery       string
		expNextCall    bool
		expStatusCode  int
	}{
		{
			desc: "Should return ok status",
			whitelistPaths: []PathConfig{
				{
					PathRegex: "^/$",
				},
			},
			reqPath:       "/test",
			expNextCall:   true,
			expStatusCode: http.StatusOK,
		},
		{
			desc: "Should return forbidden status with path",
			blacklistPaths: []PathConfig{
				{
					PathRegex: "/test",
				},
			},
			reqPath:       "/test",
			expNextCall:   false,
			expStatusCode: http.StatusForbidden,
		},
		{
			desc: "Should return forbidden status with query param",
			blacklistPaths: []PathConfig{
				{
					PathRegex:       "^/$",
					QueryParamRegex: "(category|limit|sort|page)=.*",
				},
			},
			reqPath:       "/",
			reqQuery:      "page=1&KJLhklhddopi",
			expNextCall:   false,
			expStatusCode: http.StatusForbidden,
		},
		{
			desc: "Should return ok status without query param",
			whitelistPaths: []PathConfig{
				{
					PathRegex: "^/test$",
				},
			},
			blacklistPaths: []PathConfig{
				{
					PathRegex:       "^/test$",
					QueryParamRegex: "(category|limit|sort|page)=.*",
				},
			},
			reqPath:       "/test",
			reqQuery:      "page=1&KJLhklhddopi",
			expNextCall:   true,
			expStatusCode: http.StatusOK,
		},
		{
			desc: "Should return ok status with query param",
			whitelistPaths: []PathConfig{
				{
					PathRegex:       "^/test$",
					QueryParamRegex: "(page)=.*",
				},
			},
			blacklistPaths: []PathConfig{
				{
					PathRegex:       "^/test$",
					QueryParamRegex: "(category|limit|sort|page)=.*",
				},
			},
			reqPath:       "/test",
			reqQuery:      "page=1",
			expNextCall:   true,
			expStatusCode: http.StatusOK,
		},
		{
			desc: "Should return ok status with any param",
			whitelistPaths: []PathConfig{
				{
					PathRegex:       "^/test$",
					QueryParamRegex: "(page)=.*",
				},
			},
			blacklistPaths: []PathConfig{
				{
					PathRegex:       "^/test$",
					QueryParamRegex: "(category|limit|sort|page)=.*",
				},
			},
			reqPath:       "/hello",
			reqQuery:      "sort=latest",
			expNextCall:   true,
			expStatusCode: http.StatusOK,
		},
		{
			desc: "Should return forbidden status with query param",
			whitelistPaths: []PathConfig{
				{
					PathRegex:       "^/test$",
					QueryParamRegex: "(page)=.*",
				},
			},
			blacklistPaths: []PathConfig{
				{
					PathRegex:       "^/test$",
					QueryParamRegex: "(category|limit|sort|page)=.*",
				},
			},
			reqPath:       "/test",
			reqQuery:      "category=1&KJLhklhddopi",
			expNextCall:   false,
			expStatusCode: http.StatusForbidden,
		},
		{
			desc: "Should return ok status with multiple matching rules in whitelistPaths",
			whitelistPaths: []PathConfig{
				{
					PathRegex:       "^/test$",
					QueryParamRegex: "(page)=.*",
				},
				{
					PathRegex:       "^/test$",
					QueryParamRegex: "(sort)=.*",
				},
			},
			reqPath:       "/test",
			reqQuery:      "page=1&sort=latest",
			expNextCall:   true,
			expStatusCode: http.StatusOK,
		},
		{
			desc: "Should return ok status with no matching rules in whitelistPaths and blacklistPaths",
			whitelistPaths: []PathConfig{
				{
					PathRegex:       "^/test$",
					QueryParamRegex: "(page)=.*",
				},
			},
			blacklistPaths: []PathConfig{
				{
					PathRegex:       "^/hello$",
					QueryParamRegex: "(category)=.*",
				},
			},
			reqPath:       "/other",
			reqQuery:      "limit=10",
			expNextCall:   true,
			expStatusCode: http.StatusOK,
		},
		{
			desc: "Should return ok status with matching rules in both whitelistPaths and blacklistPaths",
			whitelistPaths: []PathConfig{
				{
					PathRegex:       "^/test$",
					QueryParamRegex: "(page)=.*",
				},
			},
			blacklistPaths: []PathConfig{
				{
					PathRegex:       "^/test$",
					QueryParamRegex: "(category)=.*",
				},
			},
			reqPath:       "/test",
			reqQuery:      "page=1&category=latest",
			expNextCall:   true,
			expStatusCode: http.StatusOK,
		},
		{
			desc: "Should return ok status with empty regex in whitelistPaths",
			whitelistPaths: []PathConfig{
				{
					PathRegex:       "",
					QueryParamRegex: "",
				},
			},
			reqPath:       "/test",
			reqQuery:      "page=1",
			expNextCall:   true,
			expStatusCode: http.StatusOK,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			cfg := &Config{
				WhitelistPaths: test.whitelistPaths,
				BlacklistPaths: test.blacklistPaths,
			}

			nextCall := false
			next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				nextCall = true
			})

			handler, err := New(context.Background(), next, cfg, "requestpolicy")
			if err != nil {
				t.Fatal(err)
			}

			recorder := httptest.NewRecorder()

			req := httptest.NewRequest(http.MethodGet, test.reqPath, nil)
			req.URL.RawQuery = test.reqQuery

			handler.ServeHTTP(recorder, req)

			if nextCall != test.expNextCall {
				t.Errorf("next handler should not be called")
			}

			if recorder.Result().StatusCode != test.expStatusCode {
				t.Errorf("%s: got status code %d, want %d", test.desc, recorder.Code, test.expStatusCode)
			}
		})
	}
}
