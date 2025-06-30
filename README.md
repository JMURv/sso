[![Go Coverage](https://github.com/JMURv/sso/wiki/coverage.svg)](https://raw.githack.com/wiki/JMURv/sso/coverage.html)
## Configuration
### App
Configuration files placed in `/configs/envs`
- Create your own `.env`, `.env.dev` and `.env.prod` based on `.env.example`


## Build and run
### Docker
You can just start docker compose file that will build images and start containers automatically via:
```shell
task dc-dev
```

## Addons
Also there is ability to up svcs like: `prometheus`, `jaeger`, `node-exporter`, `grafana`:
```shell
docker compose --env-file path/to/env -f compose-base.yaml up
```
Services are available at:

| Сервис     | Адрес                  |
|------------|------------------------|
| App        | http://localhost:8080  |
| Prometheus | http://localhost:9090  |
| Jaeger     | http://localhost:16686 |
| Grafana    | http://localhost:3000  |

___

## Tests
### E2E
Spin up all containers for `E2E` tests:
```shell
task dc-test
```
Wait until all containers are ready and then run: `task t-integ`