// MIT License
//
// Copyright (c) 2019 kpango (Yusuke Kato)
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

// Package glg can quickly output that are colored and leveled logs with simple syntax
package glg

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	json "github.com/goccy/go-json"
	"github.com/kpango/fastime"
)

// Glg is glg base struct
type Glg struct {
	bs           *uint64
	logger       loggers
	levelCounter *uint32
	levelMap     levelMap
	buffer       sync.Pool
	callerDepth  int
	enableJSON   bool
}

// JSONFormat is json object structure for logging
type JSONFormat struct {
	Date   string      `json:"date,omitempty"`
	Level  string      `json:"level,omitempty"`
	File   string      `json:"file,omitempty"`
	Detail interface{} `json:"detail,omitempty"`
}

// MODE is logging mode (std only, writer only, std & writer)
type MODE uint8

// LEVEL is log level
type LEVEL uint8

type wMode uint8

type traceMode int64

type logger struct {
	tag              string
	rawtag           []byte
	writer           io.Writer
	std              io.Writer
	color            func(string) string
	isColor          bool
	traceMode        traceMode
	mode             MODE
	prevMode         MODE
	writeMode        wMode
	disableTimestamp bool
}

const (
	// DEBG is debug log level
	DEBG LEVEL = iota + 1
	// TRACE is trace log level
	TRACE
	// PRINT is print log level
	PRINT
	// LOG is log level
	LOG
	// INFO is info log level
	INFO
	// OK is success notify log level
	OK
	// WARN is warning log level
	WARN
	// ERR is error log level
	ERR
	// FAIL is failed log level
	FAIL
	// FATAL is fatal log level
	FATAL

	// UNKNOWN is unknown log level
	UNKNOWN LEVEL = LEVEL(math.MaxUint8)

	// NONE is disable Logging
	NONE MODE = iota + 1
	// STD is std log mode
	STD
	// BOTH is both log mode
	BOTH
	// WRITER is io.Writer log mode
	WRITER

	// Internal writeMode
	writeColorStd wMode = iota + 1
	writeStd
	writeWriter
	writeColorBoth
	writeBoth
	none

	// Default Format
	df = "%v %v %v %v %v %v %v %v %v %v %v %v %v %v %v %v %v %v %v %v " +
		"%v %v %v %v %v %v %v %v %v %v %v %v %v %v %v %v %v %v %v %v " +
		"%v %v %v %v %v %v %v %v %v %v %v %v %v %v %v %v %v %v %v %v " +
		"%v %v %v %v %v %v %v %v %v %v %v %v %v %v %v %v %v %v %v %v "

	dfl = len(df) / 3

	timeFormat = "2006-01-02 15:04:05"

	// return code
	rc  = "\n"
	rcl = len(rc)

	tab   = "\t"
	lsep  = tab + "["
	lsepl = len(lsep)
	sep   = "]:" + tab
	sepl  = len(sep)

	TraceLineNone traceMode = 1 << iota
	TraceLineShort
	TraceLineLong

	DefaultCallerDepth = 2
)

var (
	glg  *Glg
	once sync.Once

	// exit for Faltal error
	exit = os.Exit
)

func init() {
	Get()
}

func (l LEVEL) String() string {
	switch l {
	case DEBG:
		return "DEBG"
	case TRACE:
		return "TRACE"
	case PRINT:
		return "PRINT"
	case LOG:
		return "LOG"
	case INFO:
		return "INFO"
	case OK:
		return "OK"
	case WARN:
		return "WARN"
	case ERR:
		return "ERR"
	case FAIL:
		return "FAIL"
	case FATAL:
		return "FATAL"
	}
	return ""
}

func (l *logger) updateMode() *logger {
	switch {
	case l.mode == WRITER && l.writer != nil:
		l.writeMode = writeWriter
	case l.mode == BOTH && l.isColor && l.writer != nil:
		l.writeMode = writeColorBoth
	case l.mode == BOTH && !l.isColor && l.writer != nil:
		l.writeMode = writeBoth
	case l.isColor && ((l.mode == BOTH && l.writer == nil) || l.mode == STD):
		l.writeMode = writeColorStd
	case !l.isColor && ((l.mode == BOTH && l.writer == nil) || l.mode == STD):
		l.writeMode = writeStd
	default:
		l.writeMode = none
	}
	return l
}

