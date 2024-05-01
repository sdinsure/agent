package gohttpclient

import (
	"net/http"

	helloserviceopenapi "github.com/sdinsure/agent/example/api/go-openapiv2/client"
	helloserviceclient "github.com/sdinsure/agent/example/api/go-openapiv2/client/hello_service"
	"github.com/sdinsure/agent/pkg/http/openapi"
)

func MustNewClient(endpoint string, client *http.Client) helloserviceclient.ClientService {
	return helloserviceopenapi.New(
		openapi.MustNew(
			endpoint,
			helloserviceopenapi.DefaultBasePath,
			client,
		), nil).HelloService
}
