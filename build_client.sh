#!/bin/bash
cd cmd/url_shortener_client || exit
go build -o urls_client
chmod +x urls_client
mv urls_client ../../urls_client