// New returns plain glg instance
func New() *Glg {
	g := &Glg{
		levelCounter: new(uint32),
		callerDepth:  DefaultCallerDepth,
	}
	g.bs = new(uint64)

	atomic.StoreUint64(g.bs, uint64(len(timeFormat)+lsepl+sepl))

	g.buffer = sync.Pool{
		New: func() interface{} {
			return bytes.NewBuffer(make([]byte, 0, int(atomic.LoadUint64(g.bs))))
		},
	}

	atomic.StoreUint32(g.levelCounter, uint32(FATAL))

	for lev, log := range map[LEVEL]*logger{
		// standard out
		DEBG: {
			std:       os.Stdout,
			color:     Purple,
			isColor:   true,
			mode:      STD,
			traceMode: TraceLineNone,
		},
		TRACE: {
			std:       os.Stdout,
			color:     Yellow,
			isColor:   true,
			mode:      STD,
			traceMode: TraceLineNone,
		},
		PRINT: {
			std:       os.Stdout,
			color:     Colorless,
			isColor:   true,
			mode:      STD,
			traceMode: TraceLineNone,
		},
		LOG: {
			std:       os.Stdout,
			color:     Colorless,
			isColor:   true,
			mode:      STD,
			traceMode: TraceLineNone,
		},
		INFO: {
			std:       os.Stdout,
			color:     Green,
			isColor:   true,
			mode:      STD,
			traceMode: TraceLineNone,
		},
		OK: {
			std:       os.Stdout,
			color:     Cyan,
			isColor:   true,
			mode:      STD,
			traceMode: TraceLineNone,
		},
		WARN: {
			std:       os.Stdout,
			color:     Orange,
			isColor:   true,
			mode:      STD,
			traceMode: TraceLineNone,
		},
		// error out
		ERR: {
			std:       os.Stderr,
			color:     Red,
			isColor:   true,
			mode:      STD,
			traceMode: TraceLineShort,
		},
		FAIL: {
			std:       os.Stderr,
			color:     Red,
			isColor:   true,
			mode:      STD,
			traceMode: TraceLineShort,
		},
		FATAL: {
			std:       os.Stderr,
			color:     Red,
			isColor:   true,
			mode:      STD,
			traceMode: TraceLineLong,
		},
	} {
		log.tag = lev.String()
		log.rawtag = []byte(lsep + log.tag + sep)
		log.prevMode = log.mode
		log.updateMode()
		g.logger.Store(lev, log)
	}

	return g
}

// Get returns singleton glg instance
func Get() *Glg {
	once.Do(func() {
		fastime.SetFormat(timeFormat)
		glg = New()
	})
	return glg
}

func (g *Glg) EnableJSON() *Glg {
	g.enableJSON = true
	return g
}

func (g *Glg) DisableJSON() *Glg {
	g.enableJSON = false
	return g
}

func (g *Glg) EnablePoolBuffer(size int) *Glg {
	for range make([]struct{}, size) {
		g.buffer.Put(g.buffer.Get().(*bytes.Buffer))
	}
	return g
}

// SetLevel sets glg global log level
func (g *Glg) SetLevel(lv LEVEL) *Glg {
	g.logger.Range(func(lev LEVEL, l *logger) bool {
		if lev < lv {
			l.prevMode = l.mode
			l.mode = NONE
		} else {
			l.mode = l.prevMode
		}
		l.updateMode()
		g.logger.Store(lev, l)
		return true
	})
	return g
}

// SetMode sets glg logging mode
func (g *Glg) SetMode(mode MODE) *Glg {
	g.logger.Range(func(lev LEVEL, l *logger) bool {
		l.mode = mode
		l.prevMode = mode
		l.updateMode()
		g.logger.Store(lev, l)
		return true
	})

	return g
}

// SetLevelMode sets glg logging mode* per level
func (g *Glg) SetLevelMode(level LEVEL, mode MODE) *Glg {
	l, ok := g.logger.Load(level)
	if ok {
		l.mode = mode
		l.prevMode = mode
		l.updateMode()
		g.logger.Store(level, l)
	}
	return g
}

// SetPrefix sets Print logger prefix
func SetPrefix(lev LEVEL, pref string) *Glg {
	return glg.SetPrefix(lev, pref)
}

// SetPrefix sets Print logger prefix
func (g *Glg) SetPrefix(lev LEVEL, pref string) *Glg {
	l, ok := g.logger.Load(lev)
	if ok {
		l.tag = pref
		l.rawtag = []byte(lsep + l.tag + sep)
		g.logger.Store(lev, l)
	}
	return g
}

// GetCurrentMode returns current logging mode
func (g *Glg) GetCurrentMode(level LEVEL) MODE {
	l, ok := g.logger.Load(level)
	if ok {
		return l.mode
	}
	return NONE
}

// InitWriter is initialize glg writer
func (g *Glg) InitWriter() *Glg {
	g.logger.Range(func(lev LEVEL, l *logger) bool {
		l.writer = nil
		l.updateMode()
		g.logger.Store(lev, l)
		return true
	})
	return g
}

