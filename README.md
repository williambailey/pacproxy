pacproxy
========

[![Build Status](https://travis-ci.org/williambailey/pacproxy.svg)](https://travis-ci.org/williambailey/pacproxy)

A no-frills local HTTP proxy server powered by a [proxy auto-config (PAC) file](https://web.archive.org/web/20070602031929/http://wp.netscape.com/eng/mozilla/2.0/relnotes/demo/proxy-live.html). Especially handy when you are working in an environment with many different proxy servers and your applications don't support proxy auto-configuration.

```
$ ./pacproxy -h
Usage of ./bin/pacproxy:
  -c string
        PAC file name, url or javascript to use (required)
  -l string
        Interface and port to listen on (default "127.0.0.1:8080")
  -v    send verbose output to STDERR
```

```bash
# shell 1
pacproxy -c 'function FindProxyForURL(url, host){ console.log("hello pac world!"); return "PROXY random.example.com:8080"; }'
# shell 2
export http_proxy="127.0.0.1:8080"
export https_proxy="127.0.0.1:8080"
curl -I "http://www.example.com"
```

## License

> Copyright 2019 William Bailey
>
> Licensed under the Apache License, Version 2.0 (the "License");
> you may not use this file except in compliance with the License.
> You may obtain a copy of the License at
>
>     http://www.apache.org/licenses/LICENSE-2.0
>
> Unless required by applicable law or agreed to in writing, software
> distributed under the License is distributed on an "AS IS" BASIS,
> WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
> See the License for the specific language governing permissions and
> limitations under the License.
