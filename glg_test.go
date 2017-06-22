package glg

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"reflect"
	"strings"
	"sync"
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

func TestNew(t *testing.T) {
	t.Run("Comparing simple instances", func(t *testing.T) {
		ins1 := New()
		ins2 := New()
		if ins1.GetCurrentMode() != ins2.GetCurrentMode() {
			t.Errorf("glg mode = %v, want %v", ins1.GetCurrentMode(), ins2.GetCurrentMode())
		}

		for k, v := range ins1.writer {
			v2, ok := ins2.writer[k]
			if !ok {
				t.Error("glg writer not found")
			}

			if v2 != v {
				t.Errorf("Expect %v, want %v", v2, v)
			}
		}

		for k, v := range ins1.std {
			v2, ok := ins2.std[k]
			if !ok {
				t.Error("glg std writer not found")
			}

			if v2 != v {
				t.Errorf("Expect %v, want %v", v2, v)
			}
		}

		for k, v := range ins1.colors {
			v2, ok := ins2.colors[k]
			if !ok {
				t.Error("glg color func not found")
			}

			if v2("test") != v("test") {
				t.Errorf("Expect %v, want %v", v2("test"), v("test"))
			}
		}
	})
}

func TestGet(t *testing.T) {
	t.Run("Comparing singleton instances", func(t *testing.T) {
		ins1 := Get()
		ins2 := Get()

		if ins1 != ins2 {
			t.Errorf("Expect %v, want %v", ins2, ins1)
		}

		if ins1.GetCurrentMode() != ins2.GetCurrentMode() {
			t.Errorf("glg mode = %v, want %v", ins1.GetCurrentMode(), ins2.GetCurrentMode())
		}

		for k, v := range ins1.writer {
			v2, ok := ins2.writer[k]
			if !ok {
				t.Error("glg writer not found")
			}

			if v2 != v {
				t.Errorf("Expect %v, want %v", v2, v)
			}
		}

		for k, v := range ins1.std {
			v2, ok := ins2.std[k]
			if !ok {
				t.Error("glg std writer not found")
			}

			if v2 != v {
				t.Errorf("Expect %v, want %v", v2, v)
			}
		}

		for k, v := range ins1.colors {
			v2, ok := ins2.colors[k]
			if !ok {
				t.Error("glg color func not found")
			}

			if v2("test") != v("test") {
				t.Errorf("Expect %v, want %v", v2("test"), v("test"))
			}
		}
	})
}