// SetWriter sets writer to glg std writers
func (g *Glg) SetWriter(writer io.Writer) *Glg {
	if writer == nil {
		return g
	}

	g.logger.Range(func(lev LEVEL, l *logger) bool {
		l.writer = writer
		l.updateMode()
		g.logger.Store(lev, l)
		return true
	})

	return g
}

// AddWriter adds writer to glg std writers
func (g *Glg) AddWriter(writer io.Writer) *Glg {
	if writer == nil {
		return g
	}

	g.logger.Range(func(lev LEVEL, l *logger) bool {
		if l.writer == nil {
			l.writer = writer
		} else {
			l.writer = io.MultiWriter(l.writer, writer)
		}
		l.updateMode()
		g.logger.Store(lev, l)
		return true
	})

	return g
}

// SetLevelColor sets the color for each level
func (g *Glg) SetLevelColor(level LEVEL, color func(string) string) *Glg {
	l, ok := g.logger.Load(level)
	if ok {
		l.color = color
		g.logger.Store(level, l)
	}

	return g
}

// SetLevelWriter sets writer to glg std writer per logging level
func (g *Glg) SetLevelWriter(level LEVEL, writer io.Writer) *Glg {
	if writer == nil {
		return g
	}

	l, ok := g.logger.Load(level)
	if ok {
		l.writer = writer
		l.updateMode()
		g.logger.Store(level, l)
	}

	return g
}

// AddLevelWriter adds writer to glg std writer per logging level
func (g *Glg) AddLevelWriter(level LEVEL, writer io.Writer) *Glg {
	if writer == nil {
		return g
	}

	l, ok := g.logger.Load(level)
	if ok {
		if l.writer != nil {
			l.writer = io.MultiWriter(l.writer, writer)
		} else {
			l.writer = writer
		}
		l.updateMode()
		g.logger.Store(level, l)
	}

	return g
}

// AddStdLevel adds std log level and returns LEVEL
func (g *Glg) AddStdLevel(tag string, mode MODE, isColor bool) *Glg {
	return g.addLevel(tag, mode, isColor, os.Stdout)
}

// AddErrLevel adds error log level and returns LEVEL
func (g *Glg) AddErrLevel(tag string, mode MODE, isColor bool) *Glg {
	return g.addLevel(tag, mode, isColor, os.Stderr)
}

func (g *Glg) addLevel(tag string, mode MODE, isColor bool, std io.Writer) *Glg {
	lev := LEVEL(atomic.AddUint32(g.levelCounter, 1))
	tag = strings.ToUpper(tag)
	g.levelMap.Store(tag, lev)
	l := &logger{
		writer:   nil,
		std:      std,
		color:    Colorless,
		isColor:  isColor,
		mode:     mode,
		prevMode: mode,
		tag:      tag,
		rawtag:   []byte(lsep + tag + sep),
	}
	l.updateMode()
	g.logger.Store(lev, l)
	return g
}

// EnableTimestamp enables timestamp output
func (g *Glg) EnableTimestamp() *Glg {
	g.logger.Range(func(lev LEVEL, l *logger) bool {
		l.disableTimestamp = false
		g.logger.Store(lev, l)
		return true
	})

	return g
}

// DisableTimestamp disables timestamp output
func (g *Glg) DisableTimestamp() *Glg {
	g.logger.Range(func(lev LEVEL, l *logger) bool {
		l.disableTimestamp = true
		g.logger.Store(lev, l)
		return true
	})

	return g
}

// EnableLevelTimestamp enables timestamp output
func (g *Glg) EnableLevelTimestamp(lv LEVEL) *Glg {
	l, ok := g.logger.Load(lv)
	if ok {
		l.disableTimestamp = false
		g.logger.Store(lv, l)
	}
	return g
}

// DisableLevelTimestamp disables timestamp output
func (g *Glg) DisableLevelTimestamp(lv LEVEL) *Glg {
	l, ok := g.logger.Load(lv)
	if ok {
		l.disableTimestamp = true
		g.logger.Store(lv, l)
	}
	return g
}

// SetCallerDepth configures output line trace caller depth
func (g *Glg) SetCallerDepth(depth int) *Glg {
	if depth > DefaultCallerDepth {
		g.callerDepth = depth
	}
	return g
}

// SetLineTraceMode configures output line traceFlag
func (g *Glg) SetLineTraceMode(mode traceMode) *Glg {
	g.logger.Range(func(lev LEVEL, l *logger) bool {
		l.traceMode = mode
		g.logger.Store(lev, l)
		return true
	})
	return g
}

// SetLevelLineTraceMode configures output line traceFlag
func (g *Glg) SetLevelLineTraceMode(lv LEVEL, mode traceMode) *Glg {
	l, ok := g.logger.Load(lv)
	if ok {
		l.traceMode = mode
		g.logger.Store(lv, l)
	}
	return g
}

