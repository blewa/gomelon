# gomelon [![Build Status](https://travis-ci.org/goburrow/gomelon.svg)](https://travis-ci.org/goburrow/gomelon) [![GoDoc](https://godoc.org/github.com/goburrow/gomelon?status.svg)](https://godoc.org/github.com/goburrow/gomelon)
Lightweight Go framework for building web services, inspired by Dropwizard.

## Overview
Gomelon includes a number of libraries to build a web application quickly:

* [goji](https://github.com/zenazn/goji): a robust web framework.
* [gol](https://github.com/goburrow/gol): a simple hierarchical logging API.
* [metrics](https://github.com/codahale/metrics): a minimalist instrumentation library.
* [validator](https://github.com/go-validator/validator): extensible value validations.

Features supported:

- Commands: for controlling your application from command line.
- Bundles: for modularizing your application.
- Managed Objects: for starting and stopping your components.
- HealthChecks: for checking health of your application in production.
- Metrics: for monitoring and statistics.
- Tasks: for administration.
- Resources: for RESTful endpoints.
- Logging: for understanding behaviors of your application.
- Configuration: for application parameters.
- Banner: for fun. :)
- (maybe more)

## Example
See [example](https://github.com/goburrow/gomelon/tree/master/example)

```
INFO  [2015-02-04T22:46:18.062+10:00] gomelon.server: starting MyApp
    ______
   /\   __\______
  /..\  \  \     \
 /....\_____\  \  \
 \..../ / / /\_____\
  \../ / / /./   __/
   \/_____/./__   /
          \/_____/

INFO  [2015-02-04T22:46:18.063+10:00] gomelon.assets: registering AssetsBundle for path /static/
DEBUG [2015-02-04T22:46:18.063+10:00] gomelon.server: resources = [*main.resource]
INFO  [2015-02-04T22:46:18.063+10:00] gomelon.server: endpoints =

    GET     /user/:name (*main.resource)
    POST    /user/:name (*main.resource)

INFO  [2015-02-04T22:46:18.063+10:00] gomelon.admin: tasks =

    POST    /tasks/gc (*core.gcTask)
    POST    /tasks/log (*core.logTask)
    POST    /tasks/task1 (*main.myTask)

DEBUG [2015-02-04T22:46:18.063+10:00] gomelon.admin: health checks = [MyHealthCheck]
INFO  [2015-02-04T22:46:18.063+10:00] example: started MyComponent
INFO  [2015-02-04T22:46:18.063+10:00] gomelon.server: listening :8080
INFO  [2015-02-04T22:46:18.063+10:00] gomelon.server: listening :8081
```

## Contribution
The project still lacks of coffee and swears. Comments, issues and pull requests are welcome.
