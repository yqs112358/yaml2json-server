# yaml2json-server
A simple server, support converting yaml to json through http requests.

## Deployment

### Binary

##### linux

```shell
./yaml2json-server --listen 0.0.0.0 --port 8080 --key 3kf7^21%P9d --sub-path "/convert"
```

##### windows

```shell
yaml2json-server.exe --listen 0.0.0.0 --port 8080 --key 3kf7^21%P9d --sub-path "/convert"
```

### In Docker

docker-compose.yml

```yaml
services:
  yaml2json-server:
    container_name: "yaml2json-server"
    image: "yqs112358/yaml2json-server"
    restart: unless-stopped
    environment:
      Y2JS_LISTEN_ADDR: "0.0.0.0"
      Y2JS_LISTEN_PORT: 8080
      Y2JS_AUTH_KEY: "3kf7^21%P9d"
      Y2JS_URL_SUB_PATH: "/convert"
    ports:
      - 8080:8080
```

## Usage

### YAML in POST Request

Convert the YAML plain text in POST request body to JSON.

##### Example

```shell
curl -X POST https://<YOUR-SERVER-ADDRESS>/convert -d $'key1: 1234\nkey2:\n  nestedKey: "data"'
```

##### Response:

```json
{
    "key1": 1234,
    "key2": {
        "nestedKey": "data"
    }
}
```

### YAML from given URL

Read YAML text from the given encoded URL, then convert to JSON.

> For example, convert from https://another-domain/dir/data.yaml

##### Example:

```shell
curl https://<YOUR-SERVER-ADDRESS>/convert?url=https%3A%2F%2Fanother-domain%2Fdir%2Fdata.yaml
```

##### Response:

```json
{
    "key1": 1234,
    "key2": {
        "nestedKey": "data"
    }
}
```

### Authentication

Pre-shared auth key set in the configuration to avoid http-api abuse. Use one of the following methods to carry the key with request:

##### URL paramater

```shell
curl https://<YOUR-SERVER-ADDRESS>/convert?key=<YOUR-AUTH-KEY>& ......
```

##### HTTP Basic Auth

```shell
AUTH_DATA=$(echo -n ":<YOUR-AUTH-KEY>" | base64)
curl -H "Authorization: Basic $AUTH_DATA" https://<YOUR-SERVER-ADDRESS>/convert ...
```

## Configurations

| Command-line Arg | Environment Var   | Default Value | Description |
| ---------------- | ----------------- | ------------- | ----------- |
| `--listen`       | Y2JS_LISTEN_ADDR  | "0.0.0.0"     |             |
| `--port`         | Y2JS_LISTEN_PORT  | 8080          |             |
| `--key`          | Y2JS_AUTH_KEY     | ""            |             |
| `--sub-path`     | Y2JS_URL_SUB_PATH | "/"           |             |

## Build

```shell
go build -o ./dist -ldflags="-s -w" -trimpath yaml2json-server
```

## Reference

y2jLib in [bronze1man/yaml2json](https://github.com/bronze1man/yaml2json)
