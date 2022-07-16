package hou

import (
	"net/http"
)

var (
	httpClient = &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		},
	}
)
