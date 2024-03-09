package runtime

import (
	"context"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc/metadata"
)

var (
	httpVerb        string = "http-verb"
	httpPath        string = "http-path"
	httpPathPattern string = "http-path-pattern"
	grpcMethod      string = "grpc-method"
	remoteAddr      string = "remote-addr"
)

func ForwardHttpToMetadata(ctx context.Context, r *http.Request) metadata.MD {
	md := make(map[string]string)
	md[httpVerb] = r.Method
	md[httpPath] = r.URL.Path
	md[remoteAddr] = r.RemoteAddr
	if method, ok := runtime.RPCMethod(ctx); ok {
		md[grpcMethod] = method
	}
	if pattern, ok := runtime.HTTPPathPattern(ctx); ok {
		md[httpPathPattern] = pattern
	}
	return metadata.New(md)
}

func HttpVerb(ctx context.Context) (string, bool) {
	return getMetaValueFromCtx(ctx, httpVerb)
}

func HttpPath(ctx context.Context) (string, bool) {
	return getMetaValueFromCtx(ctx, httpPath)
}

func HttpPathPattern(ctx context.Context) (string, bool) {
	return getMetaValueFromCtx(ctx, httpPathPattern)
}

func GrpcMethod(ctx context.Context) (string, bool) {
	return getMetaValueFromCtx(ctx, grpcMethod)
}

func RemoteAddr(ctx context.Context) (string, bool) {
	return getMetaValueFromCtx(ctx, remoteAddr)
}

func XForwardedFor(ctx context.Context) ([]string, bool) {
	// x-forwarded-for is recording forwarding ips of the requests from very beginning to the handler
	// the key 'x-forwarded-for' is default key assiged by the grpc-gateway framework
	return getMetaValuesFromCtx(ctx, "x-forwarded-for")
}

func UserAgent(ctx context.Context) (string, bool) {
	return getMetaValueFromCtx(ctx, "user-agent")
}

func ForwardedHost(ctx context.Context) (string, bool) {
	return getMetaValueFromCtx(ctx, "x-forwarded-host")
}

func getMetaValueFromCtx(ctx context.Context, key string) (string, bool) {
	md, exists := metadata.FromIncomingContext(ctx)
	if !exists {
		return "", false
	}
	// --->md:map[:authority:[127.0.0.1:8081] authorization:[Bearer <>] content-type:[application/grpc] grpc-accept-encoding:[gzip] grpc-method:[/app.kafeido.Kafeido/GetProject] grpcgateway-accept:[application/json] grpcgateway-authorization:[Bearer <>] grpcgateway-user-agent:[Go-http-client/2.0] http-path:[/v1/projects/1] http-path-pattern:[/v1/projects/{projectId}] http-verb:[GET] remote-addr:[127.0.0.6:42301] user-agent:[grpc-go/1.55.0] x-forwarded-for:[10.244.0.0, 127.0.0.6] x-forwarded-host:[dev01.example.com]]
	founds := md.Get(key)
	return founds[0], true
}

func getMetaValuesFromCtx(ctx context.Context, key string) ([]string, bool) {
	md, exists := metadata.FromIncomingContext(ctx)
	if !exists {
		return nil, false
	}
	founds := md.Get(key)
	return founds, true
}
