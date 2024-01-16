#!/usr/bin/env bash
#
# ====buf installation====
#
# BIN="/usr/local/bin" && \
# VERSION="1.0.0-rc9" && \
# BINARY_NAME="buf" && \
#   curl -sSL \
#     "https://github.com/bufbuild/buf/releases/download/v${VERSION}/${BINARY_NAME}-$(uname -s)-$(uname -m)" \
#     -o "${BIN}/${BINARY_NAME}" && \
#   chmod +x "${BIN}/${BINARY_NAME}"
#
# ====Init====
# run `buf config init` to setup buf.yaml

buf mod update
buf generate
