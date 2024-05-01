package openapi

import (
	"net/http"
	"net/url"
	"strings"

	httptransport "github.com/go-openapi/runtime/client"
	sdinsurehttp "github.com/sdinsure/agent/pkg/http"
)

func MustNew(endpoint string, basePath string, client *http.Client) *httptransport.Runtime {
	tr, err := New(endpoint, basePath, client)
	if err != nil {
		panic(err)
	}
	return tr
}

func New(endpoint string, basePath string, client *http.Client) (*httptransport.Runtime, error) {
	if client == nil {
		client = sdinsurehttp.NewHttpClient()
	}

	defaultScheme := "http"
	defaultHost := endpoint
	if hasScheme(endpoint) {
		hostUrl, err := url.Parse(endpoint)
		if err != nil {
			return nil, err
		}
		if len(hostUrl.Scheme) > 0 {
			defaultScheme = hostUrl.Scheme
		}
		defaultHost = hostUrl.Host
	}

	httptransportclient := httptransport.NewWithClient(
		defaultHost,
		basePath,
		[]string{defaultScheme},
		client,
	)
	return httptransportclient, nil
}

func hasScheme(endpoint string) bool {
	if strings.HasPrefix(endpoint, "http://") || strings.HasPrefix(endpoint, "https://") {
		return true
	}
	return false
}
