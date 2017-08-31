# glg [![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT) [![release](https://img.shields.io/github/release/kpango/glg.svg)](https://github.com/kpango/glg/releases/latest) [![CircleCI](https://circleci.com/gh/kpango/glg.svg?style=shield)](https://circleci.com/gh/kpango/glg) [![codecov](https://codecov.io/gh/kpango/glg/branch/master/graph/badge.svg)](https://codecov.io/gh/kpango/glg) [![Codacy Badge](https://api.codacy.com/project/badge/Grade/a6e544eee7bc49e08a000bb10ba3deed)](https://www.codacy.com/app/i.can.feel.gravity/glg?utm_source=github.com&amp;utm_medium=referral&amp;utm_content=kpango/glg&amp;utm_campaign=Badge_Grade) [![Go Report Card](https://goreportcard.com/badge/github.com/kpango/glg)](https://goreportcard.com/report/github.com/kpango/glg) [![GoDoc](http://godoc.org/github.com/kpango/glg?status.svg)](http://godoc.org/github.com/kpango/glg) [![Join the chat at https://gitter.im/kpango/glg](https://badges.gitter.im/kpango/glg.svg)](https://gitter.im/kpango/glg?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)

glg is simple golang logging library

## Requirement
Go 1.8

## Installation
```shell
go get github.com/kpango/glg
```

## Example
```go
	infolog := glg.FileWriter("/tmp/info.log", 0666)
	defer infolog.Close()

	customLevel := "FINE"
	customErrLevel := "CRIT"

	glg.Get().
		SetMode(glg.BOTH). // default is STD
		// SetMode(glg.NONE).  //nothing
		// SetMode(glg.WRITER). // io.Writer logging
		// SetMode(glg.BOTH). // stdout and file logging
		// InitWriter(). // initialize glg logger writer
		// AddWriter(customWriter). // add stdlog output destination
		// SetWriter(customWriter). // overwrite stdlog output destination
		// AddLevelWriter(glg.LOG, customWriter). // add LOG output destination
		// AddLevelWriter(glg.INFO, customWriter). // add INFO log output destination
		// AddLevelWriter(glg.WARN, customWriter). // add WARN log output destination
		// AddLevelWriter(glg.ERR, customWriter). // add ERR log output destination
		// SetLevelWriter(glg.LOG, customWriter). // overwrite LOG output destination
		// SetLevelWriter(glg.INFO, customWriter). // overwrite INFO log output destination
		// SetLevelWriter(glg.WARN, customWriter). // overwrite WARN log output destination
		// SetLevelWriter(glg.ERR, customWriter). // overwrite ERR log output destination
		AddLevelWriter(glg.INFO, infolog). // add info log file destination
		// AddLevelWriter(glg.INFO, glg.FileWriter("/tmp/info.log", 0666)). // add info log file destination
		// AddLevelWriter(glg.ERR, glg.FileWriter("/tmp/errors.log", 0666)). // add error log file destination
		AddStdLevel(customLevel, glg.STD, false).   //user custom log level
		AddErrLevel(customErrLevel, glg.STD, true). // user custom error log level
		SetLevelColor(customErrLevel, glg.Red) // set color output to user custom level

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

	glg.Get().DisableColor()
	glg.CustomLog(customLevel, "custom logging")
	glg.Get().EnableColor()
	glg.CustomLog(customErrLevel, "custom error logging")

	// HTTP Handler Logger
	http.Handle("/glg", glg.HTTPLoggerFunc("glg sample", func(w http.ResponseWriter, r *http.Request) {

		glg.Info("glg HTTP server logger sample")
		fmt.Fprint(w, "glg HTTP server logger sample")

	}))

	http.ListenAndServe("port", nil)

	// fatal logging
	glg.Fatalln("fatal")
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