func TestGlg_SetMode(t *testing.T) {
	tests := []struct {
		name  string
		mode  int
		want  int
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
			if got := g.SetMode(tt.mode).GetCurrentMode(); !reflect.DeepEqual(got, tt.want) && !tt.isErr {
				t.Errorf("Glg.SetMode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGlg_GetCurrentMode(t *testing.T) {
	tests := []struct {
		name string
		mode int
		want int
	}{
		{
			name: "std",
			mode: STD,
			want: STD,
		},
		{
			name: "writer",
			mode: WRITER,
			want: WRITER,
		},
		{
			name: "both",
			mode: BOTH,
			want: BOTH,
		},
		{
			name: "none",
			mode: NONE,
			want: NONE,
		},
	}
	g := New()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := g.SetMode(tt.mode).GetCurrentMode(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Glg.GetCurrentMode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGlg_InitWriter(t *testing.T) {

	t.Run("InitWriter Check", func(t *testing.T) {
		ins1 := New()
		ins2 := ins1.InitWriter()
		if ins1.GetCurrentMode() != ins2.GetCurrentMode() {
			t.Errorf("glg mode = %v, want %v", ins1.GetCurrentMode(), ins2.GetCurrentMode())
		}

		if ins2.GetCurrentMode() != STD {
			t.Errorf("Expect %v, want %v", ins2.GetCurrentMode(), STD)
		}

		for k, v := range ins1.writer {
			v2, ok := ins2.writer[k]
			if !ok {
				t.Error("glg writer not found")
			}

			if v2 != v {
				t.Errorf("Expect %v, want %v", v2, v)
			}
		}

		for k, v := range ins1.std {
			v2, ok := ins2.std[k]
			if !ok {
				t.Error("glg std writer not found")
			}

			if v2 != v {
				t.Errorf("Expect %v, want %v", v2, v)
			}
		}

		for k, v := range ins1.colors {
			v2, ok := ins2.colors[k]
			if !ok {
				t.Error("glg color func not found")
			}

			if v2("test") != v("test") {
				t.Errorf("Expect %v, want %v", v2("test"), v("test"))
			}
		}
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := New().SetMode(WRITER).SetWriter(tt.want)
			g.Info(tt.msg)
			got := tt.want.(*bytes.Buffer).String()
			t.Log(got)
			if !strings.Contains(got, tt.msg) {
				t.Errorf("Glg.SetWriter() = %v, want %v", got, tt.msg)
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var writer io.Writer = new(bytes.Buffer)
			g := New().SetMode(WRITER).AddWriter(tt.want).AddWriter(writer)
			g.Info(tt.msg)
			got := tt.want.(*bytes.Buffer).String()
			want := writer.(*bytes.Buffer).String()
			if !reflect.DeepEqual(got, want) {
				t.Errorf("Glg.AddWriter() = %vwant %v", got, want)
			}
		})
	}
}

func TestGlg_SetLevelColor(t *testing.T) {
	tests := []struct {
		name  string
		level string
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
			got := g.colors[tt.level](tt.txt)
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
		level  string
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := New()
			g.SetLevelWriter(tt.level, tt.writer)
			got, ok := g.writer[tt.level]
			if !ok || !reflect.DeepEqual(got, tt.writer) {
				t.Errorf("Glg.SetLevelWriter() = %v, want %v", got, tt.writer)
			}
		})
	}
}

func TestGlg_AddLevelWriter(t *testing.T) {
	tests := []struct {
		name   string
		writer io.Writer
		level  string
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := New()
			g.AddLevelWriter(tt.level, tt.writer)
			got, ok := g.writer[tt.level]
			if !ok || !reflect.DeepEqual(got, tt.writer) {
				t.Errorf("Glg.AddLevelWriter() = %v, want %v", got, tt.writer)
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
			g := New()
			g.AddStdLevel(tt.level)
			got, ok := g.std[tt.level]
			if !ok || !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Glg.AddStdLevel() = %v, want %v", got, tt.want)
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
			g := New()
			g.AddErrLevel(tt.level)
			got, ok := g.std[tt.level]
			if !ok || !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Glg.AddErrLevel() = %v, want %v", got, tt.want)
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
			got := tt.glg.EnableColor().isColor
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
			name: "EnableColor",
			glg:  New().EnableColor(),
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.glg.DisableColor().isColor
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Glg.DisableColor() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFileWriter(t *testing.T) {
	type args struct {
		path string
		perm os.FileMode
	}
	tests := []struct {
		name string
		args args
		want *os.File
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FileWriter(tt.args.path, tt.args.perm); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FileWriter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGlg_HTTPLogger(t *testing.T) {
	type fields struct {
		writer  map[string]io.Writer
		std     map[string]io.Writer
		colors  map[string]func(string) string
		mode    int
		isColor bool
		mu      *sync.Mutex
	}
	type args struct {
		name    string
		handler http.Handler
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   http.Handler
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Glg{
				writer:  tt.fields.writer,
				std:     tt.fields.std,
				colors:  tt.fields.colors,
				mode:    tt.fields.mode,
				isColor: tt.fields.isColor,
				mu:      tt.fields.mu,
			}
			if got := g.HTTPLogger(tt.args.name, tt.args.handler); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Glg.HTTPLogger() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGlg_HTTPLoggerFunc(t *testing.T) {
	type fields struct {
		writer  map[string]io.Writer
		std     map[string]io.Writer
		colors  map[string]func(string) string
		mode    int
		isColor bool
		mu      *sync.Mutex
	}
	type args struct {
		name string
		hf   http.HandlerFunc
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   http.Handler
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Glg{
				writer:  tt.fields.writer,
				std:     tt.fields.std,
				colors:  tt.fields.colors,
				mode:    tt.fields.mode,
				isColor: tt.fields.isColor,
				mu:      tt.fields.mu,
			}
			if got := g.HTTPLoggerFunc(tt.args.name, tt.args.hf); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Glg.HTTPLoggerFunc() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHTTPLogger(t *testing.T) {
	type args struct {
		name    string
		handler http.Handler
	}
	tests := []struct {
		name string
		args args
		want http.Handler
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HTTPLogger(tt.args.name, tt.args.handler); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("HTTPLogger() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHTTPLoggerFunc(t *testing.T) {
	type args struct {
		name string
		hf   http.HandlerFunc
	}
	tests := []struct {
		name string
		args args
		want http.Handler
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HTTPLoggerFunc(tt.args.name, tt.args.hf); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("HTTPLoggerFunc() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestColorless(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Colorless(tt.args.str); got != tt.want {
				t.Errorf("Colorless() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRed(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Red(tt.args.str); got != tt.want {
				t.Errorf("Red() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGreen(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Green(tt.args.str); got != tt.want {
				t.Errorf("Green() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOrange(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Orange(tt.args.str); got != tt.want {
				t.Errorf("Orange() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPurple(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Purple(tt.args.str); got != tt.want {
				t.Errorf("Purple() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCyan(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Cyan(tt.args.str); got != tt.want {
				t.Errorf("Cyan() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestYellow(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Yellow(tt.args.str); got != tt.want {
				t.Errorf("Yellow() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBrown(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Brown(tt.args.str); got != tt.want {
				t.Errorf("Brown() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGray(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Gray(tt.args.str); got != tt.want {
				t.Errorf("Gray() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlack(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Black(tt.args.str); got != tt.want {
				t.Errorf("Black() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWhite(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := White(tt.args.str); got != tt.want {
				t.Errorf("White() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGlg_out(t *testing.T) {
	type fields struct {
		writer  map[string]io.Writer
		std     map[string]io.Writer
		colors  map[string]func(string) string
		mode    int
		isColor bool
		mu      *sync.Mutex
	}
	type args struct {
		level  string
		format string
		val    []interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Glg{
				writer:  tt.fields.writer,
				std:     tt.fields.std,
				colors:  tt.fields.colors,
				mode:    tt.fields.mode,
				isColor: tt.fields.isColor,
				mu:      tt.fields.mu,
			}
			if err := g.out(tt.args.level, tt.args.format, tt.args.val...); (err != nil) != tt.wantErr {
				t.Errorf("Glg.out() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGlg_Log(t *testing.T) {
	type fields struct {
		writer  map[string]io.Writer
		std     map[string]io.Writer
		colors  map[string]func(string) string
		mode    int
		isColor bool
		mu      *sync.Mutex
	}
	type args struct {
		val []interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Glg{
				writer:  tt.fields.writer,
				std:     tt.fields.std,
				colors:  tt.fields.colors,
				mode:    tt.fields.mode,
				isColor: tt.fields.isColor,
				mu:      tt.fields.mu,
			}
			if err := g.Log(tt.args.val...); (err != nil) != tt.wantErr {
				t.Errorf("Glg.Log() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGlg_Logf(t *testing.T) {
	type fields struct {
		writer  map[string]io.Writer
		std     map[string]io.Writer
		colors  map[string]func(string) string
		mode    int
		isColor bool
		mu      *sync.Mutex
	}
	type args struct {
		format string
		val    []interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Glg{
				writer:  tt.fields.writer,
				std:     tt.fields.std,
				colors:  tt.fields.colors,
				mode:    tt.fields.mode,
				isColor: tt.fields.isColor,
				mu:      tt.fields.mu,
			}
			if err := g.Logf(tt.args.format, tt.args.val...); (err != nil) != tt.wantErr {
				t.Errorf("Glg.Logf() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLog(t *testing.T) {
	type args struct {
		val []interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Log(tt.args.val...); (err != nil) != tt.wantErr {
				t.Errorf("Log() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLogf(t *testing.T) {
	type args struct {
		format string
		val    []interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Logf(tt.args.format, tt.args.val...); (err != nil) != tt.wantErr {
				t.Errorf("Logf() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGlg_Info(t *testing.T) {
	type fields struct {
		writer  map[string]io.Writer
		std     map[string]io.Writer
		colors  map[string]func(string) string
		mode    int
		isColor bool
		mu      *sync.Mutex
	}
	type args struct {
		val []interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Glg{
				writer:  tt.fields.writer,
				std:     tt.fields.std,
				colors:  tt.fields.colors,
				mode:    tt.fields.mode,
				isColor: tt.fields.isColor,
				mu:      tt.fields.mu,
			}
			if err := g.Info(tt.args.val...); (err != nil) != tt.wantErr {
				t.Errorf("Glg.Info() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGlg_Infof(t *testing.T) {
	type fields struct {
		writer  map[string]io.Writer
		std     map[string]io.Writer
		colors  map[string]func(string) string
		mode    int
		isColor bool
		mu      *sync.Mutex
	}
	type args struct {
		format string
		val    []interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Glg{
				writer:  tt.fields.writer,
				std:     tt.fields.std,
				colors:  tt.fields.colors,
				mode:    tt.fields.mode,
				isColor: tt.fields.isColor,
				mu:      tt.fields.mu,
			}
			if err := g.Infof(tt.args.format, tt.args.val...); (err != nil) != tt.wantErr {
				t.Errorf("Glg.Infof() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestInfo(t *testing.T) {
	type args struct {
		val []interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Info(tt.args.val...); (err != nil) != tt.wantErr {
				t.Errorf("Info() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestInfof(t *testing.T) {
	type args struct {
		format string
		val    []interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Infof(tt.args.format, tt.args.val...); (err != nil) != tt.wantErr {
				t.Errorf("Infof() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGlg_Success(t *testing.T) {
	type fields struct {
		writer  map[string]io.Writer
		std     map[string]io.Writer
		colors  map[string]func(string) string
		mode    int
		isColor bool
		mu      *sync.Mutex
	}
	type args struct {
		val []interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Glg{
				writer:  tt.fields.writer,
				std:     tt.fields.std,
				colors:  tt.fields.colors,
				mode:    tt.fields.mode,
				isColor: tt.fields.isColor,
				mu:      tt.fields.mu,
			}
			if err := g.Success(tt.args.val...); (err != nil) != tt.wantErr {
				t.Errorf("Glg.Success() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGlg_Successf(t *testing.T) {
	type fields struct {
		writer  map[string]io.Writer
		std     map[string]io.Writer
		colors  map[string]func(string) string
		mode    int
		isColor bool
		mu      *sync.Mutex
	}
	type args struct {
		format string
		val    []interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Glg{
				writer:  tt.fields.writer,
				std:     tt.fields.std,
				colors:  tt.fields.colors,
				mode:    tt.fields.mode,
				isColor: tt.fields.isColor,
				mu:      tt.fields.mu,
			}
			if err := g.Successf(tt.args.format, tt.args.val...); (err != nil) != tt.wantErr {
				t.Errorf("Glg.Successf() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSuccess(t *testing.T) {
	type args struct {
		val []interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Success(tt.args.val...); (err != nil) != tt.wantErr {
				t.Errorf("Success() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSuccessf(t *testing.T) {
	type args struct {
		format string
		val    []interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Successf(tt.args.format, tt.args.val...); (err != nil) != tt.wantErr {
				t.Errorf("Successf() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGlg_Debug(t *testing.T) {
	type fields struct {
		writer  map[string]io.Writer
		std     map[string]io.Writer
		colors  map[string]func(string) string
		mode    int
		isColor bool
		mu      *sync.Mutex
	}
	type args struct {
		val []interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Glg{
				writer:  tt.fields.writer,
				std:     tt.fields.std,
				colors:  tt.fields.colors,
				mode:    tt.fields.mode,
				isColor: tt.fields.isColor,
				mu:      tt.fields.mu,
			}
			if err := g.Debug(tt.args.val...); (err != nil) != tt.wantErr {
				t.Errorf("Glg.Debug() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGlg_Debugf(t *testing.T) {
	type fields struct {
		writer  map[string]io.Writer
		std     map[string]io.Writer
		colors  map[string]func(string) string
		mode    int
		isColor bool
		mu      *sync.Mutex
	}
	type args struct {
		format string
		val    []interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Glg{
				writer:  tt.fields.writer,
				std:     tt.fields.std,
				colors:  tt.fields.colors,
				mode:    tt.fields.mode,
				isColor: tt.fields.isColor,
				mu:      tt.fields.mu,
			}
			if err := g.Debugf(tt.args.format, tt.args.val...); (err != nil) != tt.wantErr {
				t.Errorf("Glg.Debugf() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDebug(t *testing.T) {
	type args struct {
		val []interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Debug(tt.args.val...); (err != nil) != tt.wantErr {
				t.Errorf("Debug() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDebugf(t *testing.T) {
	type args struct {
		format string
		val    []interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Debugf(tt.args.format, tt.args.val...); (err != nil) != tt.wantErr {
				t.Errorf("Debugf() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGlg_Warn(t *testing.T) {
	type fields struct {
		writer  map[string]io.Writer
		std     map[string]io.Writer
		colors  map[string]func(string) string
		mode    int
		isColor bool
		mu      *sync.Mutex
	}
	type args struct {
		val []interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Glg{
				writer:  tt.fields.writer,
				std:     tt.fields.std,
				colors:  tt.fields.colors,
				mode:    tt.fields.mode,
				isColor: tt.fields.isColor,
				mu:      tt.fields.mu,
			}
			if err := g.Warn(tt.args.val...); (err != nil) != tt.wantErr {
				t.Errorf("Glg.Warn() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGlg_Warnf(t *testing.T) {
	type fields struct {
		writer  map[string]io.Writer
		std     map[string]io.Writer
		colors  map[string]func(string) string
		mode    int
		isColor bool
		mu      *sync.Mutex
	}
	type args struct {
		format string
		val    []interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Glg{
				writer:  tt.fields.writer,
				std:     tt.fields.std,
				colors:  tt.fields.colors,
				mode:    tt.fields.mode,
				isColor: tt.fields.isColor,
				mu:      tt.fields.mu,
			}
			if err := g.Warnf(tt.args.format, tt.args.val...); (err != nil) != tt.wantErr {
				t.Errorf("Glg.Warnf() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestWarn(t *testing.T) {
	type args struct {
		val []interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Warn(tt.args.val...); (err != nil) != tt.wantErr {
				t.Errorf("Warn() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestWarnf(t *testing.T) {
	type args struct {
		format string
		val    []interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Warnf(tt.args.format, tt.args.val...); (err != nil) != tt.wantErr {
				t.Errorf("Warnf() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGlg_CustomLog(t *testing.T) {
	type fields struct {
		writer  map[string]io.Writer
		std     map[string]io.Writer
		colors  map[string]func(string) string
		mode    int
		isColor bool
		mu      *sync.Mutex
	}
	type args struct {
		level string
		val   []interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Glg{
				writer:  tt.fields.writer,
				std:     tt.fields.std,
				colors:  tt.fields.colors,
				mode:    tt.fields.mode,
				isColor: tt.fields.isColor,
				mu:      tt.fields.mu,
			}
			if err := g.CustomLog(tt.args.level, tt.args.val...); (err != nil) != tt.wantErr {
				t.Errorf("Glg.CustomLog() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGlg_CustomLogf(t *testing.T) {
	type fields struct {
		writer  map[string]io.Writer
		std     map[string]io.Writer
		colors  map[string]func(string) string
		mode    int
		isColor bool
		mu      *sync.Mutex
	}
	type args struct {
		level  string
		format string
		val    []interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Glg{
				writer:  tt.fields.writer,
				std:     tt.fields.std,
				colors:  tt.fields.colors,
				mode:    tt.fields.mode,
				isColor: tt.fields.isColor,
				mu:      tt.fields.mu,
			}
			if err := g.CustomLogf(tt.args.level, tt.args.format, tt.args.val...); (err != nil) != tt.wantErr {
				t.Errorf("Glg.CustomLogf() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCustomLog(t *testing.T) {
	type args struct {
		level string
		val   []interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := CustomLog(tt.args.level, tt.args.val...); (err != nil) != tt.wantErr {
				t.Errorf("CustomLog() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCustomLogf(t *testing.T) {
	type args struct {
		level  string
		format string
		val    []interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := CustomLogf(tt.args.level, tt.args.format, tt.args.val...); (err != nil) != tt.wantErr {
				t.Errorf("CustomLogf() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGlg_Print(t *testing.T) {
	type fields struct {
		writer  map[string]io.Writer
		std     map[string]io.Writer
		colors  map[string]func(string) string
		mode    int
		isColor bool
		mu      *sync.Mutex
	}
	type args struct {
		val []interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Glg{
				writer:  tt.fields.writer,
				std:     tt.fields.std,
				colors:  tt.fields.colors,
				mode:    tt.fields.mode,
				isColor: tt.fields.isColor,
				mu:      tt.fields.mu,
			}
			if err := g.Print(tt.args.val...); (err != nil) != tt.wantErr {
				t.Errorf("Glg.Print() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGlg_Println(t *testing.T) {
	type fields struct {
		writer  map[string]io.Writer
		std     map[string]io.Writer
		colors  map[string]func(string) string
		mode    int
		isColor bool
		mu      *sync.Mutex
	}
	type args struct {
		val []interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Glg{
				writer:  tt.fields.writer,
				std:     tt.fields.std,
				colors:  tt.fields.colors,
				mode:    tt.fields.mode,
				isColor: tt.fields.isColor,
				mu:      tt.fields.mu,
			}
			if err := g.Println(tt.args.val...); (err != nil) != tt.wantErr {
				t.Errorf("Glg.Println() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGlg_Printf(t *testing.T) {
	type fields struct {
		writer  map[string]io.Writer
		std     map[string]io.Writer
		colors  map[string]func(string) string
		mode    int
		isColor bool
		mu      *sync.Mutex
	}
	type args struct {
		format string
		val    []interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Glg{
				writer:  tt.fields.writer,
				std:     tt.fields.std,
				colors:  tt.fields.colors,
				mode:    tt.fields.mode,
				isColor: tt.fields.isColor,
				mu:      tt.fields.mu,
			}
			if err := g.Printf(tt.args.format, tt.args.val...); (err != nil) != tt.wantErr {
				t.Errorf("Glg.Printf() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPrint(t *testing.T) {
	type args struct {
		val []interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Print(tt.args.val...); (err != nil) != tt.wantErr {
				t.Errorf("Print() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPrintln(t *testing.T) {
	type args struct {
		val []interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Println(tt.args.val...); (err != nil) != tt.wantErr {
				t.Errorf("Println() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPrintf(t *testing.T) {
	type args struct {
		format string
		val    []interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Printf(tt.args.format, tt.args.val...); (err != nil) != tt.wantErr {
				t.Errorf("Printf() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGlg_Error(t *testing.T) {
	type fields struct {
		writer  map[string]io.Writer
		std     map[string]io.Writer
		colors  map[string]func(string) string
		mode    int
		isColor bool
		mu      *sync.Mutex
	}
	type args struct {
		val []interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Glg{
				writer:  tt.fields.writer,
				std:     tt.fields.std,
				colors:  tt.fields.colors,
				mode:    tt.fields.mode,
				isColor: tt.fields.isColor,
				mu:      tt.fields.mu,
			}
			if err := g.Error(tt.args.val...); (err != nil) != tt.wantErr {
				t.Errorf("Glg.Error() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGlg_Errorf(t *testing.T) {
	type fields struct {
		writer  map[string]io.Writer
		std     map[string]io.Writer
		colors  map[string]func(string) string
		mode    int
		isColor bool
		mu      *sync.Mutex
	}
	type args struct {
		format string
		val    []interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Glg{
				writer:  tt.fields.writer,
				std:     tt.fields.std,
				colors:  tt.fields.colors,
				mode:    tt.fields.mode,
				isColor: tt.fields.isColor,
				mu:      tt.fields.mu,
			}
			if err := g.Errorf(tt.args.format, tt.args.val...); (err != nil) != tt.wantErr {
				t.Errorf("Glg.Errorf() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestError(t *testing.T) {
	type args struct {
		val []interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Error(tt.args.val...); (err != nil) != tt.wantErr {
				t.Errorf("Error() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestErrorf(t *testing.T) {
	type args struct {
		format string
		val    []interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Errorf(tt.args.format, tt.args.val...); (err != nil) != tt.wantErr {
				t.Errorf("Errorf() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGlg_Fail(t *testing.T) {
	type fields struct {
		writer  map[string]io.Writer
		std     map[string]io.Writer
		colors  map[string]func(string) string
		mode    int
		isColor bool
		mu      *sync.Mutex
	}
	type args struct {
		val []interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Glg{
				writer:  tt.fields.writer,
				std:     tt.fields.std,
				colors:  tt.fields.colors,
				mode:    tt.fields.mode,
				isColor: tt.fields.isColor,
				mu:      tt.fields.mu,
			}
			if err := g.Fail(tt.args.val...); (err != nil) != tt.wantErr {
				t.Errorf("Glg.Fail() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGlg_Failf(t *testing.T) {
	type fields struct {
		writer  map[string]io.Writer
		std     map[string]io.Writer
		colors  map[string]func(string) string
		mode    int
		isColor bool
		mu      *sync.Mutex
	}
	type args struct {
		format string
		val    []interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Glg{
				writer:  tt.fields.writer,
				std:     tt.fields.std,
				colors:  tt.fields.colors,
				mode:    tt.fields.mode,
				isColor: tt.fields.isColor,
				mu:      tt.fields.mu,
			}
			if err := g.Failf(tt.args.format, tt.args.val...); (err != nil) != tt.wantErr {
				t.Errorf("Glg.Failf() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFail(t *testing.T) {
	type args struct {
		val []interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Fail(tt.args.val...); (err != nil) != tt.wantErr {
				t.Errorf("Fail() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFailf(t *testing.T) {
	type args struct {
		format string
		val    []interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Failf(tt.args.format, tt.args.val...); (err != nil) != tt.wantErr {
				t.Errorf("Failf() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGlg_Fatal(t *testing.T) {
	type fields struct {
		writer  map[string]io.Writer
		std     map[string]io.Writer
		colors  map[string]func(string) string
		mode    int
		isColor bool
		mu      *sync.Mutex
	}
	type args struct {
		val []interface{}
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Glg{
				writer:  tt.fields.writer,
				std:     tt.fields.std,
				colors:  tt.fields.colors,
				mode:    tt.fields.mode,
				isColor: tt.fields.isColor,
				mu:      tt.fields.mu,
			}
			g.Fatal(tt.args.val...)
		})
	}
}

func TestGlg_Fatalln(t *testing.T) {
	type fields struct {
		writer  map[string]io.Writer
		std     map[string]io.Writer
		colors  map[string]func(string) string
		mode    int
		isColor bool
		mu      *sync.Mutex
	}
	type args struct {
		val []interface{}
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Glg{
				writer:  tt.fields.writer,
				std:     tt.fields.std,
				colors:  tt.fields.colors,
				mode:    tt.fields.mode,
				isColor: tt.fields.isColor,
				mu:      tt.fields.mu,
			}
			g.Fatalln(tt.args.val...)
		})
	}
}

func TestGlg_Fatalf(t *testing.T) {
	type fields struct {
		writer  map[string]io.Writer
		std     map[string]io.Writer
		colors  map[string]func(string) string
		mode    int
		isColor bool
		mu      *sync.Mutex
	}
	type args struct {
		format string
		val    []interface{}
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Glg{
				writer:  tt.fields.writer,
				std:     tt.fields.std,
				colors:  tt.fields.colors,
				mode:    tt.fields.mode,
				isColor: tt.fields.isColor,
				mu:      tt.fields.mu,
			}
			g.Fatalf(tt.args.format, tt.args.val...)
		})
	}
}

func TestFatal(t *testing.T) {
	type args struct {
		val []interface{}
	}
	tests := []struct {
		name string
		args args
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Fatal(tt.args.val...)
		})
	}
}

func TestFatalf(t *testing.T) {
	type args struct {
		format string
		val    []interface{}
	}
	tests := []struct {
		name string
		args args
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Fatalf(tt.args.format, tt.args.val...)
		})
	}
}

func TestFatalln(t *testing.T) {
	type args struct {
		val []interface{}
	}
	tests := []struct {
		name string
		args args
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Fatalln(tt.args.val...)
		})
	}
}
