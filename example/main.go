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

	go func() {
		time.Sleep(time.Second * 5)
		for i := 0; i < 100; i++ {
			glg.Info("info")
		}
	}()

	go func() {
		time.Sleep(time.Second * 5)
		for i := 0; i < 100; i++ {
			glg.Debug("debug")
			time.Sleep(time.Millisecond * 100)
		}
	}()

	go func() {
		time.Sleep(time.Second * 5)
		for i := 0; i < 100; i++ {
			glg.Warn("warn")
		}
	}()

	go func() {
		time.Sleep(time.Second * 5)
		for i := 0; i < 100; i++ {
			glg.Error("error")
			time.Sleep(time.Millisecond * 100)
			glg.CustomLog(customTag, "custom logging")
		}
	}()

	glg.Get().AddLevelWriter(glg.DEBG, NetWorkLogger{}) // add info log file destination

	http.Handle("/glg", glg.HTTPLoggerFunc("glg sample", func(w http.ResponseWriter, r *http.Request) {
		glg.Info("glg HTTP server logger sample")
	}))

	http.Handle("/log", glg.HTTPLoggerFunc("log", func(w http.ResponseWriter, r *http.Request) {
		glg.Info("received")
	}))

	http.ListenAndServe(":8080", nil)

	// fatal logging
	glg.Fatalln("fatal")
}
