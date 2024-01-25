ARG BASE_TAG=latest

FROM golang:1.21.6-alpine3.19 AS build

RUN apk add --no-cache make git coreutils

WORKDIR /src
COPY . .

ARG BUILD_TAG

# Use go mod vendor to download imported package before building Docker image so no need to download here
RUN go mod download

RUN BUILDDIR=/out TAG=${BUILD_TAG} make example

FROM alpine:3.19 AS bin
COPY --from=build /out /out

EXPOSE 50090
EXPOSE 50091
