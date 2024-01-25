# gommip

HTTP service to report ip information based on MaxMind's GeoIP database with automatic database update.

# build & run

## direct build

Simply `go build` then run `gommip --help` to see help

```shell
# replace ip.foobar.one with your own deployment
# request to root path '/' will report information for your public ip
# the value of result field depends on database  
> curl https://ip.foobar.one/1.1.1.1
HTTP/1.1 200 OK
Content-Length: 375
Content-Type: application/json; charset=utf-8
Date: Thu, 25 Jan 2024 16:43:15 GMT

{
    "msg": "OK",
    "result": {
        "asn": {
            "autonomous_system_number": 13335,
            "autonomous_system_organization": "CLOUDFLARENET"
        },
        "city": {
            "registered_country": {
                "geoname_id": 2077456,
                "iso_code": "AU",
                "names": {
                    "de": "Australien",
                    "en": "Australia",
                    "es": "Australia",
                    "fr": "Australie",
                    "ja": "オーストラリア",
                    "pt-BR": "Austrália",
                    "ru": "Австралия",
                    "zh-CN": "澳大利亚"
                }
            }
        },
        "ip": "1.1.1.1"
    }
}
```

## container

Run 

```shell
docker run --rm -v "$PWD":/usr/src/app -w /usr/src/app docker.io/library/golang:alpine go build -v
```

To build inside the golang:alpine container which will produce a binrary
suitable for using with `docker build` to create a container image.

## tips

To reuse reuse the go cache in your host machine:

```shell
docker run --rm -v "$HOME/go:/go" -v "$PWD":/usr/src/app -w /usr/src/app docker.io/library/golang:alpine go build -v
```


# license

This project is released under the MIT License.

Thans for https://github.com/P3TERX/GeoLite.mmdb for dabebase used in sample config file.