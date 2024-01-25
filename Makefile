BUILDDIR ?= "./out"
BUILDTIME=$(shell date --rfc-3339=seconds)
TAG ?= ""

example: example-server example-client

example-server:
	cd example/server/main && \
	env CGO_ENABLED=0 \
	go build -ldflags '-X "github.com/sdinsure/pkg/version.BuildTime='"${BUILDTIME}"'" -extldflags "-static"' -tags="${TAG}" -o ${BUILDDIR}/example-server

example-client:
	cd example/client \
	env CGO_ENABLED=0 && \
	go build -ldflags '-X "github.com/sdinsure/pkg/version.BuildTime='"${BUILDTIME}"'" -extldflags "-static"' -tags="${TAG}" -o ${BUILDDIR}/example-client

version:
	@echo buildtime: $(BUILDTIME)
