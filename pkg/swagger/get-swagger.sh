#!/usr/bin/env bash

rm -rf swaggerui

wget -O swaggerui.tar.gz https://github.com/swagger-api/swagger-ui/archive/refs/tags/v5.11.0.tar.gz
tar xzf swaggerui.tar.gz
mv swagger-ui-5.11.0/dist swaggerui
rm -rf swagger-ui-5.11.0 swaggerui.tar.gz
