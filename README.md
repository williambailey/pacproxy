pacproxy
========

A no-frills local HTTP proxy server powered by a proxy auto-config (PAC) file. Especially handy when you are working in an environment with many different proxy servers and your applications don't support proxy auto-configuration.

```bash
pacproxy &
export http_proxy="127.0.0.1:12345"
export https_proxy="127.0.0.1:12345"
curl -v -X HEAD "http://www.example.com"
```

TODO
----

- [x] HTTP - DIRECT
- [x] HTTP - PROXY
- [x] HTTPS - DIRECT
- [x] HTTPS - PROXY
- [x] Via HTTP response header (HTTP only)
- [ ] CLI flag - PAC file location
- [ ] CLI flag - listen spec
- [ ] CLI flag - verbose
- [ ] CLI flag - help
- [ ] PAC - Multi-value FindProxyForURL() support
- [ ] SIGHUP to reload PAC file
- [ ] Pac lookup via cli
- [ ] HTTP stats/lookup/admin UI
