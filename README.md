# glg [![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg?style=flat-square)](https://opensource.org/licenses/Apache-2.0) [![release](https://img.shields.io/github/release/kpango/glg.svg?style=flat-square)](https://github.com/kpango/glg/releases/latest) [![CircleCI](https://circleci.com/gh/kpango/glg.svg)](https://circleci.com/gh/kpango/glg) [![codecov](https://codecov.io/gh/kpango/glg/branch/master/graph/badge.svg?token=2CzooNJtUu&style=flat-square)](https://codecov.io/gh/kpango/glg) [![Codacy Badge](https://api.codacy.com/project/badge/Grade/a6e544eee7bc49e08a000bb10ba3deed)](https://www.codacy.com/app/i.can.feel.gravity/glg?utm_source=github.com&amp;utm_medium=referral&amp;utm_content=kpango/glg&amp;utm_campaign=Badge_Grade) [![Go Report Card](https://goreportcard.com/badge/github.com/kpango/glg)](https://goreportcard.com/report/github.com/kpango/glg) [![GolangCI](https://golangci.com/badges/github.com/kpango/glg.svg?style=flat-square)](https://golangci.com/r/github.com/kpango/glg) [![Go Walker](http://gowalker.org/api/v1/badge)](https://gowalker.org/github.com/kpango/glg) [![GoDoc](http://godoc.org/github.com/kpango/glg?status.svg)](http://godoc.org/github.com/kpango/glg) [![DepShield Badge](https://depshield.sonatype.org/badges/kpango/glg/depshield.svg)](https://depshield.github.io) [![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fkpango%2Fglg.svg?type=shield)](https://app.fossa.io/projects/git%2Bgithub.com%2Fkpango%2Fglg?ref=badge_shield)


glg is simple golang logging library

## Requirement
Go 1.11

## Installation
```shell
go get github.com/kpango/glg
```

## Example
```go
package main

import (
	"net/http"
	"time"

	"github.com/kpango/glg"
)

// NetWorkLogger sample network logger
type NetWorkLogger struct{}

func (n NetWorkLogger) Write(b []byte) (int, error) {
	// http.Post("localhost:8080/log", "", bytes.NewReader(b))
	http.Get("http://127.0.0.1:8080/log")
	glg.Success("Requested")
	glg.Infof("RawString is %s", glg.RawString(b))
	return 1, nil
}

func main() {

	// var errWriter io.Writer
	// var customWriter io.Writer
	infolog := glg.FileWriter("/tmp/info.log", 0666)

	customTag := "FINE"
	customErrTag := "CRIT"

	errlog := glg.FileWriter("/tmp/error.log", 0666)
	defer infolog.Close()
	defer errlog.Close()
	glg.Get().
		SetMode(glg.BOTH). // default is STD
		// DisableColor().
		// SetMode(glg.NONE).
		// SetMode(glg.WRITER).
		// SetMode(glg.BOTH).
		// InitWriter().
		// AddWriter(customWriter).
		// SetWriter(customWriter).
		// AddLevelWriter(glg.LOG, customWriter).
		// AddLevelWriter(glg.INFO, customWriter).
		// AddLevelWriter(glg.WARN, customWriter).
		// AddLevelWriter(glg.ERR, customWriter).
		// SetLevelWriter(glg.LOG, customWriter).
		// SetLevelWriter(glg.INFO, customWriter).
		// SetLevelWriter(glg.WARN, customWriter).
		// SetLevelWriter(glg.ERR, customWriter).
		AddLevelWriter(glg.INFO, infolog).                         // add info log file destination
		AddLevelWriter(glg.ERR, errlog).                           // add error log file destination
		AddStdLevel(customTag, glg.STD, false).                    //user custom log level
		AddErrLevel(customErrTag, glg.STD, true).                  // user custom error log level
		SetLevelColor(glg.TagStringToLevel(customTag), glg.Cyan).  // set color output to user custom level
		SetLevelColor(glg.TagStringToLevel(customErrTag), glg.Red) // set color output to user custom level

	glg.Info("info")
	glg.Infof("%s : %s", "info", "formatted")
	glg.Log("log")
	glg.Logf("%s : %s", "info", "formatted")
	glg.Debug("debug")
	glg.Debugf("%s : %s", "info", "formatted")
	glg.Warn("warn")
	glg.Warnf("%s : %s", "info", "formatted")
	glg.Error("error")
	glg.Errorf("%s : %s", "info", "formatted")
	glg.Success("ok")
	glg.Successf("%s : %s", "info", "formatted")
	glg.Fail("fail")
	glg.Failf("%s : %s", "info", "formatted")
	glg.Print("Print")
	glg.Println("Println")
	glg.Printf("%s : %s", "printf", "formatted")
	glg.CustomLog(customTag, "custom logging")
	glg.CustomLog(customErrTag, "custom error logging")

	glg.Get().AddLevelWriter(glg.DEBG, NetWorkLogger{}) // add info log file destination

	http.Handle("/glg", glg.HTTPLoggerFunc("glg sample", func(w http.ResponseWriter, r *http.Request) {
		glg.New().
		AddLevelWriter(glg.Info, NetWorkLogger{}).
		AddLevelWriter(glg.Info, w).
		Info("glg HTTP server logger sample")
	}))

	http.ListenAndServe("port", nil)

	// fatal logging
	glg.Fatalln("fatal")
}
```

![Sample Logs](https://github.com/kpango/glg/raw/master/images/sample.png)

## Benchmarks

![Bench](https://github.com/kpango/glg/raw/master/images/bench.png)

## Contribution
1. Fork it ( https://github.com/kpango/glg/fork )
2. Create your feature branch (git checkout -b my-new-feature)
3. Commit your changes (git commit -am 'Add some feature')
4. Push to the branch (git push origin my-new-feature)
5. Create new Pull Request

## Author
[kpango](https://github.com/kpango)

## LICENSE
glg released under MIT license, refer [LICENSE](https://github.com/kpango/glg/blob/master/LICENSE) file.  
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fkpango%2Fglg.svg?type=large)](https://app.fossa.io/projects/git%2Bgithub.com%2Fkpango%2Fglg?ref=badge_large)
