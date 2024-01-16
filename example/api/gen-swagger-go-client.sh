#!/usr/bin/env bash

# to install swagger: 
# wget https://github.com/go-swagger/go-swagger/releases/download/v0.30.5/swagger_darwin_amd64 && \
# mv swagger_darwin_amd64 swagger
# chmod +x swagger

mkdir go-openapiv2
swagger generate client -f openapiv2/*.swagger.json --target go-openapiv2
