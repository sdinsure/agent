package otel

import (
	"context"
	"errors"
)

type CanShutdown interface {
	Shutdown(ctx context.Context) error
}

func ShutdownAll(ctx context.Context, canShutdownServices ...CanShutdown) error {
	errSlice := make([]error, len(canShutdownServices), len(canShutdownServices))
	for index, canShutdownService := range canShutdownServices {
		errSlice[index] = canShutdownService.Shutdown(ctx)
	}
	return errors.Join(errSlice...)
}
