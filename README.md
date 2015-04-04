pacproxy
========

A no-frills local HTTP proxy server powered by a [proxy auto-config (PAC) file](https://web.archive.org/web/20070602031929/http://wp.netscape.com/eng/mozilla/2.0/relnotes/demo/proxy-live.html). Especially handy when you are working in an environment with many different proxy servers and your applications don't support proxy auto-configuration.

```bash
pacproxy &
export http_proxy="127.0.0.1:12345"
export https_proxy="127.0.0.1:12345"
curl -I "http://www.example.com"
```

TODO
----

- [ ] PAC - Multi-value FindProxyForURL() support
- [ ] PAC - SOCKS support
- [ ] SIGHUP to reload PAC file
- [ ] Pac lookup via cli
- [ ] HTTP stats/lookup/admin UI