// EnableColor enables color output
func (g *Glg) EnableColor() *Glg {
	g.logger.Range(func(lev LEVEL, l *logger) bool {
		l.isColor = true
		l.updateMode()
		g.logger.Store(lev, l)
		return true
	})

	return g
}

// DisableColor disables color output
func (g *Glg) DisableColor() *Glg {
	g.logger.Range(func(lev LEVEL, l *logger) bool {
		l.isColor = false
		l.updateMode()
		g.logger.Store(lev, l)
		return true
	})

	return g
}

// EnableLevelColor enables color output
func (g *Glg) EnableLevelColor(lv LEVEL) *Glg {
	l, ok := g.logger.Load(lv)
	if ok {
		l.isColor = true
		l.updateMode()
		g.logger.Store(lv, l)
	}
	return g
}

// DisableLevelColor disables color output
func (g *Glg) DisableLevelColor(lv LEVEL) *Glg {
	l, ok := g.logger.Load(lv)
	if ok {
		l.isColor = false
		l.updateMode()
		g.logger.Store(lv, l)
	}
	return g
}

// RawString returns raw log string exclude time & tags
func (g *Glg) RawString(data []byte) string {
	str := *(*string)(unsafe.Pointer(&data))
	return str[strings.Index(str, sep)+sepl : len(str)-rcl]
}

// RawString returns raw log string exclude time & tags
func RawString(data []byte) string {
	return glg.RawString(data)
}

// Atol converts level string to Glg.LEVEL
func (g *Glg) Atol(tag string) LEVEL {
	return g.TagStringToLevel(tag)
}

// Atol converts level string to Glg.LEVEL
func Atol(tag string) LEVEL {
	return glg.TagStringToLevel(tag)

}

// TagStringToLevel converts level string to Glg.LEVEL
func (g *Glg) TagStringToLevel(tag string) LEVEL {
	tag = strings.TrimSpace(strings.ToUpper(tag))
	lv, ok := g.levelMap.Load(tag)
	if ok {
		return lv
	}
	switch tag {
	case DEBG.String(), "DBG", "DEBUG", "D":
		return DEBG
	case TRACE.String(), "TRC", "TRA", "TR", "T":
		return TRACE
	case PRINT.String(), "PRINT", "PNT", "P":
		return PRINT
	case LOG.String(), "LO", "LG", "L":
		return LOG
	case INFO.String(), "IFO", "INF", "I":
		return INFO
	case OK.String(), "O", "K":
		return OK
	case WARN.String(), "WARNING", "WRN", "W":
		return WARN
	case ERR.String(), "ERROR", "ER", "E":
		return ERR
	case FAIL.String(), "FAILED", "FI":
		return FAIL
	case FATAL.String(), "FAT", "FL", "F":
		return FATAL
	}
	return UNKNOWN
}

// TagStringToLevel converts level string to glg.LEVEL
func TagStringToLevel(tag string) LEVEL {
	return glg.TagStringToLevel(tag)
}

// FileWriter generates *osFile -> io.Writer
func FileWriter(path string, perm os.FileMode) *os.File {
	if path == "" {
		return nil
	}

	var err error
	var file *os.File
	if _, err = os.Stat(path); err != nil {
		if _, err = os.Stat(filepath.Dir(path)); err != nil {
			err = os.MkdirAll(filepath.Dir(path), perm)
			if err != nil {
				return nil
			}
		}
		file, err = os.Create(path)
		if err != nil {
			return nil
		}

		err = file.Close()
		if err != nil {
			return nil
		}
	}

	file, err = os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, perm)

	if err != nil {
		return nil
	}

	return file
}

// HTTPLogger is simple http access logger
func (g *Glg) HTTPLogger(name string, handler http.Handler) http.Handler {
	return g.HTTPLoggerFunc(name, handler.ServeHTTP)
}

// HTTPLoggerFunc is simple http access logger
func (g *Glg) HTTPLoggerFunc(name string, hf http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := fastime.UnixNanoNow()

		hf(w, r)

		start -= fastime.UnixNanoNow()

		err := g.Logf("Method: %s\tURI: %s\tName: %s\tTime: %s",
			r.Method, r.RequestURI, name, (*(*time.Duration)(unsafe.Pointer(&start))).String())
		if err != nil {
			err = g.Error(err)
			if err != nil {
				fmt.Println(err)
			}
		}
	})
}

// HTTPLogger is simple http access logger
func HTTPLogger(name string, handler http.Handler) http.Handler {
	return glg.HTTPLogger(name, handler)
}

// HTTPLoggerFunc is simple http access logger
func HTTPLoggerFunc(name string, hf http.HandlerFunc) http.Handler {
	return glg.HTTPLoggerFunc(name, hf)
}

