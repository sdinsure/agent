package metadata

import (
	"context"
	"net/http"
	"strings"

	"google.golang.org/grpc/metadata"
)

var (
	cookieAuthorization string = "authorization"
)

func HttpCookiesToGrpcMetadata(ctx context.Context, r *http.Request) metadata.MD {
	md := make(map[string]string)

	// add cookies support
	// convert cookies: Authorization=$value
	// into md['Authorization']=$value
	cookies := r.Cookies()
	for _, cookie := range cookies {
		if strings.ToLower(cookie.Name) == cookieAuthorization {
			md[cookieAuthorization] = cookie.Value
		}
	}
	return metadata.New(md)
}
