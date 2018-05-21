# beacon.api

| | | | | | |
| :--- | :---- | :--- | :--- | :--- | :--- |
| [![travis-img]][2] | [![codecov-img]][4] | [![version-img]][8] | [![docker-img]][6] | [![docs-img]][10] | [![reportcard]][11] |

The [golang] api server for the beacon [blink(1)] platform.


## Setup

To get started locally, you will need to have a running [redis] server. The connection used by the application at 
runtime can be configured using the `REDIS_URI` environment variable or the `-redisuri` command line argument (env var
will take precedence).


#### Server &amp; Device Keys

Communication between the server and prospective devices leverage rsa public &amp; private keys to authenticate with 
one another. The server loads this key at runtime and will provide the devices with a public version of it on "welcome".

To generate the server key:

```
mkdir .keys
openssl genrsa -des3 -out .keys/private.pem 2048
openssl rsa -in .keys/private.pem -outform PEM -pubout -out .keys/public.pem
```

## Contributing

All contributions welcome.

[travis-img]: https://img.shields.io/travis/dadleyy/beacon.api.svg?style=flat-square
[2]: https://travis-ci.org/dadleyy/beacon.api
[codecov-img]: https://img.shields.io/codecov/c/github/dadleyy/beacon.api.svg?style=flat-square
[4]: https://codecov.io/gh/dadleyy/beacon.api
[docker-img]: https://img.shields.io/docker/pulls/dadleyy/beacon-api.svg?style=flat-square
[6]: https://hub.docker.com/r/dadleyy/beacon-api
[version-img]: https://img.shields.io/github/release/dadleyy/beacon.api.svg?style=flat-square
[8]: https://github.com/dadleyy/beacon.api/releases
[docs-img]: http://img.shields.io/badge/godoc-reference-5272B4.svg?style=flat-square
[10]: https://godoc.org/github.com/dadleyy/beacon.api
[11]: https://goreportcard.com/report/github.com/dadleyy/beacon.api
[reportcard]: https://goreportcard.com/badge/github.com/dadleyy/beacon.api?style=flat-square
[blink(1)]: https://blink1.thingm.com/
[golang]: https://golang.org
[redis]: https://redis.io/