// Colorless returns colorless string
func Colorless(str string) string {
	return str
}

// Red returns red colored string
func Red(str string) string {
	return "\033[31m" + str + "\033[39m"
}

// Green returns green colored string
func Green(str string) string {
	return "\033[32m" + str + "\033[39m"
}

// Orange returns orange colored string
func Orange(str string) string {
	return "\033[33m" + str + "\033[39m"
}

// Purple returns purple colored string
func Purple(str string) string {
	return "\033[34m" + str + "\033[39m"
}

// Cyan returns cyan colored string
func Cyan(str string) string {
	return "\033[36m" + str + "\033[39m"
}

// Yellow returns yellow colored string
func Yellow(str string) string {
	return "\033[93m" + str + "\033[39m"
}

// Brown returns Brown colored string
func Brown(str string) string {
	return "\033[96m" + str + "\033[39m"
}

// Gray returns Gray colored string
func Gray(str string) string {
	return "\033[90m" + str + "\033[39m"
}

// Black returns Black colored string
func Black(str string) string {
	return "\033[30m" + str + "\033[39m"
}

// White returns white colored string
func White(str string) string {
	return "\033[97m" + str + "\033[39m"
}

func (g *Glg) out(level LEVEL, format string, val ...interface{}) error {
	log, ok := g.logger.Load(level)
	if !ok {
		return fmt.Errorf("error:\tLog Level %d Not Found", level)
	}

	if log.mode == NONE {
		return nil
	}

	var fl string
	if log.traceMode&(TraceLineLong|TraceLineShort) != 0 {
		_, file, line, ok := runtime.Caller(g.callerDepth)
		switch {
		case !ok:
			fl = "???:0"
		case log.traceMode&TraceLineShort != 0:
			for i := len(file) - 1; i > 0; i-- {
				if file[i] == '/' {
					file = file[i+1:]
					break
				}
			}
			fl = file + ":" + strconv.Itoa(line)
		case strings.HasPrefix(file, runtime.GOROOT()+"/src"):
			fl = "https://github.com/golang/go/blob/" + runtime.Version() + strings.TrimPrefix(file, runtime.GOROOT()) + "#L" + strconv.Itoa(line)
		case strings.Contains(file, "go/pkg/mod/"):
			fl = "https:/"
			for _, path := range strings.Split(strings.SplitN(file, "go/pkg/mod/", 2)[1], "/") {
				if strings.Contains(path, "@") {
					sv := strings.SplitN(path, "@", 2)
					if strings.Count(sv[1], "-") > 2 {
						path = sv[0] + "/blob/master"
					} else {
						path = sv[0] + "/blob/" + sv[1]
					}
				}
				fl += "/" + path
			}
			fl += "#L" + strconv.Itoa(line)
		case strings.Contains(file, "go/src"):
			fl = "https:/"
			cnt := 0
			for _, path := range strings.Split(strings.SplitN(file, "go/src/", 2)[1], "/") {
				if cnt == 3 {
					path = "blob/master/" + path
				}
				fl += "/" + path
				cnt++
			}
			fl += "#L" + strconv.Itoa(line)
		default:
			fl = file + ":" + strconv.Itoa(line)
		}
	}

	if g.enableJSON {
		var w io.Writer
		switch log.writeMode {
		case writeStd, writeColorStd:
			w = log.std
		case writeWriter:
			w = log.writer
		case writeBoth, writeColorBoth:
			w = io.MultiWriter(log.std, log.writer)
		default:
			return nil
		}
		var detail interface{}
		if format != "" {
			detail = fmt.Sprintf(format, val...)
		} else if len(val) > 1 {
			detail = val
		} else {
			detail = val[0]
		}
		var timestamp string
		if !log.disableTimestamp {
			fn := fastime.FormattedNow()
			timestamp = *(*string)(unsafe.Pointer(&fn))
		}
		return json.NewEncoder(w).Encode(JSONFormat{
			Date:   timestamp,
			Level:  log.tag,
			File:   fl,
			Detail: detail,
		})
	}

	var (
		buf []byte
		err error
		b   = g.buffer.Get().(*bytes.Buffer)
	)

	if log.disableTimestamp {
		b.Write(log.rawtag[len(tab):])
	} else {
		b.Write(fastime.FormattedNow())
		b.Write(log.rawtag)
	}
	if len(fl) != 0 {
		b.WriteString("(" + fl + "):\t")
	}
	b.WriteString(format)

	switch {
	case log.writeMode^writeColorStd == 0:
		buf = b.Bytes()
		_, err = fmt.Fprintf(log.std, log.color(*(*string)(unsafe.Pointer(&buf)))+rc, val...)
	case log.writeMode^writeStd == 0:
		b.WriteString(rc)
		buf = b.Bytes()
		_, err = fmt.Fprintf(log.std, *(*string)(unsafe.Pointer(&buf)), val...)
	case log.writeMode^writeWriter == 0:
		b.WriteString(rc)
		buf = b.Bytes()
		_, err = fmt.Fprintf(log.writer, *(*string)(unsafe.Pointer(&buf)), val...)
	case log.writeMode^writeColorBoth == 0:
		buf = b.Bytes()
		str := *(*string)(unsafe.Pointer(&buf))
		_, err = fmt.Fprintf(log.std, log.color(str)+rc, val...)
		if err == nil {
			_, err = fmt.Fprintf(log.writer, str+rc, val...)
		}
	case log.writeMode^writeBoth == 0:
		b.WriteString(rc)
		buf = b.Bytes()
		_, err = fmt.Fprintf(io.MultiWriter(log.std, log.writer), *(*string)(unsafe.Pointer(&buf)), val...)
	}
	bl := uint64(len(buf))
	if atomic.LoadUint64(g.bs) < bl {
		atomic.StoreUint64(g.bs, bl)
	}
	b.Reset()
	g.buffer.Put(b)

	return err
}

