package glg

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"
)

type ExitError int

func (e ExitError) Error() string {
	return fmt.Sprintf("exited with code %d", int(e))
}

func init() {
	exit = func(n int) {
		panic(ExitError(n))
	}
}

func testExit(code int, f func()) (err error) {
	defer func() {
		e := recover()
		switch t := e.(type) {
		case ExitError:
			if int(t) == code {
				err = nil
			} else {
				err = fmt.Errorf("expected exit with %v but %v", code, e)
			}
		default:
			err = fmt.Errorf("expected exit with %v but %v", code, e)
		}
	}()
	f()
	return errors.New("expected exited but not")
}

func TestLEVEL_String(t *testing.T) {
	l := LEVEL(10)
	if l.String() != "" {
		t.Error("invalid value")
	}
}

func TestNew(t *testing.T) {
	t.Run("Comparing simple instances", func(t *testing.T) {
		ins1 := New()
		ins2 := New()
		if ins1.GetCurrentMode(LOG) != ins2.GetCurrentMode(LOG) {
			t.Errorf("glg mode = %v, want %v", ins1.GetCurrentMode(LOG), ins2.GetCurrentMode(LOG))
		}

		ins1.logger.Range(func(key, val interface{}) bool {
			lev1, ok := val.(*logger)
			if !ok {
				t.Errorf("invalid glg instance 1 type:\t%v", val)
			}
			i2, ok := ins2.logger.Load(key.(LEVEL))
			if !ok {
				t.Error("glg instance 2 not found")
			}
			lev2, ok := i2.(*logger)
			if !ok {
				t.Errorf("invalid glg instance 2 type:\t%v", val)
			}
			if lev1.tag != lev2.tag || lev1.mode != lev2.mode {
				t.Errorf("Expect %v, want %v", lev2, lev1)
				return false
			}
			return true
		})
	})
}

func TestGet(t *testing.T) {
	t.Run("Comparing singleton instances", func(t *testing.T) {
		ins1 := Get()
		ins2 := Get()

		if !reflect.DeepEqual(ins1, ins2) {
			t.Errorf("Expect %v, want %v", ins2, ins1)
		}
		ins1.logger.Range(func(key, val interface{}) bool {
			lev1, ok := val.(*logger)
			if !ok {
				t.Errorf("invalid glg instance 1 type:\t%v", val)
			}
			i2, ok := ins2.logger.Load(key.(LEVEL))
			if !ok {
				t.Error("glg instance 2 not found")
			}
			lev2, ok := i2.(*logger)
			if !ok {
				t.Errorf("invalid glg instance 2 type:\t%v", val)
			}
			if !reflect.DeepEqual(lev1, lev2) {
				t.Errorf("Expect %v, want %v", lev2, lev1)
				return false
			}
			return true
		})
	})
}

