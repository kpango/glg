// Package glg can quickly output that are colored and leveled logs with simple syntax
package glg

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
	"unsafe"
)

// Glg is glg base struct
type Glg struct {
	// user cutom writer
	writer map[string]io.Writer
	// writer for stdout or stderr
	std     map[string]io.Writer
	colors  map[string]func(string) string
	isColor map[string]bool
	mode    map[string]int
	mu      *sync.Mutex
}

const (
	// LOG is log level
	LOG = "LOG"
	// PRINT is print log level
	PRINT = "PRINT"
	// INFO is info log level
	INFO = "INFO"
	// DEBG is debug log level
	DEBG = "DEBG"
	// OK is success notify log level
	OK = "OK"
	// WARN is warning log level
	WARN = "WARN"
	// ERR is error log level
	ERR = "ERR"
	// FAIL is failed log level
	FAIL = "FAIL"
	// FATAL is fatal log level
	FATAL = "FATAL"

	// NONE is disable Logging
	NONE = iota
	// STD is std log mode
	STD
	// BOTH is both log mode
	BOTH
	// WRITER is io.Writer log mode
	WRITER
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

// New returns plain glg instance
func New() *Glg {
	return (&Glg{
		writer: make(map[string]io.Writer),
		std: map[string]io.Writer{
			// standard out
			PRINT: os.Stdout,
			LOG:   os.Stdout,
			INFO:  os.Stdout,
			DEBG:  os.Stdout,
			OK:    os.Stdout,
			WARN:  os.Stdout,
			// error out
			ERR:   os.Stderr,
			FAIL:  os.Stderr,
			FATAL: os.Stderr,
		},
		colors: map[string]func(string) string{
			PRINT: Colorless,
			LOG:   Colorless,
			INFO:  Green,
			DEBG:  Purple,
			OK:    Cyan,
			WARN:  Orange,
			// error out
			ERR:   Red,
			FAIL:  Red,
			FATAL: Red,
		},
		isColor: map[string]bool{
			// standard out
			PRINT: true,
			LOG:   true,
			INFO:  true,
			DEBG:  true,
			OK:    true,
			WARN:  true,
			// error out
			ERR:   true,
			FAIL:  true,
			FATAL: true,
		},
		mode: map[string]int{
			// standard out
			PRINT: STD,
			LOG:   STD,
			INFO:  STD,
			DEBG:  STD,
			OK:    STD,
			WARN:  STD,
			// error out
			ERR:   STD,
			FAIL:  STD,
			FATAL: STD,
		},
		mu: new(sync.Mutex),
	})
}

// Get returns singleton glg instance
func Get() *Glg {
	once.Do(func() {
		glg = New()
	})
	return glg
}

// SetMode sets glg logging mode
func (g *Glg) SetMode(mode int) *Glg {
	g.mu.Lock()
	for level := range g.mode {
		g.mode[level] = mode
	}
	g.mu.Unlock()
	return g
}

// SetLevelMode set glg logging mode per level
func (g *Glg) SetLevelMode(level string, mode int) *Glg {
	g.mu.Lock()
	g.mode[level] = mode
	g.mu.Unlock()
	return g
}

// GetCurrentMode returns current logging mode
func (g *Glg) GetCurrentMode(level string) int {
	return g.mode[level]
}

// InitWriter is initialize glg writer
func (g *Glg) InitWriter() *Glg {
	g.mu.Lock()
	defer g.mu.Unlock()
	g = New()
	return g
}

// SetWriter sets writer to glg std writers
func (g *Glg) SetWriter(writer io.Writer) *Glg {
	if writer == nil {
		return g
	}
	g.mu.Lock()
	if len(g.writer) == 0 {
		for k := range g.std {
			g.writer[k] = writer
		}
	} else {
		for k := range g.writer {
			g.writer[k] = writer
		}
	}
	g.mu.Unlock()
	return g
}

// AddWriter adds writer to glg std writers
func (g *Glg) AddWriter(writer io.Writer) *Glg {
	if writer == nil {
		return g
	}
	g.mu.Lock()
	if len(g.writer) == 0 {
		for k := range g.std {
			g.writer[k] = writer
		}
	} else {
		for k := range g.writer {
			g.writer[k] = io.MultiWriter(g.writer[k], writer)
		}
	}
	g.mu.Unlock()
	return g
}

// SetLevelColor sets the color for each level
func (g *Glg) SetLevelColor(level string, color func(string) string) *Glg {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.colors[level] = color
	return g
}

// SetLevelWriter sets writer to glg std writer per logging level
func (g *Glg) SetLevelWriter(level string, writer io.Writer) *Glg {
	if writer == nil {
		return g
	}
	g.mu.Lock()
	g.writer[level] = writer
	g.mu.Unlock()
	return g
}

// AddLevelWriter adds writer to glg std writer per logging level
func (g *Glg) AddLevelWriter(level string, writer io.Writer) *Glg {
	if writer == nil {
		return g
	}

	g.mu.Lock()
	w, ok := g.writer[level]
	if ok {
		g.writer[level] = io.MultiWriter(w, writer)
	} else {
		g.writer[level] = writer
	}
	g.mu.Unlock()
	return g
}

// AddStdLevel adds std log level
func (g *Glg) AddStdLevel(level string, mode int, isColor bool) *Glg {
	g.mu.Lock()
	g.writer[level] = g.writer[INFO]
	g.std[level] = os.Stdout
	g.mode[level] = mode
	g.colors[level] = Colorless
	g.isColor[level] = isColor
	g.mu.Unlock()
	return g
}

// AddErrLevel adds error log level
func (g *Glg) AddErrLevel(level string, mode int, isColor bool) *Glg {
	g.mu.Lock()
	g.writer[level] = g.writer[ERR]
	g.std[level] = os.Stderr
	g.mode[level] = mode
	g.colors[level] = Red
	g.isColor[level] = isColor
	g.mu.Unlock()
	return g
}

// EnableColor enables color output
func (g *Glg) EnableColor() *Glg {
	g.mu.Lock()
	for level := range g.isColor {
		g.isColor[level] = true
	}
	g.mu.Unlock()
	return g
}

// DisableColor disables color output
func (g *Glg) DisableColor() *Glg {
	g.mu.Lock()
	for level := range g.isColor {
		g.isColor[level] = false
	}
	g.mu.Unlock()
	return g
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

func (g *Glg) out(level, format string, val ...interface{}) error {
	if g.mode[level] == NONE {
		return nil
	}

	var buf = make([]byte, 0, len(level)+len(format)+25)
	buf = append(append(append(append(time.Now().AppendFormat(buf[:0], "2006-01-02 15:04:05"), "\t["...), level...), "]:\t"...), format...)
	var str = *(*string)(unsafe.Pointer(&buf))

	var err error
	if g.mode[level] == STD || g.mode[level] == BOTH {
		_, ok := g.colors[level]
		if g.isColor[level] && ok {
			g.mu.Lock()
			_, err = fmt.Fprintf(g.std[level], g.colors[level](str)+"\n", val...)
			g.mu.Unlock()
		} else {
			g.mu.Lock()
			_, err = fmt.Fprintf(g.std[level], str+"\n", val...)
			g.mu.Unlock()
		}
		if err != nil {
			return err
		}
	}

	if g.mode[level] == WRITER || g.mode[level] == BOTH {
		g.mu.Lock()
		w, ok := g.writer[level]
		g.mu.Unlock()
		if ok && w != nil {
			_, err = fmt.Fprintf(w, str+"\n", val...)
		}
	}
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
	if _, ok := g.std[level]; ok {
		return g.out(level, "%v", val...)
	}
	return fmt.Errorf("Log Level %s Not Found", level)
}

// CustomLogf outputs formatted custom level log
func (g *Glg) CustomLogf(level, format string, val ...interface{}) error {
	if _, ok := g.std[level]; ok {
		return g.out(level, format, val...)
	}
	return fmt.Errorf("Log Level %s Not Found", level)
}

// CustomLog outputs custom level log
func CustomLog(level string, val ...interface{}) error {
	if _, ok := glg.std[level]; ok {
		return glg.out(level, "%v", val...)
	}
	return fmt.Errorf("Log Level %s Not Found", level)
}

// CustomLogf outputs formatted custom level log
func CustomLogf(level, format string, val ...interface{}) error {
	if _, ok := glg.std[level]; ok {
		return glg.out(level, format, val...)
	}
	return fmt.Errorf("Log Level %s Not Found", level)
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