// Log writes std log event
func (g *Glg) Log(val ...interface{}) error {
	return g.out(LOG, g.blankFormat(len(val)), val...)
}

// Logf writes std log event with format
func (g *Glg) Logf(format string, val ...interface{}) error {
	return g.out(LOG, format, val...)
}

// LogFunc outputs Log level log returned from the function
func (g *Glg) LogFunc(f func() string) error {
	if g.isModeEnable(LOG) {
		return g.out(LOG, "%s", f())
	}
	return nil
}

// Log writes std log event
func Log(val ...interface{}) error {
	return glg.out(LOG, glg.blankFormat(len(val)), val...)
}

// Logf writes std log event with format
func Logf(format string, val ...interface{}) error {
	return glg.out(LOG, format, val...)
}

// LogFunc outputs Log level log returned from the function
func LogFunc(f func() string) error {
	if isModeEnable(LOG) {
		return glg.out(LOG, "%s", f())
	}
	return nil
}

// Info outputs Info level log
func (g *Glg) Info(val ...interface{}) error {
	return g.out(INFO, g.blankFormat(len(val)), val...)
}

// Infof outputs formatted Info level log
func (g *Glg) Infof(format string, val ...interface{}) error {
	return g.out(INFO, format, val...)
}

// InfoFunc outputs Info level log returned from the function
func (g *Glg) InfoFunc(f func() string) error {
	if g.isModeEnable(INFO) {
		return g.out(INFO, "%s", f())
	}
	return nil
}

// Info outputs Info level log
func Info(val ...interface{}) error {
	return glg.out(INFO, glg.blankFormat(len(val)), val...)
}

// Infof outputs formatted Info level log
func Infof(format string, val ...interface{}) error {
	return glg.out(INFO, format, val...)
}

// InfoFunc outputs Info level log returned from the function
func InfoFunc(f func() string) error {
	if isModeEnable(INFO) {
		return glg.out(INFO, "%s", f())
	}
	return nil
}

// Success outputs Success level log
func (g *Glg) Success(val ...interface{}) error {
	return g.out(OK, g.blankFormat(len(val)), val...)
}

// Successf outputs formatted Success level log
func (g *Glg) Successf(format string, val ...interface{}) error {
	return g.out(OK, format, val...)
}

// SuccessFunc outputs Success level log returned from the function
func (g *Glg) SuccessFunc(f func() string) error {
	if g.isModeEnable(OK) {
		return g.out(OK, "%s", f())
	}
	return nil
}

// Success outputs Success level log
func Success(val ...interface{}) error {
	return glg.out(OK, glg.blankFormat(len(val)), val...)
}

// Successf outputs formatted Success level log
func Successf(format string, val ...interface{}) error {
	return glg.out(OK, format, val...)
}

// SuccessFunc outputs Success level log returned from the function
func SuccessFunc(f func() string) error {
	if isModeEnable(OK) {
		return glg.out(OK, "%s", f())
	}
	return nil
}

// Debug outputs Debug level log
func (g *Glg) Debug(val ...interface{}) error {
	return g.out(DEBG, g.blankFormat(len(val)), val...)
}

// Debugf outputs formatted Debug level log
func (g *Glg) Debugf(format string, val ...interface{}) error {
	return g.out(DEBG, format, val...)
}

// DebugFunc outputs Debug level log returned from the function
func (g *Glg) DebugFunc(f func() string) error {
	if g.isModeEnable(DEBG) {
		return g.out(DEBG, "%s", f())
	}
	return nil
}

// Debug outputs Debug level log
func Debug(val ...interface{}) error {
	return glg.out(DEBG, glg.blankFormat(len(val)), val...)
}

