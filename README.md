# log-zap

[![Go Reference](https://pkg.go.dev/badge/github.com/go-haru/log-zap.svg)](https://pkg.go.dev/github.com/go-haru/log-zap)
[![License](https://img.shields.io/github/license/go-haru/log-zap)](./LICENSE)
[![Release](https://img.shields.io/github/v/release/go-haru/log-zap.svg?style=flat-square)](https://github.com/go-haru/log-zap/releases)
[![Go Test](https://github.com/go-haru/log-zap/actions/workflows/go.yml/badge.svg)](https://github.com/go-haru/log-zap/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/go-haru/log-zap)](https://goreportcard.com/report/github.com/go-haru/log-zap)

log-zap is wrapper of [go.uber.org/zap](https://github.com/uber-go/zap) for [log](https://github.com/go-haru/log)

## Usage

```go
package main

import (
    "log"

    hlog "github.com/go-haru/log"

    logger "github.com/go-haru/log-zap"
)

func main() {
    var config = logger.Options{
        Level:     hlog.InfoLevel.String(),
        Format:    logger.FormatText,
        LongTime:  true,
        WithColor: true,
    }

    // build wrapped zap logger
    var zapLogger, err = logger.New(config)
    if err != nil {
        panic(err)
    }

    // register logger
    hlog.Use(zapLogger)

    // takeover sdk logger's output 
    log.SetFlags(0)
    log.SetOutput(zapLogger.
        WithLevel(hlog.InfoLevel).
        AddDepth(-1).
        Standard().Writer(),
    )
}
```

## Contributing

For convenience of PM, please commit all issue to [Document Repo](https://github.com/go-haru/go-haru/issues).

## License

This project is licensed under the `Apache License Version 2.0`.

Use and contributions signify your agreement to honor the terms of this [LICENSE](./LICENSE).

Commercial support or licensing is conditionally available through organization email.
