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
	"net/http"
	"os"
	"path/filepath"
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
	bufferSize   atomic.Value
	logger       loggers
	levelCounter *uint32
	levelMap     levelMap
	buffer       sync.Pool
	enableJSON   bool
}

type JSONFormat struct {
	Date   string
	Level  string
	Detail interface{}
}

// MODE is logging mode (std only, writer only, std & writer)
type MODE uint8

// LEVEL is log level
type LEVEL uint8

type wMode uint8

type logger struct {
	tag       string
	writer    io.Writer
	std       io.Writer
	color     func(string) string
	isColor   bool
	mode      MODE
	writeMode wMode
}

const (
	// LOG is log level
	LOG LEVEL = iota
	// PRINT is print log level
	PRINT
	// INFO is info log level
	INFO
	// DEBG is debug log level
	DEBG
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

	// NONE is disable Logging
	NONE MODE = iota
	// STD is std log mode
	STD
	// BOTH is both log mode
	BOTH
	// WRITER is io.Writer log mode
	WRITER

	// Internal writeMode
	writeColorStd wMode = iota
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

	sep  = "]:\t"
	sepl = len(sep)
)

var (
	glg  *Glg
	once sync.Once

	// exit for Faltal error
	exit = os.Exit

	bufferSize = 500
)

func init() {
	Get()
}

func (l LEVEL) String() string {
	switch l {
	case LOG:
		return "LOG"
	case PRINT:
		return "PRINT"
	case INFO:
		return "INFO"
	case DEBG:
		return "DEBG"
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
	}

	g.bufferSize.Store(bufferSize)

	g.buffer = sync.Pool{
		New: func() interface{} {
			return bytes.NewBuffer(make([]byte, 0, g.bufferSize.Load().(int)))
		},
	}

	atomic.StoreUint32(g.levelCounter, uint32(FATAL))

	for lev, log := range map[LEVEL]*logger{
		// standard out
		PRINT: {
			std:     os.Stdout,
			color:   Colorless,
			isColor: true,
			mode:    STD,
		},
		LOG: {
			std:     os.Stdout,
			color:   Colorless,
			isColor: true,
			mode:    STD,
		},
		INFO: {
			std:     os.Stdout,
			color:   Green,
			isColor: true,
			mode:    STD,
		},
		DEBG: {
			std:     os.Stdout,
			color:   Purple,
			isColor: true,
			mode:    STD,
		},
		OK: {
			std:     os.Stdout,
			color:   Cyan,
			isColor: true,
			mode:    STD,
		},
		WARN: {
			std:     os.Stdout,
			color:   Orange,
			isColor: true,
			mode:    STD,
		},
		// error out
		ERR: {
			std:     os.Stderr,
			color:   Red,
			isColor: true,
			mode:    STD,
		},
		FAIL: {
			std:     os.Stderr,
			color:   Red,
			isColor: true,
			mode:    STD,
		},
		FATAL: {
			std:     os.Stderr,
			color:   Red,
			isColor: true,
			mode:    STD,
		},
	} {
		log.tag = lev.String()
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

// SetMode sets glg logging mode
func (g *Glg) SetMode(mode MODE) *Glg {
	g.logger.Range(func(lev LEVEL, l *logger) bool {
		l.mode = mode
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

func (g *Glg) SetLogBufferSize(size int) *Glg {
	if size > (len(timeFormat) + len("[") + len(sep)) {
		g.bufferSize.Store(size)
	}
	return g
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
	atomic.AddUint32(g.levelCounter, 1)
	lev := LEVEL(atomic.LoadUint32(g.levelCounter))
	tag = strings.ToUpper(tag)
	g.levelMap.Store(tag, lev)
	l := &logger{
		writer:  nil,
		std:     os.Stdout,
		color:   Colorless,
		isColor: isColor,
		mode:    mode,
		tag:     tag,
	}
	l.updateMode()
	g.logger.Store(lev, l)
	return g
}

// AddErrLevel adds error log level and returns LEVEL
func (g *Glg) AddErrLevel(tag string, mode MODE, isColor bool) *Glg {
	atomic.AddUint32(g.levelCounter, 1)
	lev := LEVEL(atomic.LoadUint32(g.levelCounter))
	tag = strings.ToUpper(tag)
	g.levelMap.Store(tag, lev)
	l := &logger{
		writer:  nil,
		std:     os.Stderr,
		color:   Red,
		isColor: isColor,
		mode:    mode,
		tag:     tag,
	}
	l.updateMode()
	g.logger.Store(lev, l)
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

// TagStringToLevel converts level string to Glg.LEVEL
func (g *Glg) TagStringToLevel(tag string) LEVEL {
	tag = strings.ToUpper(tag)
	lv, ok := g.levelMap.Load(tag)
	if !ok {
		return 255
	}
	return lv
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
		return fmt.Errorf("error:\tLog Level %s Not Found", level)
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
		} else {
			detail = val
		}
		return json.NewEncoder(w).Encode(JSONFormat{
			Date:   string(fastime.FormattedNow()),
			Level:  level.String(),
			Detail: detail,
		})
	}

	var (
		buf []byte
		err error
		b   = g.buffer.Get().(*bytes.Buffer)
	)

	b.Write(fastime.FormattedNow())
	b.WriteString("\t[")
	b.WriteString(log.tag)
	b.WriteString(sep)
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
		var str = *(*string)(unsafe.Pointer(&buf))
		_, err = fmt.Fprintf(log.std, log.color(str)+rc, val...)
		if err == nil {
			_, err = fmt.Fprintf(log.writer, str+rc, val...)
		}
	case log.writeMode^writeBoth == 0:
		b.WriteString(rc)
		buf = b.Bytes()
		_, err = fmt.Fprintf(io.MultiWriter(log.std, log.writer), *(*string)(unsafe.Pointer(&buf)), val...)
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
