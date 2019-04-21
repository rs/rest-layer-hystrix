# REST Layer Hystrix storage handler wrapper

[![godoc](http://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/rs/rest-layer-hystrix) [![license](http://img.shields.io/badge/license-MIT-red.svg?style=flat)](https://raw.githubusercontent.com/rs/rest-layer-hystrix/master/LICENSE) [![build](https://img.shields.io/travis/rs/rest-layer-hystrix.svg?style=flat)](https://travis-ci.org/rs/rest-layer-hystrix)

This [REST Layer](https://github.com/rs/rest-layer) resource storage wrapper uses [hystrix-go](https://github.com/afex/hystrix-go) to add circuit breaker support to any REST Layer resource storage handler.

## Usage

```go
import "github.com/rs/rest-layer-hystrix"
```

Wrap existing storage handler with a name that will be used to construct hystrix commands:

```go
s := restrix.Wrap("myResource", mem.NewHandler())
```

Use this handler with a resource:

```go
index.Bind("foo", foo, s, resource.DefaultConf)
```

Customize the hystrix commands:

```go
// Configure hystrix commands
hystrix.Configure(map[string]hystrix.CommandConfig{
    "posts.Find": {
        Timeout:               1000,
        MaxConcurrentRequests: 100,
        ErrorPercentThreshold: 25,
    },
    "posts.Insert": {
        Timeout:               1000,
        MaxConcurrentRequests: 50,
        ErrorPercentThreshold: 25,
    },
    ...
})
```

Start the metrics stream handler:

```go
hystrixStreamHandler := hystrix.NewStreamHandler()
hystrixStreamHandler.Start()
log.Print("Serving Hystrix metrics on http://localhost:8081")
go http.ListenAndServe(net.JoinHostPort("", "8081"), hystrixStreamHandler)
```