// Debugf outputs formatted Debug level log
func Debugf(format string, val ...interface{}) error {
	return glg.out(DEBG, format, val...)
}

// DebugFunc outputs Debug level log returned from the function
func DebugFunc(f func() string) error {
	if isModeEnable(DEBG) {
		return glg.out(DEBG, "%s", f())
	}
	return nil
}

// Warn outputs Warn level log
func (g *Glg) Warn(val ...interface{}) error {
	return g.out(WARN, g.blankFormat(len(val)), val...)
}

// Warnf outputs formatted Warn level log
func (g *Glg) Warnf(format string, val ...interface{}) error {
	return g.out(WARN, format, val...)
}

// WarnFunc outputs Warn level log returned from the function
func (g *Glg) WarnFunc(f func() string) error {
	if g.isModeEnable(WARN) {
		return g.out(WARN, "%s", f())
	}
	return nil
}

// Warn outputs Warn level log
func Warn(val ...interface{}) error {
	return glg.out(WARN, glg.blankFormat(len(val)), val...)
}

// Warnf outputs formatted Warn level log
func Warnf(format string, val ...interface{}) error {
	return glg.out(WARN, format, val...)
}

// WarnFunc outputs Warn level log returned from the function
func WarnFunc(f func() string) error {
	if isModeEnable(WARN) {
		return glg.out(WARN, "%s", f())
	}
	return nil
}

// CustomLog outputs custom level log
func (g *Glg) CustomLog(level string, val ...interface{}) error {
	return g.out(g.TagStringToLevel(level), g.blankFormat(len(val)), val...)
}

// CustomLogf outputs formatted custom level log
func (g *Glg) CustomLogf(level string, format string, val ...interface{}) error {
	return g.out(g.TagStringToLevel(level), format, val...)
}

// CustomLogFunc outputs custom level log returned from the function
func (g *Glg) CustomLogFunc(level string, f func() string) error {
	lv := g.TagStringToLevel(level)
	if g.isModeEnable(lv) {
		return g.out(lv, "%s", f())
	}
	return nil
}

// CustomLog outputs custom level log
func CustomLog(level string, val ...interface{}) error {
	return glg.out(glg.TagStringToLevel(level), glg.blankFormat(len(val)), val...)
}

// CustomLogf outputs formatted custom level log
func CustomLogf(level string, format string, val ...interface{}) error {
	return glg.out(glg.TagStringToLevel(level), format, val...)
}

// CustomLogFunc outputs custom level log returned from the function
func CustomLogFunc(level string, f func() string) error {
	lv := TagStringToLevel(level)
	if isModeEnable(lv) {
		return glg.out(lv, "%s", f())
	}
	return nil
}

// Trace outputs Trace level log
func (g *Glg) Trace(val ...interface{}) error {
	return g.out(TRACE, g.blankFormat(len(val)), val...)
}

// Tracef outputs formatted Trace level log
func (g *Glg) Tracef(format string, val ...interface{}) error {
	return g.out(TRACE, format, val...)
}

// TraceFunc outputs Trace level log returned from the function
func (g *Glg) TraceFunc(f func() string) error {
	if g.isModeEnable(TRACE) {
		return g.out(TRACE, "%s", f())
	}
	return nil
}

// Trace outputs Trace level log
func Trace(val ...interface{}) error {
	return glg.out(TRACE, glg.blankFormat(len(val)), val...)
}

// Tracef outputs formatted Trace level log
func Tracef(format string, val ...interface{}) error {
	return glg.out(TRACE, format, val...)
}

// TraceFunc outputs Trace log returned from the function
func TraceFunc(f func() string) error {
	if isModeEnable(TRACE) {
		return glg.out(TRACE, "%s", f())
	}
	return nil
}

// Print outputs Print log
func (g *Glg) Print(val ...interface{}) error {
	return g.out(PRINT, g.blankFormat(len(val)), val...)
}

// Println outputs fixed line Print log
func (g *Glg) Println(val ...interface{}) error {
	return g.out(PRINT, g.blankFormat(len(val)), val...)
}

// Printf outputs formatted Print log
func (g *Glg) Printf(format string, val ...interface{}) error {
	return g.out(PRINT, format, val...)
}

// PrintFunc outputs Print log returned from the function
func (g *Glg) PrintFunc(f func() string) error {
	if g.isModeEnable(PRINT) {
		return g.out(PRINT, "%s", f())
	}
	return nil
}

// Print outputs Print log
func Print(val ...interface{}) error {
	return glg.out(PRINT, glg.blankFormat(len(val)), val...)
}

// Println outputs fixed line Print log
func Println(val ...interface{}) error {
	return glg.out(PRINT, glg.blankFormat(len(val)), val...)
}