func TestGlg_SetMode(t *testing.T) {
	tests := []struct {
		name  string
		mode  MODE
		want  MODE
		isErr bool
	}{
		{
			name:  "std",
			mode:  STD,
			want:  STD,
			isErr: false,
		},
		{
			name:  "writer",
			mode:  WRITER,
			want:  WRITER,
			isErr: false,
		},
		{
			name:  "both",
			mode:  BOTH,
			want:  BOTH,
			isErr: false,
		},
		{
			name:  "none",
			mode:  NONE,
			want:  NONE,
			isErr: false,
		},
		{
			name:  "writer-both",
			mode:  WRITER,
			want:  BOTH,
			isErr: true,
		},
		{
			name:  "different mode",
			mode:  NONE,
			want:  STD,
			isErr: true,
		},
	}
	g := New()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := g.SetMode(tt.mode).GetCurrentMode(LOG); !reflect.DeepEqual(got, tt.want) && !tt.isErr {
				t.Errorf("Glg.SetMode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGlg_SetLevelMode(t *testing.T) {
	tests := []struct {
		name  string
		mode  MODE
		want  MODE
		level LEVEL
		isErr bool
	}{
		{
			name:  "std",
			mode:  STD,
			want:  STD,
			level: LOG,
			isErr: false,
		},
		{
			name:  "writer",
			mode:  WRITER,
			want:  WRITER,
			level: LOG,
			isErr: false,
		},
		{
			name:  "both",
			mode:  BOTH,
			want:  BOTH,
			level: LOG,
			isErr: false,
		},
		{
			name:  "none",
			mode:  NONE,
			want:  NONE,
			level: LOG,
			isErr: false,
		},
		{
			name:  "writer-both",
			mode:  WRITER,
			want:  BOTH,
			level: LOG,
			isErr: true,
		},
		{
			name:  "different mode",
			mode:  NONE,
			want:  STD,
			level: LOG,
			isErr: true,
		},
		{
			name:  "different mode",
			mode:  WRITER,
			want:  NONE,
			level: 123,
			isErr: false,
		},
	}
	g := New()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := g.SetLevelMode(tt.level, tt.mode).GetCurrentMode(tt.level); !reflect.DeepEqual(got, tt.want) && !tt.isErr {
				t.Errorf("Glg.SetMode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGlg_GetCurrentMode(t *testing.T) {
	tests := []struct {
		name  string
		mode  MODE
		want  MODE
		level LEVEL
	}{
		{
			name:  "std",
			mode:  STD,
			want:  STD,
			level: LOG,
		},
		{
			name:  "writer",
			mode:  WRITER,
			want:  WRITER,
			level: LOG,
		},
		{
			name:  "both",
			mode:  BOTH,
			want:  BOTH,
			level: LOG,
		},
		{
			name:  "none",
			mode:  NONE,
			want:  NONE,
			level: LOG,
		},
		{
			name:  "different mode",
			mode:  WRITER,
			want:  NONE,
			level: 123,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New().SetMode(tt.mode).GetCurrentMode(tt.level); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Glg.GetCurrentMode(LOG) = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGlg_InitWriter(t *testing.T) {

	t.Run("InitWriter Check", func(t *testing.T) {
		ins1 := New()
		ins2 := ins1.InitWriter()
		if ins1.GetCurrentMode(LOG) != ins2.GetCurrentMode(LOG) {
			t.Errorf("glg mode = %v, want %v", ins1.GetCurrentMode(LOG), ins2.GetCurrentMode(LOG))
		}

		if ins2.GetCurrentMode(LOG) != STD {
			t.Errorf("Expect %v, want %v", ins2.GetCurrentMode(LOG), STD)
		}

		ins1.logger.Range(func(key, val interface{}) bool {

			lev1, ok := val.(*logger)
			if !ok {
				t.Errorf("invalid glg instance 1 type:\t%v", val)
			}
			i2, ok := ins2.logger.Load(key.(LEVEL))
			if !ok {
				t.Error("glg instance 2 not found")
			}
			lev2, ok := i2.(*logger)
			if !ok {
				t.Errorf("invalid glg instance 2 type:\t%v", val)
			}
			if !reflect.DeepEqual(lev1, lev2) {
				t.Errorf("Expect %v, want %v", lev2, lev1)
				return false
			}
			return true
		})
	})

}

func TestGlg_SetWriter(t *testing.T) {

	tests := []struct {
		name string
		want io.Writer
		msg  string
	}{
		{
			name: "Set Custom writer",
			want: new(bytes.Buffer),
			msg:  "test",
		},
		{
			name: "Set nil writer",
			want: nil,
			msg:  "nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := New().SetMode(WRITER).SetWriter(tt.want)
			g.Info(tt.msg)
			if tt.want != nil {
				got := tt.want.(*bytes.Buffer).String()
				t.Log(got)
				if !strings.Contains(got, tt.msg) {
					t.Errorf("Glg.SetWriter() = %v, want %v", got, tt.msg)
				}
			} else {
				l, ok := g.logger.Load(INFO)
				if !ok {
					t.Error("glg instance not found")
				}
				ins, ok := l.(*logger)
				if !ok {
					t.Errorf("invalid glg instance type:\t%v", l)
				}
				if ins.writer != nil {
					t.Errorf("Glg.SetWriter() = %v, want %v", ins.writer, tt.want)
				}
			}
		})
	}
}

func TestGlg_AddWriter(t *testing.T) {
	tests := []struct {
		name string
		want io.Writer
		msg  string
	}{
		{
			name: "Add Custom writer",
			want: new(bytes.Buffer),
			msg:  "test",
		},
		{
			name: "Add nil writer",
			want: nil,
			msg:  "nil",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var writer io.Writer = new(bytes.Buffer)
			g := New().SetMode(WRITER).AddWriter(tt.want).AddWriter(writer)
			g.Info(tt.msg)
			if tt.want != nil {
				got := tt.want.(*bytes.Buffer).String()
				want := writer.(*bytes.Buffer).String()
				if !reflect.DeepEqual(got, want) {
					t.Errorf("Glg.AddWriter() = %vwant %v", got, want)
				}
			} else {
				l, ok := g.logger.Load(INFO)
				if !ok {
					t.Error("glg instance not found")
				}
				ins, ok := l.(*logger)
				if !ok {
					t.Errorf("invalid glg instance type:\t%v", l)
				}
				if ins.writer == nil {
					t.Errorf("Glg.AddWriter() = %v, want %v", ins.writer, tt.want)
				}
			}
		})
	}
}

func TestGlg_SetLevelColor(t *testing.T) {
	tests := []struct {
		name  string
		level LEVEL
		color func(string) string
		txt   string
		want  string
	}{
		{
			name:  "Set Level Color INFO=Green",
			level: INFO,
			color: Green,
			txt:   "green",
			want:  Green("green"),
		},
		{
			name:  "Set Level Color DEBG=Purple",
			level: DEBG,
			color: Purple,
			txt:   "purple",
			want:  Purple("purple"),
		},
		{
			name:  "Set Level Color WARN=Orange",
			level: WARN,
			color: Orange,
			txt:   "orange",
			want:  Orange("orange"),
		},
		{
			name:  "Set Level Color ERR=Red",
			level: ERR,
			color: Red,
			txt:   "red",
			want:  Red("red"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := New()
			g.SetLevelColor(tt.level, tt.color)
			l, ok := g.logger.Load(tt.level)
			if !ok {
				t.Error("glg instance not found")
			}
			ins, ok := l.(*logger)
			if !ok {
				t.Errorf("invalid glg instance type:\t%v", l)
			}
			got := ins.color(tt.txt)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Glg.SetLevelColor() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGlg_SetLevelWriter(t *testing.T) {
	tests := []struct {
		name   string
		writer io.Writer
		level  LEVEL
	}{
		{
			name:   "Info level",
			writer: new(bytes.Buffer),
			level:  INFO,
		},
		{
			name:   "Error level",
			writer: new(bytes.Buffer),
			level:  ERR,
		},
		{
			name:   "Set INFO level nil writer",
			writer: nil,
			level:  INFO,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := New()
			g.SetLevelWriter(tt.level, tt.writer)
			l, ok := g.logger.Load(tt.level)
			if !ok {
				t.Error("glg instance not found")
			}
			ins, ok := l.(*logger)
			if !ok {
				t.Errorf("invalid glg instance type:\t%v", l)
			}
			if !reflect.DeepEqual(ins.writer, tt.writer) {
				t.Errorf("Glg.SetLevelWriter() = %v, want %v", ins.writer, tt.writer)
			}
		})
	}
}

func TestGlg_AddLevelWriter(t *testing.T) {

	tests := []struct {
		glg    *Glg
		name   string
		writer io.Writer
		level  LEVEL
		multi  bool
	}{
		{
			glg:    New(),
			name:   "Info level",
			writer: new(bytes.Buffer),
			level:  INFO,
			multi:  false,
		},
		{
			glg:    New(),
			name:   "Error level",
			writer: new(bytes.Buffer),
			level:  ERR,
			multi:  false,
		},
		{
			glg:    New(),
			name:   "Append DEBG level",
			writer: new(bytes.Buffer),
			level:  DEBG,
			multi:  false,
		},
		{
			glg:    New(),
			name:   "Add INFO level nil writer",
			writer: nil,
			level:  INFO,
			multi:  false,
		},
		{
			glg:    Get().AddStdLevel("glg is fast", BOTH, false),
			name:   "Add Custom",
			writer: new(bytes.Buffer),
			level:  TagStringToLevel("glg is fast"),
			multi:  false,
		},
		{
			glg: Get().AddStdLevel("glg", BOTH, false).
				AddLevelWriter(TagStringToLevel("glg"),
					new(bytes.Buffer)),
			name:   "Add Custom",
			writer: new(bytes.Buffer),
			level:  TagStringToLevel("glg"),
			multi:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := tt.glg
			g.AddLevelWriter(tt.level, tt.writer)
			l, ok := g.logger.Load(tt.level)
			if !ok {
				t.Error("glg instance not found")
			}
			ins, ok := l.(*logger)
			if !ok {
				t.Errorf("invalid glg instance type:\t%v", l)
			}
			if !tt.multi && tt.writer != nil && !reflect.DeepEqual(ins.writer, tt.writer) {
				t.Errorf("Glg.AddLevelWriter() = %v, want %v", ins.writer, tt.writer)
			}
		})
	}
}

func TestGlg_AddStdLevel(t *testing.T) {
	tests := []struct {
		name  string
		level string
		want  io.Writer
	}{
		{
			name:  "custom std",
			level: "STD2",
			want:  os.Stdout,
		},
		{
			name:  "custom xxxx",
			level: "XXXX",
			want:  os.Stdout,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := New().AddStdLevel(tt.level, STD, false)
			l, ok := g.logger.Load(g.TagStringToLevel(tt.level))
			if !ok {
				t.Error("glg instance not found")
			}
			ins, ok := l.(*logger)
			if !ok || ins == nil {
				t.Errorf("invalid glg instance type:\t%v", l)
			}
		})
	}
}

func TestGlg_AddErrLevel(t *testing.T) {
	tests := []struct {
		name  string
		level string
		want  io.Writer
	}{
		{
			name:  "custom err",
			level: "ERR2",
			want:  os.Stderr,
		},
		{
			name:  "custom xxxx",
			level: "XXXX",
			want:  os.Stderr,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := New().AddErrLevel(tt.level, STD, false)
			l, ok := g.logger.Load(g.TagStringToLevel(tt.level))
			if !ok {
				t.Error("glg instance not found")
			}
			ins, ok := l.(*logger)
			if !ok || ins == nil {
				t.Errorf("invalid glg instance type:\t%v", l)
			}
		})
	}
}

func TestSetPrefix(t *testing.T) {
	tests := []struct {
		prefix string
		name   string
		want   string
	}{
		{
			name:   "Prefix GLG",
			prefix: "GLG",
			want:   "GLG",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetPrefix(tt.prefix)
			buf := new(bytes.Buffer)
			Get().SetWriter(buf)
			Get().SetMode(WRITER)
			Println("sample")
			if !strings.Contains(buf.String(), tt.want) {
				t.Errorf("SetPrefix = got %v want %v", buf.String(), tt.want)
			}
		})
	}
}

func TestGlg_SetPrefix(t *testing.T) {
	tests := []struct {
		prefix string
		name   string
		glg    *Glg
		want   string
	}{
		{
			name:   "Prefix GLG",
			glg:    New(),
			prefix: "GLG",
			want:   "GLG",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.glg.SetPrefix(tt.prefix)
			buf := new(bytes.Buffer)
			tt.glg.SetWriter(buf)
			tt.glg.SetMode(WRITER)
			tt.glg.Println("sample")
			if !strings.Contains(buf.String(), tt.want) {
				t.Errorf("SetPrefix = got %v want %v", buf.String(), tt.want)
			}
		})
	}

}

func TestGlg_EnableColor(t *testing.T) {
	tests := []struct {
		name string
		glg  *Glg
		want bool
	}{
		{
			name: "EnableColor",
			glg:  New().DisableColor(),
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l, ok := tt.glg.EnableColor().logger.Load(LOG)
			if !ok {
				t.Error("glg instance not found")
			}
			got := l.(*logger).isColor
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Glg.EnableColor() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGlg_DisableColor(t *testing.T) {
	tests := []struct {
		name string
		glg  *Glg
		want bool
	}{
		{
			name: "DisableColor",
			glg:  New().EnableColor(),
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l, ok := tt.glg.DisableColor().logger.Load(LOG)
			if !ok {
				t.Error("glg instance not found")
			}
			got := l.(*logger).isColor
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Glg.DisableColor() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGlg_EnableLevelColor(t *testing.T) {
	tests := []struct {
		name  string
		glg   *Glg
		want  bool
		level LEVEL
	}{
		{
			name:  "EnableColor",
			glg:   New().DisableColor(),
			want:  true,
			level: INFO,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l, ok := tt.glg.EnableLevelColor(tt.level).logger.Load(tt.level)
			if !ok {
				t.Error("glg instance not found")
			}
			got := l.(*logger).isColor
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Glg.DisableColor() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGlg_DisableLevelColor(t *testing.T) {
	tests := []struct {
		name  string
		glg   *Glg
		want  bool
		level LEVEL
	}{
		{
			name:  "DisableColor",
			glg:   New().EnableColor(),
			want:  false,
			level: WARN,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l, ok := tt.glg.DisableLevelColor(tt.level).logger.Load(tt.level)
			if !ok {
				t.Error("glg instance not found")
			}
			got := l.(*logger).isColor
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Glg.DisableColor() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTagStringToLevel(t *testing.T) {
	tests := []struct {
		name      string
		g         *Glg
		tag       string
		want      LEVEL
		createFlg bool
	}{
		{
			name:      "customTag",
			g:         Get().Reset(),
			tag:       "customTag",
			want:      TagStringToLevel("customTag"),
			createFlg: true,
		},
		{
			name:      "customTag No create",
			g:         Get(),
			tag:       "customTagFail",
			want:      255,
			createFlg: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.createFlg {
				tt.g.AddStdLevel(tt.tag, STD, false)
			}
			got := TagStringToLevel(tt.tag)
			if got != tt.want {
				t.Errorf("Glg.TagStringToLevel = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGlg_TagStringToLevel(t *testing.T) {
	tests := []struct {
		name      string
		g         *Glg
		tag       string
		want      LEVEL
		createFlg bool
	}{
		{
			name:      "customTag",
			g:         New(),
			tag:       "customTag",
			want:      TagStringToLevel("customTag"),
			createFlg: true,
		},
		{
			name:      "customTag No create",
			g:         New(),
			tag:       "customTagFail",
			want:      255,
			createFlg: false,
		}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.createFlg {
				tt.g.AddStdLevel(tt.tag, STD, false)
			}
			got := glg.TagStringToLevel(tt.tag)
			if got != tt.want {
				t.Errorf("Glg.TagStringToLevel = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFileWriter(t *testing.T) {

	tests := []struct {
		name  string
		path  string
		want  string
		isErr bool
	}{

		{
			name:  "sample file log",
			path:  "./sample.log",
			want:  "./sample.log",
			isErr: false,
		},
		{
			name:  "error file log",
			path:  "./error.log",
			want:  "./error.log",
			isErr: false,
		},
		{
			name:  "empty",
			path:  "",
			want:  "",
			isErr: false,
		},
		{
			name:  "root file log",
			path:  "/root.log",
			want:  "/root.log",
			isErr: false,
		},
		{
			name:  "root file log",
			path:  "/usr/tmp/root.log",
			want:  "/usr/tmp/root.log",
			isErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := FileWriter(tt.path, 0755)
			if f != nil {
				got := f.Name()
				if !tt.isErr && !reflect.DeepEqual(got, tt.want) {
					t.Errorf("FileWriter() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestGlg_HTTPLogger(t *testing.T) {

	type args struct {
		name string
		uri  string
	}
	tests := []struct {
		name string
		args args
		mode MODE
	}{
		{
			name: "http logger simple",
			args: args{
				name: "simple",
				uri:  "/",
			},
			mode: WRITER,
		},
		{
			name: "http logger err",
			args: args{
				name: "err",
				uri:  "err",
			},
			mode: WRITER,
		},
		{
			name: "none logger simple",
			args: args{
				name: "none",
				uri:  "/",
			},
			mode: NONE,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := new(bytes.Buffer)

			req, err := http.NewRequest(http.MethodGet, tt.args.uri, nil)
			if err != nil {
				t.Fatal(err)
			}
			rr := httptest.NewRecorder()

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
			want := fmt.Sprintf("Method: %s\tURI: %s\tName: %s",
				req.Method, req.RequestURI, tt.args.name)

			g := New().SetMode(tt.mode).SetWriter(w)

			g.HTTPLogger(tt.args.name, handler).ServeHTTP(rr, req)

			if !strings.Contains(w.String(), want) && tt.mode != NONE {
				t.Errorf("Glg.HTTPLogger() = %v, want %v", w.String(), want)
			}
		})
	}
}

func TestGlg_HTTPLoggerFunc(t *testing.T) {
	type args struct {
		name string
		uri  string
	}
	tests := []struct {
		name string
		args args
		mode MODE
	}{
		{
			name: "http logger simple",
			args: args{
				name: "simple",
				uri:  "/",
			},
			mode: WRITER,
		},
		{
			name: "http logger err",
			args: args{
				name: "err",
				uri:  "err",
			},
			mode: WRITER,
		},
		{
			name: "none logger simple",
			args: args{
				name: "none",
				uri:  "/",
			},
			mode: NONE,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := new(bytes.Buffer)

			req, err := http.NewRequest(http.MethodGet, tt.args.uri, nil)
			if err != nil {
				t.Fatal(err)
			}
			rr := httptest.NewRecorder()

			want := fmt.Sprintf("Method: %s\tURI: %s\tName: %s",
				req.Method, req.RequestURI, tt.args.name)

			g := New().SetMode(tt.mode).SetWriter(w)

			g.HTTPLoggerFunc(tt.args.name, func(w http.ResponseWriter, r *http.Request) {}).ServeHTTP(rr, req)

			if !strings.Contains(w.String(), want) && tt.mode != NONE {
				t.Errorf("Glg.HTTPLogger() = %v, want %v", w.String(), want)
			}
		})
	}
}

func TestHTTPLogger(t *testing.T) {
	type args struct {
		name string
		uri  string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "http logger simple",
			args: args{
				name: "simple",
				uri:  "/",
			},
		},
		{
			name: "http logger err",
			args: args{
				name: "err",
				uri:  "err",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := new(bytes.Buffer)

			req, err := http.NewRequest(http.MethodGet, tt.args.uri, nil)
			if err != nil {
				t.Fatal(err)
			}
			rr := httptest.NewRecorder()

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
			want := fmt.Sprintf("Method: %s\tURI: %s\tName: %s",
				req.Method, req.RequestURI, tt.args.name)

			Get().SetMode(WRITER).SetWriter(w)

			HTTPLogger(tt.args.name, handler).ServeHTTP(rr, req)

			if !strings.Contains(w.String(), want) {
				t.Errorf("HTTPLogger() = %v, want %v", w.String(), want)
			}
		})
	}
}

func TestHTTPLoggerFunc(t *testing.T) {
	type args struct {
		name string
		uri  string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "http logger simple",
			args: args{
				name: "simple",
				uri:  "/",
			},
		},
		{
			name: "http logger err",
			args: args{
				name: "err",
				uri:  "err",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := new(bytes.Buffer)

			req, err := http.NewRequest(http.MethodGet, tt.args.uri, nil)
			if err != nil {
				t.Fatal(err)
			}
			rr := httptest.NewRecorder()

			want := fmt.Sprintf("Method: %s\tURI: %s\tName: %s",
				req.Method, req.RequestURI, tt.args.name)

			Get().SetMode(WRITER).SetWriter(w)

			HTTPLoggerFunc(tt.args.name, func(w http.ResponseWriter, r *http.Request) {}).ServeHTTP(rr, req)

			if !strings.Contains(w.String(), want) {
				t.Errorf("HTTPLoggerFunc() = %v, want %v", w.String(), want)
			}
		})
	}
}

func TestColorless(t *testing.T) {
	tests := []struct {
		name string
		txt  string
		want string
	}{
		{
			name: "colorless",
			txt:  "message",
			want: "message",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Colorless(tt.txt); got != tt.want {
				t.Errorf("Colorless() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRed(t *testing.T) {
	tests := []struct {
		name string
		txt  string
		want string
	}{
		{
			name: "red",
			txt:  "message",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Red(tt.txt); !strings.HasPrefix(got, "\033[31m") || !strings.HasSuffix(got, "\033[39m") {
				t.Errorf("Red() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGreen(t *testing.T) {
	tests := []struct {
		name string
		txt  string
		want string
	}{
		{
			name: "green",
			txt:  "message",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Green(tt.txt); !strings.HasPrefix(got, "\033[32m") || !strings.HasSuffix(got, "\033[39m") {
				t.Errorf("Green() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOrange(t *testing.T) {
	tests := []struct {
		name string
		txt  string
		want string
	}{
		{
			name: "orange",
			txt:  "message",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Orange(tt.txt); !strings.HasPrefix(got, "\033[33m") || !strings.HasSuffix(got, "\033[39m") {
				t.Errorf("Orange() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPurple(t *testing.T) {
	tests := []struct {
		name string
		txt  string
		want string
	}{
		{
			name: "purple",
			txt:  "message",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Purple(tt.txt); !strings.HasPrefix(got, "\033[34m") || !strings.HasSuffix(got, "\033[39m") {
				t.Errorf("Purple() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCyan(t *testing.T) {
	tests := []struct {
		name string
		txt  string
		want string
	}{
		{
			name: "cyan",
			txt:  "message",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Cyan(tt.txt); !strings.HasPrefix(got, "\033[36m") || !strings.HasSuffix(got, "\033[39m") {
				t.Errorf("Cyan() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestYellow(t *testing.T) {
	tests := []struct {
		name string
		txt  string
		want string
	}{
		{
			name: "yellow",
			txt:  "message",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Yellow(tt.txt); !strings.HasPrefix(got, "\033[93m") || !strings.HasSuffix(got, "\033[39m") {
				t.Errorf("Yellow() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBrown(t *testing.T) {
	tests := []struct {
		name string
		txt  string
		want string
	}{
		{
			name: "brown",
			txt:  "message",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Brown(tt.txt); !strings.HasPrefix(got, "\033[96m") || !strings.HasSuffix(got, "\033[39m") {
				t.Errorf("Brown() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGray(t *testing.T) {
	tests := []struct {
		name string
		txt  string
		want string
	}{
		{
			name: "gray",
			txt:  "message",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Gray(tt.txt); !strings.HasPrefix(got, "\033[90m") || !strings.HasSuffix(got, "\033[39m") {
				t.Errorf("Gray() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlack(t *testing.T) {
	tests := []struct {
		name string
		txt  string
		want string
	}{
		{
			name: "black",
			txt:  "message",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Black(tt.txt); !strings.HasPrefix(got, "\033[30m") || !strings.HasSuffix(got, "\033[39m") {
				t.Errorf("Black() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWhite(t *testing.T) {
	tests := []struct {
		name string
		txt  string
		want string
	}{
		{
			name: "white",
			txt:  "message",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := White(tt.txt); !strings.HasPrefix(got, "\033[97m") || !strings.HasSuffix(got, "\033[39m") {
				t.Errorf("White() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGlg_out(t *testing.T) {
	tests := []struct {
		glg    *Glg
		name   string
		level  LEVEL
		format string
		val    []interface{}
	}{
		{
			glg:    New().SetMode(WRITER),
			name:   "sample info",
			level:  INFO,
			format: "%d%s%f",
			val: []interface{}{
				2,
				"aaa",
				3.6,
			},
		},
		{
			glg:    New().SetMode(WRITER),
			name:   "sample log",
			level:  LOG,
			format: "%d%s%f",
			val: []interface{}{
				2,
				"aaa",
				3.6,
			},
		},
		{
			glg:    New().SetMode(NONE),
			name:   "no log",
			level:  LOG,
			format: "%d%s%f",
			val: []interface{}{
				2,
				"aaa",
				3.6,
			},
		},
		{
			glg:    New().SetMode(STD),
			name:   "no log",
			level:  LOG,
			format: "%d%s%f",
			val: []interface{}{
				2,
				"aaa",
				3.6,
			},
		},
		{
			glg:    New().SetMode(BOTH),
			name:   "no log",
			level:  LOG,
			format: "%d%s%f",
			val: []interface{}{
				2,
				"aaa",
				3.6,
			},
		},
		{
			glg:    New().SetMode(STD).DisableColor(),
			name:   "no log",
			level:  LOG,
			format: "%d%s%f",
			val: []interface{}{
				2,
				"aaa",
				3.6,
			},
		},
		{
			glg:    New().SetMode(BOTH).DisableColor(),
			name:   "no log",
			level:  LOG,
			format: "%d%s%f",
			val: []interface{}{
				2,
				"aaa",
				3.6,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			g := tt.glg.SetWriter(buf)
			g.out(tt.level, tt.format, tt.val...)
			want := fmt.Sprintf(tt.format, tt.val...)
			if !strings.Contains(buf.String(), want) && tt.glg.GetCurrentMode(LOG) != NONE && tt.glg.GetCurrentMode(LOG) != STD {
				t.Errorf("Glg.out() = got %v want %v", buf.String(), want)
			}
		})
	}
}

func TestGlg_Log(t *testing.T) {
	tests := []struct {
		name string
		val  []interface{}
	}{
		{
			name: "sample log",
			val: []interface{}{
				"sample log",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			g := New().SetMode(WRITER).SetWriter(buf)
			g.Log(tt.val...)
			want := fmt.Sprintf("%v", tt.val...)
			if !strings.Contains(buf.String(), want) {
				t.Errorf("Glg.Log() = got %v want %v", buf.String(), want)
			}
		})
	}
}

func TestGlg_Logf(t *testing.T) {
	tests := []struct {
		name   string
		format string
		val    []interface{}
	}{
		{
			name:   "sample log",
			format: "%d%s%f",
			val: []interface{}{
				2,
				"aaa",
				3.6,
			},
		},
		{
			name:   "sample log",
			format: "%d%s%f",
			val: []interface{}{
				2,
				"aaa",
				3.6,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			g := New().SetMode(WRITER).SetWriter(buf)
			g.Logf(tt.format, tt.val...)
			want := fmt.Sprintf(tt.format, tt.val...)
			if !strings.Contains(buf.String(), want) {
				t.Errorf("Glg.Log() = got %v want %v", buf.String(), want)
			}
		})
	}
}

func TestLog(t *testing.T) {
	tests := []struct {
		name string
		val  []interface{}
	}{
		{
			name: "sample log",
			val: []interface{}{
				"sample log",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			Get().SetMode(WRITER).SetWriter(buf)
			Log(tt.val...)
			want := fmt.Sprintf("%v", tt.val...)
			if !strings.Contains(buf.String(), want) {
				t.Errorf("Log() = got %v want %v", buf.String(), want)
			}
		})
	}
}

func TestLogf(t *testing.T) {
	tests := []struct {
		name   string
		format string
		val    []interface{}
	}{
		{
			name:   "sample info",
			format: "%d%s%f",
			val: []interface{}{
				2,
				"aaa",
				3.6,
			},
		},
		{
			name:   "sample log",
			format: "%d%s%f",
			val: []interface{}{
				2,
				"aaa",
				3.6,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			Get().SetMode(WRITER).SetWriter(buf)
			Logf(tt.format, tt.val...)
			want := fmt.Sprintf(tt.format, tt.val...)
			if !strings.Contains(buf.String(), want) {
				t.Errorf("Logf() = got %v want %v", buf.String(), want)
			}
		})
	}
}

func TestGlg_Info(t *testing.T) {
	tests := []struct {
		name string
		val  []interface{}
	}{
		{
			name: "sample info",
			val: []interface{}{
				"sample info",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			g := New().SetMode(WRITER).SetWriter(buf)
			g.Info(tt.val...)
			want := fmt.Sprintf("%v", tt.val...)
			if !strings.Contains(buf.String(), want) {
				t.Errorf("Glg.Info() = got %v want %v", buf.String(), want)
			}
		})
	}
}

func TestGlg_Infof(t *testing.T) {
	tests := []struct {
		name   string
		format string
		val    []interface{}
	}{
		{
			name:   "sample info",
			format: "%d%s%f",
			val: []interface{}{
				2,
				"aaa",
				3.6,
			},
		},
		{
			name:   "sample info",
			format: "%d%s%f",
			val: []interface{}{
				2,
				"aaa",
				3.6,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			g := New().SetMode(WRITER).SetWriter(buf)
			g.Infof(tt.format, tt.val...)
			want := fmt.Sprintf(tt.format, tt.val...)
			if !strings.Contains(buf.String(), want) {
				t.Errorf("Glg.Infof() = got %v want %v", buf.String(), want)
			}
		})
	}
}

func TestInfo(t *testing.T) {
	tests := []struct {
		name string
		val  []interface{}
	}{
		{
			name: "sample info",
			val: []interface{}{
				"sample info",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			Get().SetMode(WRITER).SetWriter(buf)
			Info(tt.val...)
			want := fmt.Sprintf("%v", tt.val...)
			if !strings.Contains(buf.String(), want) {
				t.Errorf("Info() = got %v want %v", buf.String(), want)
			}
		})
	}
}

func TestInfof(t *testing.T) {
	tests := []struct {
		name   string
		format string
		val    []interface{}
	}{
		{
			name:   "sample info",
			format: "%d%s%f",
			val: []interface{}{
				2,
				"aaa",
				3.6,
			},
		},
		{
			name:   "sample info",
			format: "%d%s%f",
			val: []interface{}{
				2,
				"aaa",
				3.6,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			Get().SetMode(WRITER).SetWriter(buf)
			Infof(tt.format, tt.val...)
			want := fmt.Sprintf(tt.format, tt.val...)
			if !strings.Contains(buf.String(), want) {
				t.Errorf("Infof() = got %v want %v", buf.String(), want)
			}
		})
	}
}

func TestGlg_Success(t *testing.T) {
	tests := []struct {
		name string
		val  []interface{}
	}{
		{
			name: "sample success",
			val: []interface{}{
				"sample success",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			g := New().SetMode(WRITER).SetWriter(buf)
			g.Success(tt.val...)
			want := fmt.Sprintf("%v", tt.val...)
			if !strings.Contains(buf.String(), want) {
				t.Errorf("Glg.Success() = got %v want %v", buf.String(), want)
			}
		})
	}
}

func TestGlg_Successf(t *testing.T) {
	tests := []struct {
		name   string
		format string
		val    []interface{}
	}{
		{
			name:   "sample success",
			format: "%d%s%f",
			val: []interface{}{
				2,
				"aaa",
				3.6,
			},
		},
		{
			name:   "sample success",
			format: "%d%s%f",
			val: []interface{}{
				2,
				"aaa",
				3.6,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			g := New().SetMode(WRITER).SetWriter(buf)
			g.Successf(tt.format, tt.val...)
			want := fmt.Sprintf(tt.format, tt.val...)
			if !strings.Contains(buf.String(), want) {
				t.Errorf("Glg.Successf() = got %v want %v", buf.String(), want)
			}
		})
	}
}

func TestSuccess(t *testing.T) {
	tests := []struct {
		name string
		val  []interface{}
	}{
		{
			name: "sample success",
			val: []interface{}{
				"sample success",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			Get().SetMode(WRITER).SetWriter(buf)
			Success(tt.val...)
			want := fmt.Sprintf("%v", tt.val...)
			if !strings.Contains(buf.String(), want) {
				t.Errorf("Success() = got %v want %v", buf.String(), want)
			}
		})
	}

}

func TestSuccessf(t *testing.T) {
	tests := []struct {
		name   string
		format string
		val    []interface{}
	}{
		{
			name:   "sample success",
			format: "%d%s%f",
			val: []interface{}{
				2,
				"aaa",
				3.6,
			},
		},
		{
			name:   "sample success",
			format: "%d%s%f",
			val: []interface{}{
				2,
				"aaa",
				3.6,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			Get().SetMode(WRITER).SetWriter(buf)
			Successf(tt.format, tt.val...)
			want := fmt.Sprintf(tt.format, tt.val...)
			if !strings.Contains(buf.String(), want) {
				t.Errorf("Successf() = got %v want %v", buf.String(), want)
			}
		})
	}
}

func TestGlg_Debug(t *testing.T) {
	tests := []struct {
		name string
		val  []interface{}
	}{
		{
			name: "sample debug",
			val: []interface{}{
				"sample debug",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			g := New().SetMode(WRITER).SetWriter(buf)
			g.Debug(tt.val...)
			want := fmt.Sprintf("%v", tt.val...)
			if !strings.Contains(buf.String(), want) {
				t.Errorf("Glg.Debug() = got %v want %v", buf.String(), want)
			}
		})
	}
}

func TestGlg_Debugf(t *testing.T) {
	tests := []struct {
		name   string
		format string
		val    []interface{}
	}{
		{
			name:   "sample debug",
			format: "%d%s%f",
			val: []interface{}{
				2,
				"aaa",
				3.6,
			},
		},
		{
			name:   "sample debug",
			format: "%d%s%f",
			val: []interface{}{
				2,
				"aaa",
				3.6,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			g := New().SetMode(WRITER).SetWriter(buf)
			g.Debugf(tt.format, tt.val...)
			want := fmt.Sprintf(tt.format, tt.val...)
			if !strings.Contains(buf.String(), want) {
				t.Errorf("Glg.Debugf() = got %v want %v", buf.String(), want)
			}
		})
	}
}

func TestDebug(t *testing.T) {
	tests := []struct {
		name string
		val  []interface{}
	}{
		{
			name: "sample debug",
			val: []interface{}{
				"sample debug",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			Get().SetMode(WRITER).SetWriter(buf)
			Debug(tt.val...)
			want := fmt.Sprintf("%v", tt.val...)
			if !strings.Contains(buf.String(), want) {
				t.Errorf("Debug() = got %v want %v", buf.String(), want)
			}
		})
	}
}

func TestDebugf(t *testing.T) {
	tests := []struct {
		name   string
		format string
		val    []interface{}
	}{
		{
			name:   "sample debug",
			format: "%d%s%f",
			val: []interface{}{
				2,
				"aaa",
				3.6,
			},
		},
		{
			name:   "sample debug",
			format: "%d%s%f",
			val: []interface{}{
				2,
				"aaa",
				3.6,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			Get().SetMode(WRITER).SetWriter(buf)
			Debugf(tt.format, tt.val...)
			want := fmt.Sprintf(tt.format, tt.val...)
			if !strings.Contains(buf.String(), want) {
				t.Errorf("Debugf() = got %v want %v", buf.String(), want)
			}
		})
	}
}

func TestGlg_Warn(t *testing.T) {
	tests := []struct {
		name string
		val  []interface{}
	}{
		{
			name: "sample warn",
			val: []interface{}{
				"sample warn",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			g := New().SetMode(WRITER).SetWriter(buf)
			g.Warn(tt.val...)
			want := fmt.Sprintf("%v", tt.val...)
			if !strings.Contains(buf.String(), want) {
				t.Errorf("Glg.Warn() = got %v want %v", buf.String(), want)
			}
		})
	}
}

func TestGlg_Warnf(t *testing.T) {
	tests := []struct {
		name   string
		format string
		val    []interface{}
	}{
		{
			name:   "sample warnf",
			format: "%d%s%f",
			val: []interface{}{
				2,
				"aaa",
				3.6,
			},
		},
		{
			name:   "sample warnf",
			format: "%d%s%f",
			val: []interface{}{
				2,
				"aaa",
				3.6,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			g := New().SetMode(WRITER).SetWriter(buf)
			g.Warnf(tt.format, tt.val...)
			want := fmt.Sprintf(tt.format, tt.val...)
			if !strings.Contains(buf.String(), want) {
				t.Errorf("Glg.Warnf() = got %v want %v", buf.String(), want)
			}
		})
	}
}

func TestWarn(t *testing.T) {
	tests := []struct {
		name string
		val  []interface{}
	}{
		{
			name: "sample warn",
			val: []interface{}{
				"sample warn",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			Get().SetMode(WRITER).SetWriter(buf)
			Warn(tt.val...)
			want := fmt.Sprintf("%v", tt.val...)
			if !strings.Contains(buf.String(), want) {
				t.Errorf("Warn() = got %v want %v", buf.String(), want)
			}
		})
	}
}

func TestWarnf(t *testing.T) {
	tests := []struct {
		name   string
		format string
		val    []interface{}
	}{
		{
			name:   "sample warnf",
			format: "%d%s%f",
			val: []interface{}{
				2,
				"aaa",
				3.6,
			},
		},
		{
			name:   "sample warnf",
			format: "%d%s%f",
			val: []interface{}{
				2,
				"aaa",
				3.6,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			Get().SetMode(WRITER).SetWriter(buf)
			Warnf(tt.format, tt.val...)
			want := fmt.Sprintf(tt.format, tt.val...)
			if !strings.Contains(buf.String(), want) {
				t.Errorf("Warnf() = got %v want %v", buf.String(), want)
			}
		})
	}
}

func TestGlg_CustomLog(t *testing.T) {
	tests := []struct {
		name  string
		level string
		val   []interface{}
	}{
		{
			name:  "sample custom",
			level: "custom",
			val: []interface{}{
				"sample custom",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			g := New()
			g.SetMode(WRITER).AddStdLevel(tt.level, WRITER, false)
			g.SetWriter(buf)
			g.CustomLog(tt.level, tt.val...)
			want := fmt.Sprintf("%v", tt.val...)
			if !strings.Contains(buf.String(), want) {
				t.Errorf("Glg.CustomLog() = got %v want %v", buf.String(), want)
			}
		})
	}
}

func TestGlg_CustomLogf(t *testing.T) {
	tests := []struct {
		name   string
		format string
		level  string
		val    []interface{}
	}{
		{
			name:   "sample customf",
			level:  "custom",
			format: "%d%s%f",
			val: []interface{}{
				2,
				"aaa",
				3.6,
			},
		},
		{
			name:   "sample customf",
			level:  "custom",
			format: "%d%s%f",
			val: []interface{}{
				2,
				"aaa",
				3.6,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			g := New()
			g.SetMode(WRITER).AddStdLevel(tt.level, WRITER, false)
			g.SetWriter(buf)
			g.CustomLogf(tt.level, tt.format, tt.val...)
			want := fmt.Sprintf(tt.format, tt.val...)
			if !strings.Contains(buf.String(), want) {
				t.Errorf("Glg.CustomLogf() = got %v want %v", buf.String(), want)
			}
		})
	}
}

func TestCustomLog(t *testing.T) {
	tests := []struct {
		name  string
		level string
		val   []interface{}
	}{
		{
			name:  "sample custom",
			level: "custom",
			val: []interface{}{
				"sample custom",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			Get().SetMode(WRITER).AddStdLevel(tt.level, WRITER, false)
			Get().SetWriter(buf)
			CustomLog(tt.level, tt.val...)
			want := fmt.Sprintf("%v", tt.val...)
			if !strings.Contains(buf.String(), want) {
				t.Errorf("CustomLog() = got %v want %v", buf.String(), want)
			}
		})
	}
}

func TestCustomLogf(t *testing.T) {
	tests := []struct {
		name   string
		format string
		level  string
		val    []interface{}
	}{
		{
			name:   "sample customf",
			level:  "custom",
			format: "%d%s%f",
			val: []interface{}{
				2,
				"aaa",
				3.6,
			},
		},
		{
			name:   "sample customf",
			level:  "custom",
			format: "%d%s%f",
			val: []interface{}{
				2,
				"aaa",
				3.6,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			Get().SetMode(WRITER).AddStdLevel(tt.level, WRITER, false)
			Get().SetWriter(buf)
			CustomLogf(tt.level, tt.format, tt.val...)
			want := fmt.Sprintf(tt.format, tt.val...)
			if !strings.Contains(buf.String(), want) {
				t.Errorf("Glg.Warnf() = got %v want %v", buf.String(), want)
			}
		})
	}
}

func TestGlg_Print(t *testing.T) {
	tests := []struct {
		name string
		val  []interface{}
	}{
		{
			name: "sample print",
			val: []interface{}{
				"sample print",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			g := New().SetMode(WRITER).SetWriter(buf)
			g.Print(tt.val...)
			want := fmt.Sprintf("%v", tt.val...)
			if !strings.Contains(buf.String(), want) {
				t.Errorf("Glg.Print() = got %v want %v", buf.String(), want)
			}
		})
	}
}

func TestGlg_Println(t *testing.T) {
	tests := []struct {
		name string
		val  []interface{}
	}{
		{
			name: "sample println",
			val: []interface{}{
				"sample println",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			g := New().SetMode(WRITER).SetWriter(buf)
			g.Println(tt.val...)
			want := fmt.Sprintf("%v", tt.val...)
			if !strings.Contains(buf.String(), want) {
				t.Errorf("Glg.Println() = got %v want %v", buf.String(), want)
			}
		})
	}
}

func TestGlg_Printf(t *testing.T) {
	tests := []struct {
		name   string
		format string
		val    []interface{}
	}{
		{
			name:   "sample printf",
			format: "%d%s%f",
			val: []interface{}{
				2,
				"aaa",
				3.6,
			},
		},
		{
			name:   "sample printf",
			format: "%d%s%f",
			val: []interface{}{
				2,
				"aaa",
				3.6,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			g := New().SetMode(WRITER).SetWriter(buf)
			g.Printf(tt.format, tt.val...)
			want := fmt.Sprintf(tt.format, tt.val...)
			if !strings.Contains(buf.String(), want) {
				t.Errorf("Glg.Printf() = got %v want %v", buf.String(), want)
			}
		})
	}
}

func TestPrint(t *testing.T) {
	tests := []struct {
		name string
		val  []interface{}
	}{
		{
			name: "sample print",
			val: []interface{}{
				"sample print",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			Get().SetMode(WRITER).SetWriter(buf)
			Print(tt.val...)
			want := fmt.Sprintf("%v", tt.val...)
			if !strings.Contains(buf.String(), want) {
				t.Errorf("Print() = got %v want %v", buf.String(), want)
			}
		})
	}
}

func TestPrintln(t *testing.T) {
	tests := []struct {
		name string
		val  []interface{}
	}{
		{
			name: "sample println",
			val: []interface{}{
				"sample println",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			Get().SetMode(WRITER).SetWriter(buf)
			Println(tt.val...)
			want := fmt.Sprintf("%v", tt.val...)
			if !strings.Contains(buf.String(), want) {
				t.Errorf("Println() = got %v want %v", buf.String(), want)
			}
		})
	}
}

func TestPrintf(t *testing.T) {
	tests := []struct {
		name   string
		format string
		val    []interface{}
	}{
		{
			name:   "sample printf",
			format: "%d%s%f",
			val: []interface{}{
				2,
				"aaa",
				3.6,
			},
		},
		{
			name:   "sample printf",
			format: "%d%s%f",
			val: []interface{}{
				2,
				"aaa",
				3.6,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			Get().SetMode(WRITER).SetWriter(buf)
			Printf(tt.format, tt.val...)
			want := fmt.Sprintf(tt.format, tt.val...)
			if !strings.Contains(buf.String(), want) {
				t.Errorf("Printf() = got %v want %v", buf.String(), want)
			}
		})
	}
}

func TestGlg_Error(t *testing.T) {
	tests := []struct {
		name string
		val  []interface{}
	}{
		{
			name: "sample error",
			val: []interface{}{
				"sample error",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			g := New().SetMode(WRITER).SetWriter(buf)
			g.Error(tt.val...)
			want := fmt.Sprintf("%v", tt.val...)
			if !strings.Contains(buf.String(), want) {
				t.Errorf("Glg.Error() = got %v want %v", buf.String(), want)
			}
		})
	}
}

func TestGlg_Errorf(t *testing.T) {
	tests := []struct {
		name   string
		format string
		val    []interface{}
	}{
		{
			name:   "sample errorf",
			format: "%d%s%f",
			val: []interface{}{
				2,
				"aaa",
				3.6,
			},
		},
		{
			name:   "sample errorf",
			format: "%d%s%f",
			val: []interface{}{
				2,
				"aaa",
				3.6,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			g := New().SetMode(WRITER).SetWriter(buf)
			g.Errorf(tt.format, tt.val...)
			want := fmt.Sprintf(tt.format, tt.val...)
			if !strings.Contains(buf.String(), want) {
				t.Errorf("Glg.Errorf() = got %v want %v", buf.String(), want)
			}
		})
	}
}

func TestError(t *testing.T) {
	tests := []struct {
		name string
		val  []interface{}
	}{
		{
			name: "sample error",
			val: []interface{}{
				"sample error",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			Get().SetMode(WRITER).SetWriter(buf)
			Error(tt.val...)
			want := fmt.Sprintf("%v", tt.val...)
			if !strings.Contains(buf.String(), want) {
				t.Errorf("Glg.Error() = got %v want %v", buf.String(), want)
			}
		})
	}
}

func TestErrorf(t *testing.T) {
	tests := []struct {
		name   string
		format string
		val    []interface{}
	}{
		{
			name:   "sample errorf",
			format: "%d%s%f",
			val: []interface{}{
				2,
				"aaa",
				3.6,
			},
		},
		{
			name:   "sample errorf",
			format: "%d%s%f",
			val: []interface{}{
				2,
				"aaa",
				3.6,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			Get().SetMode(WRITER).SetWriter(buf)
			Errorf(tt.format, tt.val...)
			want := fmt.Sprintf(tt.format, tt.val...)
			if !strings.Contains(buf.String(), want) {
				t.Errorf("Errorf() = got %v want %v", buf.String(), want)
			}
		})
	}
}

func TestGlg_Fail(t *testing.T) {
	tests := []struct {
		name string
		val  []interface{}
	}{
		{
			name: "sample fail",
			val: []interface{}{
				"sample fail",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			g := New().SetMode(WRITER).SetWriter(buf)
			g.Fail(tt.val...)
			want := fmt.Sprintf("%v", tt.val...)
			if !strings.Contains(buf.String(), want) {
				t.Errorf("Glg.Fail() = got %v want %v", buf.String(), want)
			}
		})
	}
}

func TestGlg_Failf(t *testing.T) {
	tests := []struct {
		name   string
		format string
		val    []interface{}
	}{
		{
			name:   "sample failf",
			format: "%d%s%f",
			val: []interface{}{
				2,
				"aaa",
				3.6,
			},
		},
		{
			name:   "sample failf",
			format: "%d%s%f",
			val: []interface{}{
				2,
				"aaa",
				3.6,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			g := New().SetMode(WRITER).SetWriter(buf)
			g.Failf(tt.format, tt.val...)
			want := fmt.Sprintf(tt.format, tt.val...)
			if !strings.Contains(buf.String(), want) {
				t.Errorf("Glg.Failf() = got %v want %v", buf.String(), want)
			}
		})
	}
}

func TestFail(t *testing.T) {
	tests := []struct {
		name string
		val  []interface{}
	}{
		{
			name: "sample fail",
			val: []interface{}{
				"sample fail",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			Get().SetMode(WRITER).SetWriter(buf)
			Fail(tt.val...)
			want := fmt.Sprintf("%v", tt.val...)
			if !strings.Contains(buf.String(), want) {
				t.Errorf("Fail() = got %v want %v", buf.String(), want)
			}
		})
	}
}

func TestFailf(t *testing.T) {
	tests := []struct {
		name   string
		format string
		val    []interface{}
	}{
		{
			name:   "sample failf",
			format: "%d%s%f",
			val: []interface{}{
				2,
				"aaa",
				3.6,
			},
		},
		{
			name:   "sample failf",
			format: "%d%s%f",
			val: []interface{}{
				2,
				"aaa",
				3.6,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			Get().SetMode(WRITER).SetWriter(buf)
			Failf(tt.format, tt.val...)
			want := fmt.Sprintf(tt.format, tt.val...)
			if !strings.Contains(buf.String(), want) {
				t.Errorf("Glg.Failf() = got %v want %v", buf.String(), want)
			}
		})
	}
}

func TestGlg_Fatal(t *testing.T) {
	tests := []struct {
		name string
		val  []interface{}
	}{
		{
			name: "sample fatal",
			val: []interface{}{
				"aaa",
			},
		},
		{
			name: "sample fatal",
			val: []interface{}{
				"aaa",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			g := New().SetMode(WRITER).SetWriter(buf)
			testExit(0, func() {
				g.Fatal(tt.val...)
			})
			want := fmt.Sprintf("%v", tt.val...)
			if !strings.Contains(buf.String(), want) {
				t.Errorf("Glg.Fatal() = got %v want %v", buf.String(), want)
			}
		})
	}
}

func TestGlg_Fatalln(t *testing.T) {
	tests := []struct {
		name string
		val  []interface{}
	}{
		{
			name: "sample fatalln",
			val: []interface{}{
				"aaa",
			},
		},
		{
			name: "sample fatalln",
			val: []interface{}{
				"aaa",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			g := New().SetMode(WRITER).SetWriter(buf)
			testExit(0, func() {
				g.Fatalln(tt.val...)
			})
			want := fmt.Sprintf("%v", tt.val...)
			if !strings.Contains(buf.String(), want) {
				t.Errorf("Glg.Fatalln() = got %v want %v", buf.String(), want)
			}
		})
	}
}

func TestGlg_Fatalf(t *testing.T) {
	tests := []struct {
		name   string
		format string
		val    []interface{}
	}{
		{
			name:   "sample fatalf",
			format: "%d%s%f",
			val: []interface{}{
				2,
				"aaa",
				3.6,
			},
		},
		{
			name:   "sample fatalf",
			format: "%d%s%f",
			val: []interface{}{
				2,
				"aaa",
				3.6,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			g := New().SetMode(WRITER).SetWriter(buf)
			testExit(0, func() {
				g.Fatalf(tt.format, tt.val...)
			})
			want := fmt.Sprintf(tt.format, tt.val...)
			if !strings.Contains(buf.String(), want) {
				t.Errorf("Glg.Fatalf() = got %v want %v", buf.String(), want)
			}
		})
	}
}

func TestFatal(t *testing.T) {
	tests := []struct {
		name string
		val  []interface{}
	}{
		{
			name: "sample fatal",
			val: []interface{}{
				"aaa",
			},
		},
		{
			name: "sample fatal",
			val: []interface{}{
				"aaa",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			Get().SetMode(WRITER).SetWriter(buf)
			testExit(0, func() {
				Fatal(tt.val...)
			})
			want := fmt.Sprintf("%v", tt.val...)
			if !strings.Contains(buf.String(), want) {
				t.Errorf("Fatal() = got %v want %v", buf.String(), want)
			}
		})
	}
}

func TestFatalf(t *testing.T) {
	tests := []struct {
		name   string
		format string
		val    []interface{}
	}{
		{
			name:   "sample fatalf",
			format: "%d%s%f",
			val: []interface{}{
				2,
				"aaa",
				3.6,
			},
		},
		{
			name:   "sample fatalf",
			format: "%d%s%f",
			val: []interface{}{
				2,
				"aaa",
				3.6,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			Get().SetMode(WRITER).SetWriter(buf)
			testExit(0, func() {
				Fatalf(tt.format, tt.val...)
			})
			want := fmt.Sprintf(tt.format, tt.val...)
			if !strings.Contains(buf.String(), want) {
				t.Errorf("Fatalf() = got %v want %v", buf.String(), want)
			}
		})
	}
}

func TestFatalln(t *testing.T) {
	tests := []struct {
		name string
		val  []interface{}
	}{
		{
			name: "sample fatalln",
			val: []interface{}{
				"aaa",
			},
		},
		{
			name: "sample fatalln",
			val: []interface{}{
				"aaa",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			Get().SetMode(WRITER).SetWriter(buf)
			testExit(0, func() {
				Fatalln(tt.val...)
			})
			want := fmt.Sprintf("%v", tt.val...)
			if !strings.Contains(buf.String(), want) {
				t.Errorf("Fatalln() = got %v want %v", buf.String(), want)
			}
		})
	}
}

func TestReplaceExitFunc(t *testing.T) {
	tests := []struct {
		name string
		fn   func(i int)
	}{
		{
			name: "just pass",
			fn:   func(i int) {},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ReplaceExitFunc(tt.fn)
		})
	}
}

func TestGlg_Reset(t *testing.T) {
	tests := []struct {
		name string
		tag  string
		g    *Glg
		want LEVEL
	}{
		{
			name: "reset",
			tag:  "glg",
			g:    Get().Reset(),
			want: 255,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.g.AddStdLevel(tt.tag, NONE, false)
			tt.g.Reset()
			got := tt.g.TagStringToLevel(tt.tag)
			if tt.want == got {
				t.Errorf("Reset() = got %v want %v", got, tt.want)
			}
		})
	}
}
