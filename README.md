# glg

[![Join the chat at https://gitter.im/kpango/glg](https://badges.gitter.im/kpango/glg.svg)](https://gitter.im/kpango/glg?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)
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
		AddStdLevel(customLevel).              //user custom log level
		AddErrLevel(customErrLevel).           // user custom error log level
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

	glg.CustomLog(customLevel, "custom logging")
	glg.CustomLog(customErrLevel, "custom error logging")

	// HTTP Handler Logger
	http.Handle("/glg", glg.HTTPLoggerFunc("glg sample", func(w http.ResponseWriter, r *http.Request) {

		glg.Info("glg HTTP server logger sample")
		fmt.Fprint(w, "glg HTTP server logger sample")

	}))

	http.ListenAndServe(":8080", nil)

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
