#!/usr/bin/env bash
#
docker run --rm --name postgres \
    -e TZ=gmt+8 \
    -e POSTGRES_USER=postgres \
    -e POSTGRES_PASSWORD=password \
    -e POSTGRES_DB=unittest \
    -p 5432:5432 -d library/postgres:14.1

