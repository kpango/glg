// Package glg can quickly output that are colored and leveled logs with simple syntax
package glg

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

// Glg is glg base struct
type Glg struct {
	logger       sync.Map      // map[uint8]*logger
	timer        *atomic.Value // []byte
	levelCounter *uint32
	levelMap     sync.Map
	buffer       sync.Pool
	cancel       context.CancelFunc
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
)

var (
	glg  *Glg
	once sync.Once

	// exit for Faltal error
	exit = os.Exit

	bufferSize = 2000
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
	case l.mode == BOTH && l.isColor && l.writer == nil:
		l.writeMode = writeColorStd
	case l.mode == BOTH && l.isColor && l.writer != nil:
		l.writeMode = writeColorBoth
	case l.mode == BOTH && !l.isColor && l.writer == nil:
		l.writeMode = writeStd
	case l.mode == BOTH && !l.isColor && l.writer != nil:
		l.writeMode = writeBoth
	case l.mode == STD && l.isColor:
		l.writeMode = writeColorStd
	case l.mode == STD && !l.isColor:
		l.writeMode = writeStd
	default:
		l.writeMode = none
	}
	return l
}

// New returns plain glg instance
func New() *Glg {
	m := map[LEVEL]*logger{
		// standard out
		PRINT: &logger{
			std:     os.Stdout,
			color:   Colorless,
			isColor: true,
			mode:    STD,
		},
		LOG: &logger{
			std:     os.Stdout,
			color:   Colorless,
			isColor: true,
			mode:    STD,
		},
		INFO: &logger{
			std:     os.Stdout,
			color:   Green,
			isColor: true,
			mode:    STD,
		},
		DEBG: &logger{
			std:     os.Stdout,
			color:   Purple,
			isColor: true,
			mode:    STD,
		},
		OK: &logger{
			std:     os.Stdout,
			color:   Cyan,
			isColor: true,
			mode:    STD,
		},
		WARN: &logger{
			std:     os.Stdout,
			color:   Orange,
			isColor: true,
			mode:    STD,
		},
		// error out
		ERR: &logger{
			std:     os.Stderr,
			color:   Red,
			isColor: true,
			mode:    STD,
		},
		FAIL: &logger{
			std:     os.Stderr,
			color:   Red,
			isColor: true,
			mode:    STD,
		},
		FATAL: &logger{
			std:     os.Stderr,
			color:   Red,
			isColor: true,
			mode:    STD,
		},
	}

	g := &Glg{
		levelCounter: new(uint32),
		buffer: sync.Pool{
			New: func() interface{} {
				return make([]byte, bufferSize)
			},
		},
	}

	atomic.StoreUint32(g.levelCounter, uint32(FATAL))

	for k, v := range m {
		v.tag = k.String()
		v.updateMode()
		g.logger.Store(k, v)
	}

	return g.startTimerD()
}

// Get returns singleton glg instance
func Get() *Glg {
	once.Do(func() {
		glg = New()
	})
	return glg
}

func (g *Glg) startTimerD() *Glg {
	g.timer = new(atomic.Value)
	timeFormat := "2006-01-02 15:04:05"

	buf := g.buffer.Get().([]byte)
	g.timer.Store(time.Now().AppendFormat(buf[:0], timeFormat))
	g.buffer.Put(buf[:0])

	var ctx context.Context
	ctx, g.cancel = context.WithCancel(context.Background())

	go func() {
		ticker := time.NewTicker(time.Millisecond * 500)
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				buf := g.buffer.Get().([]byte)
				g.timer.Store(time.Now().AppendFormat(buf[:0], timeFormat))
				g.buffer.Put(buf[:0])
			}
		}
	}()

	return g
}

// Stop stops glg timer daemon
func (g *Glg) Stop() *Glg {
	g.cancel()
	return g
}

// SetMode sets glg logging mode
func (g *Glg) SetMode(mode MODE) *Glg {
	g.logger.Range(func(key, val interface{}) bool {
		l := val.(*logger)
		l.mode = mode
		l.updateMode()
		g.logger.Store(key.(LEVEL), l)
		return true
	})

	return g
}

