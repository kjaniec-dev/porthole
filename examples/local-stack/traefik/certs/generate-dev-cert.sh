#!/bin/sh

set -eu

script_dir=$(CDPATH= cd -- "$(dirname "$0")" && pwd)

openssl req \
  -x509 \
  -nodes \
  -newkey rsa:2048 \
  -days 3650 \
  -keyout "$script_dir/localstack.key" \
  -out "$script_dir/localstack.crt" \
  -subj "/CN=frontend.localhost" \
  -addext "subjectAltName=DNS:frontend.localhost,DNS:api.localhost,DNS:localhost"