// Printf outputs formatted Print log
func Printf(format string, val ...interface{}) error {
	return glg.out(PRINT, format, val...)
}

// PrintFunc outputs Print log returned from the function
func PrintFunc(f func() string) error {
	if isModeEnable(PRINT) {
		return glg.out(PRINT, "%s", f())
	}
	return nil
}

// Error outputs Error log
func (g *Glg) Error(val ...interface{}) error {
	return g.out(ERR, g.blankFormat(len(val)), val...)
}

// Errorf outputs formatted Error log
func (g *Glg) Errorf(format string, val ...interface{}) error {
	return g.out(ERR, format, val...)
}

// ErrorFunc outputs Error level log returned from the function
func (g *Glg) ErrorFunc(f func() string) error {
	if g.isModeEnable(ERR) {
		return g.out(ERR, "%s", f())
	}
	return nil
}

// Error outputs Error log
func Error(val ...interface{}) error {
	return glg.out(ERR, glg.blankFormat(len(val)), val...)
}

// Errorf outputs formatted Error log
func Errorf(format string, val ...interface{}) error {
	return glg.out(ERR, format, val...)
}

// ErrorFunc outputs Error level log returned from the function
func ErrorFunc(f func() string) error {
	if isModeEnable(ERR) {
		return glg.out(ERR, "%s", f())
	}
	return nil
}

// Fail outputs Failed log
func (g *Glg) Fail(val ...interface{}) error {
	return g.out(FAIL, g.blankFormat(len(val)), val...)
}

// Failf outputs formatted Failed log
func (g *Glg) Failf(format string, val ...interface{}) error {
	return g.out(FAIL, format, val...)
}

// FailFunc outputs Fail level log returned from the function
func (g *Glg) FailFunc(f func() string) error {
	if g.isModeEnable(FAIL) {
		return g.out(FAIL, "%s", f())
	}
	return nil
}

// Fail outputs Failed log
func Fail(val ...interface{}) error {
	return glg.out(FAIL, glg.blankFormat(len(val)), val...)
}

// Failf outputs formatted Failed log
func Failf(format string, val ...interface{}) error {
	return glg.out(FAIL, format, val...)
}

// FailFunc outputs Fail level log returned from the function
func FailFunc(f func() string) error {
	if isModeEnable(FAIL) {
		return glg.out(FAIL, "%s", f())
	}
	return nil
}

// Fatal outputs Failed log and exit program
func (g *Glg) Fatal(val ...interface{}) {
	err := g.out(FATAL, g.blankFormat(len(val)), val...)
	if err != nil {
		err = g.Error(err.Error())
		if err != nil {
			panic(err)
		}
	}
	exit(1)
}

// Fatalln outputs line fixed Failed log and exit program
func (g *Glg) Fatalln(val ...interface{}) {
	err := g.out(FATAL, g.blankFormat(len(val)), val...)
	if err != nil {
		err = g.Error(err.Error())
		if err != nil {
			panic(err)
		}
	}
	exit(1)
}

// Fatalf outputs formatted Failed log and exit program
func (g *Glg) Fatalf(format string, val ...interface{}) {
	err := g.out(FATAL, format, val...)
	if err != nil {
		err = g.Error(err.Error())
		if err != nil {
			panic(err)
		}
	}
	exit(1)
}

// Fatal outputs Failed log and exit program
func Fatal(val ...interface{}) {
	glg.Fatal(val...)
}

// Fatalf outputs formatted Failed log and exit program
func Fatalf(format string, val ...interface{}) {
	glg.Fatalf(format, val...)
}

// Fatalln outputs line fixed Failed log and exit program
func Fatalln(val ...interface{}) {
	glg.Fatalln(val...)
}

// ReplaceExitFunc replaces exit function.
// If you do not want to start os.Exit at glg.Fatal error,
// use this function to register arbitrary function
func ReplaceExitFunc(fn func(i int)) {
	exit = fn
}

// Reset provides parameter reset function for glg struct instance
func Reset() *Glg {
	glg = glg.Reset()
	return glg
}

// Reset provides parameter reset function for glg struct instance
func (g *Glg) Reset() *Glg {
	g = New()
	return g
}

func (g *Glg) blankFormat(l int) string {
	if g.enableJSON {
		return ""
	}
	if dfl > l {
		return df[:l*3-1]
	}
	format := df
	for c := l / dfl; c >= 0; c-- {
		format += df
	}
	return format[:l*3-1]
}

// isModeEnable returns the level has already turned on the logging
func isModeEnable(l LEVEL) bool {
	return Get().GetCurrentMode(l) != NONE
}

// isModeEnable returns the level has already turned on the logging
func (g *Glg) isModeEnable(l LEVEL) bool {
	return g.GetCurrentMode(l) != NONE
}