// SetLevelMode set glg logging mode* per level
func (g *Glg) SetLevelMode(level LEVEL, mode MODE) *Glg {
	lev, ok := g.logger.Load(level)
	if ok {
		l := lev.(*logger)
		l.mode = mode
		l.updateMode()
		g.logger.Store(level, l)
	}

	return g
}

// SetPrefix set Print logger prefix
func (g *Glg) SetPrefix(pref string) *Glg {
	v, ok := g.logger.Load(PRINT)
	if ok {
		value := v.(*logger)
		value.tag = pref
		g.logger.Store(PRINT, value)
	}
	return g
}

// GetCurrentMode returns current logging mode
func (g *Glg) GetCurrentMode(level LEVEL) MODE {
	l, ok := g.logger.Load(level)
	if ok {
		return l.(*logger).mode
	}
	return NONE
}

// InitWriter is initialize glg writer
func (g *Glg) InitWriter() *Glg {
	g.logger.Range(func(key, val interface{}) bool {
		l := val.(*logger)
		l.writer = nil
		l.updateMode()
		g.logger.Store(key.(LEVEL), l)
		return true
	})
	return g
}

// SetWriter sets writer to glg std writers
func (g *Glg) SetWriter(writer io.Writer) *Glg {
	if writer == nil {
		return g
	}

	g.logger.Range(func(key, val interface{}) bool {
		l := val.(*logger)
		l.writer = writer
		l.updateMode()
		g.logger.Store(key.(LEVEL), l)
		return true
	})

	return g
}

// AddWriter adds writer to glg std writers
func (g *Glg) AddWriter(writer io.Writer) *Glg {
	if writer == nil {
		return g
	}

	g.logger.Range(func(key, val interface{}) bool {
		l := val.(*logger)
		if l.writer == nil {
			l.writer = writer
		} else {
			l.writer = io.MultiWriter(l.writer, writer)
		}
		l.updateMode()
		g.logger.Store(key.(LEVEL), l)
		return true
	})

	return g
}

