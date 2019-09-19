#!/usr/bin/env bash

mkdir -p certs
openssl genrsa -out certs/server.key 2048
openssl req -new -sha256 -key certs/server.key -out certs/server.csr -subj /CN=httwaterius
openssl x509 -req -sha256 -in certs/server.csr -signkey certs/server.key -out certs/server.crt -days 3650
chmod o+r certs/server.key certs/server.crt