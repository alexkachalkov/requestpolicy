# RequestPolicy Middleware

[![Build Status](https://github.com/alexkachalkov/requestpolicy/workflows/Main/badge.svg?branch=master)](https://github.com/alexkachalkov/requestpolicy/actions)

RequestPolicy Middleware is a Traefik plugin that allows you to control access to HTTP requests based on path and query parameter configurations.

## Features

- Support for whitelist and blacklist for paths and query parameters.
- Ability to specify regular expressions for paths and query parameters.
- Validation of regular expressions when creating a new instance of the middleware.

## Configuration

### Static

```yaml
pilot:
    token: "xxx"

experimental:
    plugins:
        requestpolicy:
            modulename: "github.com/alexkachalkov/requestpolicy"
            version: "v0.0.1"
```

### Dynamic

To configure the `RequestPolicy` plugin you should create a middleware in your dynamic configuration. The following example creates and uses the `requestpolicy` middleware plugin to allow all HTTP requests with a path starting with `/api` and block all HTTP requests with a path starting with `/admin`.

```yaml
http:
    routers:
        my-router:
            rule: "Host(`localhost`)"
            middlewares: ["requestpolicy"]
            service: "my-service"

    middlewares:
        requestpolicy:
            plugin:
                requestpolicy:
                    whitelistPaths:
                        - PathRegex: "^/api/.*"
                          QueryParamRegex: "(page)=.*"
                    blacklistPaths:
                        - PathRegex: "^/admin/.*"
                          QueryParamRegex: "(token)=.*"

    services:
        my-service:
            loadBalancer:
                servers:
                    - url: "http://127.0.0.1"
```