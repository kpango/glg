package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/kpango/glg"
)

var dummyObject = struct {
	Age          int    `json:"age,omitempty"`
	Name         string `json:"name,omitempty"`
	IsProgrammer bool   `json:"is_programmer,omitempty"`
}{
	Age:          29,
	Name:         "Yusuke Kato",
	IsProgrammer: true,
}

// NetWorkLogger sample network logger
type NetWorkLogger struct{}

func (n NetWorkLogger) Write(b []byte) (int, error) {
	// http.Post("localhost:8080/log", "", bytes.NewReader(b))
	http.Get("http://127.0.0.1:8080/log")
	glg.Success("Requested")
	glg.Infof("RawString is %s", glg.RawString(b))
	return 1, nil
}

type RotateWriter struct {
	writer io.Writer
	dur    time.Duration
	once   sync.Once
	cancel context.CancelFunc
	mu     sync.Mutex
	buf    *bytes.Buffer
}

func NewRotateWriter(w io.Writer, dur time.Duration, buf *bytes.Buffer) io.WriteCloser {
	return &RotateWriter{
		writer: w,
		dur:    dur,
		buf:    buf,
	}
}

func (r *RotateWriter) Write(b []byte) (int, error) {
	if r.buf == nil || r.writer == nil {
		return 0, errors.New("error invalid rotate config")
	}
	r.once.Do(func() {
		var ctx context.Context
		ctx, r.cancel = context.WithCancel(context.Background())
		go func() {
			tick := time.NewTicker(r.dur)
			for {
				select {
				case <-ctx.Done():
					tick.Stop()
					return
				case <-tick.C:
					r.mu.Lock()
					r.writer.Write(r.buf.Bytes())
					r.buf.Reset()
					r.mu.Unlock()
				}
			}
		}()
	})
	r.mu.Lock()
	r.buf.Write(b)
	r.mu.Unlock()
	return len(b), nil
}

func (r *RotateWriter) Close() error {
	if r.cancel != nil {
		r.cancel()
	}
	return nil
}

func main() {
	// var errWriter io.Writer
	// var customWriter io.Writer
	infolog := glg.FileWriter("/tmp/info.log", 0o666)

	customTag := "FINE"
	customErrTag := "CRIT"

	errlog := glg.FileWriter("/tmp/error.log", 0o666)
	rotate := NewRotateWriter(os.Stdout, time.Second*10, bytes.NewBuffer(make([]byte, 0, 4096)))

	defer infolog.Close()
	defer errlog.Close()
	defer rotate.Close()

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
		// EnableJSON().
		AddLevelWriter(glg.INFO, infolog). // add info log file destination
		AddLevelWriter(glg.ERR, errlog).   // add error log file destination
		AddLevelWriter(glg.WARN, rotate)   // add error log file destination

	glg.Info("info")
	glg.Infof("%s : %s", "info", "formatted")
	glg.Log("log")
	glg.Logf("%s : %s", "info", "formatted")
	glg.Debug("debug")
	glg.Debugf("%s : %s", "info", "formatted")
	glg.Trace("Trace")
	glg.Tracef("%s : %s", "tracef", "formatted")
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

	// set global log level to ERR level
	glg.Info("before setting level to ERR this message will show")
	glg.Get().SetLevel(glg.ERR)
	glg.Info("after setting level to ERR this message will not show")
	glg.Error("this log is ERR level this will show")
	glg.Get().SetLevel(glg.DEBG)
	glg.Info("log level is now DEBG, this INFO level log will show")

	glg.Get().
		AddStdLevel(customTag, glg.STD, false).                    // user custom log level
		AddErrLevel(customErrTag, glg.STD, true).                  // user custom error log level
		SetLevelColor(glg.TagStringToLevel(customTag), glg.Cyan).  // set color output to user custom level
		SetLevelColor(glg.TagStringToLevel(customErrTag), glg.Red) // set color output to user custom level
	glg.CustomLog(customTag, "custom logging")
	glg.CustomLog(customErrTag, "custom error logging")

	// glg.Info("kpango's glg supports disable timestamp for logging")
	glg.Get().DisableTimestamp()
	glg.Info("timestamp disabled")
	glg.Warn("timestamp disabled")
	glg.Log("timestamp disabled")
	glg.Get().EnableTimestamp()
	glg.Info("timestamp enabled")
	glg.Warn("timestamp enabled")
	glg.Log("timestamp enabled")

	glg.Info("kpango's glg support json logging")
	glg.Get().EnableJSON()
	err := glg.Warn("kpango's glg", "support", "json", "logging")
	if err != nil {
		glg.Get().DisableJSON()
		glg.Error(err)
		glg.Get().EnableJSON()
	}
	err = glg.Info("hello", struct {
		Name   string
		Age    int
		Gender string
	}{
		Name:   "kpango",
		Age:    28,
		Gender: "male",
	}, 2020)
	if err != nil {
		glg.Get().DisableJSON()
		glg.Error(err)
		glg.Get().EnableJSON()
	}

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
			glg.CustomLog(customTag, dummyObject)
		}
	}()

	glg.Get().AddLevelWriter(glg.DEBG, NetWorkLogger{}).EnableJSON() // add info log file destination

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
