# Local Stack Example

This example gives `porthole` something real to inspect:

- Traefik with live routers and services
- a React frontend service
- a Python backend service
- a self-signed local TLS certificate loaded by Traefik

## What It Starts

- `traefik` on `http://127.0.0.1:8080` for the dashboard and API
- `frontend` on `https://frontend.localhost`
- `backend` on `https://api.localhost`

The `*.localhost` names should resolve locally on modern systems without editing `/etc/hosts`.

## Run It

From the repository root:

```sh
docker compose -f examples/local-stack/compose.yaml up --build
```

Then, in another terminal:

```sh
cp examples/local-stack/.porthole.yaml.example .porthole.yaml
./porthole
```

## Quick Checks

Traefik API:

```sh
curl http://127.0.0.1:8080/api/http/routers
curl http://127.0.0.1:8080/api/http/services
openssl x509 -in examples/local-stack/traefik/certs/localstack.crt -noout -subject -issuer -dates
```

Frontend and backend:

```sh
curl -sk --noproxy '*' --resolve frontend.localhost:443:127.0.0.1 https://frontend.localhost
curl -sk --noproxy '*' --resolve frontend.localhost:443:127.0.0.1 https://frontend.localhost/api/hello
curl -sk --noproxy '*' --resolve api.localhost:443:127.0.0.1 https://api.localhost/hello
```

## Notes

- The certificate is self-signed for local development only.
- The example pins Traefik `v2.10` for a stable local stack, but the Traefik API still does not expose certificate metadata through the documented endpoints.
- The example config sets `traefik.certificate_files` so `porthole` can read the local PEM file and populate the Certs tab for file-provider setups.
- The frontend calls the backend through `https://frontend.localhost/api/...` so the browser only needs one trusted origin for the interactive demo.
- The frontend loads React directly in the browser from `esm.sh` to avoid a build tool in this example.
- `porthole` talks to the Traefik API over HTTP on `127.0.0.1:8080`, so the self-signed cert does not affect the dashboard connection.
