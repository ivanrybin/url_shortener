#!/bin/bash
cd pkg/grpc || exit

protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       url_shortener.proto
