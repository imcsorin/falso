# Falso

It is a CLI that allows you to mock requests/responses between you and any server without any configuration or previous
knowledge about how it works. It works at the TCP level, in theory you can use it to mock any TCP connection.

### Why Falso exists

I needed a tool to mock requests in my E2E tests, but I couldn't find anything easy to use/configure or
that doesn't require a lot of dependencies. I'm mainly using this with Detox for React Native and
Cypress for web.

### How to download

Download the latest release and make it executable:

#### Linux
```
curl -sL https://github.com/sorinsi/falso/releases/latest/download/falso.linux --output falso && chmod +x falso
```

#### MacOS
```
curl -sL https://github.com/sorinsi/falso/releases/latest/download/falso.darwin --output falso && chmod +x falso
```

#### Windows
Binary release available at:
`https://github.com/sorinsi/falso/releases/latest/download/falso.windows`

### Available args

- `--address` CLI will listen at this address, default is `localhost:8080`
- `--remoteAddress` remote server address to mock.
- `--mode` there are 2 modes, `proxy` (records) and `mock` (serves saved data).
- `--dataPath` custom path to save data in proxy mode, default is `./falsoData`.
- `--bufferSize` in case you need larger responses/requests, default is 65535 (64K).
- `--overwrite` will overwrite existing files in proxy mode, default is false.

### How to proxy requests.

`falso --address localhost:8080 --remoteAddress localhost:8000 --mode proxy`

Now you can run your tests once, and it will automatically save everything.

### How to mock requests

`falso --address localhost:8080 --mode mock`

You will be able to run your tests without connecting to the remote server.

### Author

Sorin C.
