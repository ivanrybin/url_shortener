#!/bin/bash
cd cmd/url_shortener_server || exit
go build -o urls_server
chmod +x urls_server
mv urls_server ../../urls_server