// SetLevelColor sets the color for each level
func (g *Glg) SetLevelColor(level LEVEL, color func(string) string) *Glg {
	lev, ok := g.logger.Load(level)
	if ok {
		l := lev.(*logger)
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

	lev, ok := g.logger.Load(level)
	if ok {
		l := lev.(*logger)
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

	lev, ok := g.logger.Load(level)
	if ok {
		l := lev.(*logger)
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

	g.logger.Range(func(key, val interface{}) bool {
		l := val.(*logger)
		l.isColor = true
		l.updateMode()
		g.logger.Store(key.(LEVEL), l)
		return true
	})

	return g
}

// DisableColor disables color output
func (g *Glg) DisableColor() *Glg {

	g.logger.Range(func(key, val interface{}) bool {
		l := val.(*logger)
		l.isColor = false
		l.updateMode()
		g.logger.Store(key.(LEVEL), l)
		return true
	})

	return g
}

// EnableLevelColor enables color output
func (g *Glg) EnableLevelColor(lv LEVEL) *Glg {
	ins, ok := g.logger.Load(lv)
	if ok {
		l := ins.(*logger)
		l.isColor = true
		l.updateMode()
		g.logger.Store(lv, l)
	}
	return g
}

// DisableLevelColor disables color output
func (g *Glg) DisableLevelColor(lv LEVEL) *Glg {
	ins, ok := g.logger.Load(lv)
	if ok {
		l := ins.(*logger)
		l.isColor = false
		l.updateMode()
		g.logger.Store(lv, l)
	}
	return g
}

// TagStringToLevel converts level string to Glg.LEVEL
func (g *Glg) TagStringToLevel(tag string) LEVEL {
	l, ok := g.levelMap.Load(tag)
	if !ok {
		return 255
	}
	return l.(LEVEL)
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
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		handler.ServeHTTP(w, r)

		err := g.Logf("Method: %s\tURI: %s\tName: %s\tTime: %s",
			r.Method, r.RequestURI, name, time.Since(start).String())

		if err != nil {
			err = g.Error(err)
			if err != nil {
				fmt.Println(err)
			}
		}
	})
}

// HTTPLoggerFunc is simple http access logger
func (g *Glg) HTTPLoggerFunc(name string, hf http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		hf(w, r)

		err := g.Logf("Method: %s\tURI: %s\tName: %s\tTime: %s",
			r.Method, r.RequestURI, name, time.Since(start).String())

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

// Colorless return colorless string
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
	l, ok := g.logger.Load(level)
	if !ok {
		return fmt.Errorf("Log Level %s Not Found", level)
	}
	log := l.(*logger)

	var err error
	var buf = g.buffer.Get().([]byte)
	switch log.writeMode {
	case writeColorStd:
		buf = append(append(append(append(g.timer.Load().([]byte), "\t["...), log.tag...), "]:\t"...), format...)
		_, err = fmt.Fprintf(log.std, log.color(*(*string)(unsafe.Pointer(&buf)))+"\n", val...)
	case writeStd:
		buf = append(append(append(append(append(g.timer.Load().([]byte), "\t["...), log.tag...), "]:\t"...), format...), "\n"...)
		_, err = fmt.Fprintf(log.std, *(*string)(unsafe.Pointer(&buf)), val...)
	case writeWriter:
		buf = append(append(append(append(append(g.timer.Load().([]byte), "\t["...), log.tag...), "]:\t"...), format...), "\n"...)
		_, err = fmt.Fprintf(log.writer, *(*string)(unsafe.Pointer(&buf)), val...)
	case writeColorBoth:
		buf = append(append(append(append(g.timer.Load().([]byte), "\t["...), log.tag...), "]:\t"...), format...)
		var str = *(*string)(unsafe.Pointer(&buf))
		_, err = fmt.Fprintf(log.std, log.color(str)+"\n", val...)
		_, err = fmt.Fprintf(log.writer, str+"\n", val...)
	case writeBoth:
		buf = append(append(append(append(append(g.timer.Load().([]byte), "\t["...), log.tag...), "]:\t"...), format...), "\n"...)
		_, err = fmt.Fprintf(io.MultiWriter(log.std, log.writer), *(*string)(unsafe.Pointer(&buf)), val...)
	}
	g.buffer.Put(buf[:0])

	return err
}

// Log writes std log event
func (g *Glg) Log(val ...interface{}) error {
	return g.out(LOG, "%v", val...)
}

// Logf writes std log event with format
func (g *Glg) Logf(format string, val ...interface{}) error {
	return g.out(LOG, format, val...)
}

// Log writes std log event
func Log(val ...interface{}) error {
	return glg.out(LOG, "%v", val...)
}

// Logf writes std log event with format
func Logf(format string, val ...interface{}) error {
	return glg.out(LOG, format, val...)
}

// Info outputs Info level log
func (g *Glg) Info(val ...interface{}) error {
	return g.out(INFO, "%v", val...)
}

// Infof outputs formatted Info level log
func (g *Glg) Infof(format string, val ...interface{}) error {
	return g.out(INFO, format, val...)
}

// Info outputs Info level log
func Info(val ...interface{}) error {
	return glg.out(INFO, "%v", val...)
}

// Infof outputs formatted Info level log
func Infof(format string, val ...interface{}) error {
	return glg.out(INFO, format, val...)
}

// Success outputs Success level log
func (g *Glg) Success(val ...interface{}) error {
	return g.out(OK, "%v", val...)
}

// Successf outputs formatted Success level log
func (g *Glg) Successf(format string, val ...interface{}) error {
	return g.out(OK, format, val...)
}

// Success outputs Success level log
func Success(val ...interface{}) error {
	return glg.out(OK, "%v", val...)
}

// Successf outputs formatted Success level log
func Successf(format string, val ...interface{}) error {
	return glg.out(OK, format, val...)
}

// Debug outputs Debug level log
func (g *Glg) Debug(val ...interface{}) error {
	return g.out(DEBG, "%v", val...)
}

// Debugf outputs formatted Debug level log
func (g *Glg) Debugf(format string, val ...interface{}) error {
	return g.out(DEBG, format, val...)
}

// Debug outputs Debug level log
func Debug(val ...interface{}) error {
	return glg.out(DEBG, "%v", val...)
}

// Debugf outputs formatted Debug level log
func Debugf(format string, val ...interface{}) error {
	return glg.out(DEBG, format, val...)
}

// Warn outputs Warn level log
func (g *Glg) Warn(val ...interface{}) error {
	return g.out(WARN, "%v", val...)
}

// Warnf outputs formatted Warn level log
func (g *Glg) Warnf(format string, val ...interface{}) error {
	return g.out(WARN, format, val...)
}

// Warn outputs Warn level log
func Warn(val ...interface{}) error {
	return glg.out(WARN, "%v", val...)
}

// Warnf outputs formatted Warn level log
func Warnf(format string, val ...interface{}) error {
	return glg.out(WARN, format, val...)
}

// CustomLog outputs custom level log
func (g *Glg) CustomLog(level string, val ...interface{}) error {
	return g.out(g.TagStringToLevel(level), "%v", val...)
}

// CustomLogf outputs formatted custom level log
func (g *Glg) CustomLogf(level string, format string, val ...interface{}) error {
	return g.out(g.TagStringToLevel(level), format, val...)
}

// CustomLog outputs custom level log
func CustomLog(level string, val ...interface{}) error {
	return glg.out(glg.TagStringToLevel(level), "%v", val...)
}

// CustomLogf outputs formatted custom level log
func CustomLogf(level string, format string, val ...interface{}) error {
	return glg.out(glg.TagStringToLevel(level), format, val...)
}

// Print outputs Print log
func (g *Glg) Print(val ...interface{}) error {
	return g.out(PRINT, "%v", val...)
}

// Println outputs fixed line Print log
func (g *Glg) Println(val ...interface{}) error {
	return g.out(PRINT, "%v\n", val...)
}

// Printf outputs formatted Print log
func (g *Glg) Printf(format string, val ...interface{}) error {
	return g.out(PRINT, format, val...)
}

// Print outputs Print log
func Print(val ...interface{}) error {
	return glg.out(PRINT, "%v", val...)
}

// Println outputs fixed line Print log
func Println(val ...interface{}) error {
	return glg.out(PRINT, "%v\n", val...)
}

// Printf outputs formatted Print log
func Printf(format string, val ...interface{}) error {
	return glg.out(PRINT, format, val...)
}

// Error outputs Error log
func (g *Glg) Error(val ...interface{}) error {
	return g.out(ERR, "%v", val...)
}

// Errorf outputs formatted Error log
func (g *Glg) Errorf(format string, val ...interface{}) error {
	return g.out(ERR, format, val...)
}

// Error outputs Error log
func Error(val ...interface{}) error {
	return glg.out(ERR, "%v", val...)
}

// Errorf outputs formatted Error log
func Errorf(format string, val ...interface{}) error {
	return glg.out(ERR, format, val...)
}

// Fail outputs Failed log
func (g *Glg) Fail(val ...interface{}) error {
	return g.out(FAIL, "%v", val...)
}

// Failf outputs formatted Failed log
func (g *Glg) Failf(format string, val ...interface{}) error {
	return g.out(FAIL, format, val...)
}

// Fail outputs Failed log
func Fail(val ...interface{}) error {
	return glg.out(FAIL, "%v", val...)
}

// Failf outputs formatted Failed log
func Failf(format string, val ...interface{}) error {
	return glg.out(FAIL, format, val...)
}

// Fatal outputs Failed log and exit program
func (g *Glg) Fatal(val ...interface{}) {
	g.Fatalf("%v", val...)
}

// Fatalln outputs line fixed Failed log and exit program
func (g *Glg) Fatalln(val ...interface{}) {
	g.Fatalf("%v\n", val...)
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
	glg.Fatalf("%v", val...)
}

// Fatalf outputs formatted Failed log and exit program
func Fatalf(format string, val ...interface{}) {
	glg.Fatalf(format, val...)
}

// Fatalln outputs line fixed Failed log and exit program
func Fatalln(val ...interface{}) {
	glg.Fatalf("%v\n", val...)
}
