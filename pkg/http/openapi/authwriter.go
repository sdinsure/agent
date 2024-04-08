package openapi

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"

	sdinsureerrors "github.com/sdinsure/agent/pkg/errors"
	sdinsureruntime "github.com/sdinsure/agent/pkg/grpc/server/runtime"
)

type ContextClientAuthInfoWriter struct {
	ctx context.Context
}

func NewContextClientAuthInfoWriter(ctx context.Context) *ContextClientAuthInfoWriter {
	return &ContextClientAuthInfoWriter{ctx}
}

var (
	_ runtime.ClientAuthInfoWriter = &ContextClientAuthInfoWriter{}
)

func (c *ContextClientAuthInfoWriter) AuthenticateRequest(clientRequest runtime.ClientRequest, registry strfmt.Registry) error {
	keyVal, foundKey := sdinsureruntime.KeyInfo(c.ctx)
	if !foundKey {
		return sdinsureerrors.NewInvalidAuth(errors.New("auth: keyinfo not found"))
	}
	return clientRequest.SetHeaderParam("Authorization", fmt.Sprintf("Bearer %s", keyVal))
}
