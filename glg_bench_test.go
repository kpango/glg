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

package glg_test

import (
	"log"
	"testing"

	"github.com/kpango/glg"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	testMsg = `benchmark sample message blow
MIT License

Copyright (c) 2019 kpango (Yusuke Kato)

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.`

	testInt   = 9999
	testFloat = 10.10

	testFormat = "format %s,\t%d,%f\n"

	testJSON = JSONMessage{
		Message: testMsg,
		Number:  testInt,
		Float:   testFloat,
	}
)

type JSONMessage struct {
	Message string  `json:"message,omitempty"`
	Number  int     `json:"number,omitempty"`
	Float   float64 `json:"float,omitempty"`
}

type MockWriter struct {
}

func (m MockWriter) Write(b []byte) (int, error) {
	_ = b
	return 0, nil
}

func BenchmarkDefaultLog(b *testing.B) {
	log.SetOutput(&MockWriter{})
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			log.Println("\t[" + "LOG" + "]:\t" + testMsg)
			log.Println("\t[" + "LOG" + "]:\t" + testMsg)
			log.Println("\t[" + "LOG" + "]:\t" + testMsg)
			log.Println("\t[" + "LOG" + "]:\t" + testMsg)
			log.Println("\t[" + "LOG" + "]:\t" + testMsg)
		}
	})
}

func BenchmarkGlg(b *testing.B) {
	glg.Reset()
	glg.Get().SetMode(glg.WRITER).SetWriter(&MockWriter{}).EnablePoolBuffer(100)
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			glg.Log(testMsg)
			glg.Log(testMsg)
			glg.Log(testMsg)
			glg.Log(testMsg)
			glg.Log(testMsg)
		}
	})
}

func BenchmarkZap(b *testing.B) {
	cfg := zap.NewProductionConfig()
	logger := zap.New(zapcore.NewCore(zapcore.NewConsoleEncoder(cfg.EncoderConfig), zapcore.AddSync(&MockWriter{}), cfg.Level))
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info(testMsg)
			logger.Info(testMsg)
			logger.Info(testMsg)
			logger.Info(testMsg)
			logger.Info(testMsg)
		}
	})
}

func BenchmarkDefaultLogf(b *testing.B) {
	log.SetOutput(&MockWriter{})
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			log.Printf("\t["+"LOG"+"]:\t"+testFormat+"\n", testMsg, testInt, testFloat)
			log.Printf("\t["+"LOG"+"]:\t"+testFormat+"\n", testMsg, testInt, testFloat)
			log.Printf("\t["+"LOG"+"]:\t"+testFormat+"\n", testMsg, testInt, testFloat)
			log.Printf("\t["+"LOG"+"]:\t"+testFormat+"\n", testMsg, testInt, testFloat)
			log.Printf("\t["+"LOG"+"]:\t"+testFormat+"\n", testMsg, testInt, testFloat)
		}
	})
}

func BenchmarkGlgf(b *testing.B) {
	glg.Reset()
	glg.Get().SetMode(glg.WRITER).SetWriter(&MockWriter{}).EnablePoolBuffer(100)
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			glg.Logf(testFormat, testMsg, testInt, testFloat)
			glg.Logf(testFormat, testMsg, testInt, testFloat)
			glg.Logf(testFormat, testMsg, testInt, testFloat)
			glg.Logf(testFormat, testMsg, testInt, testFloat)
			glg.Logf(testFormat, testMsg, testInt, testFloat)
		}
	})
}

func BenchmarkZapf(b *testing.B) {
	cfg := zap.NewProductionConfig()
	logger := zap.New(zapcore.NewCore(zapcore.NewConsoleEncoder(cfg.EncoderConfig), zapcore.AddSync(&MockWriter{}), cfg.Level)).Sugar()
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Infof(testFormat, testMsg, testInt, testFloat)
			logger.Infof(testFormat, testMsg, testInt, testFloat)
			logger.Infof(testFormat, testMsg, testInt, testFloat)
			logger.Infof(testFormat, testMsg, testInt, testFloat)
			logger.Infof(testFormat, testMsg, testInt, testFloat)
		}
	})
}

func BenchmarkGlgJSON(b *testing.B) {
	glg.Reset()
	glg.Get().SetMode(glg.WRITER).SetWriter(&MockWriter{}).EnablePoolBuffer(100).EnableJSON()
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			glg.Log(testJSON)
			glg.Log(testJSON)
			glg.Log(testJSON)
			glg.Log(testJSON)
			glg.Log(testJSON)
		}
	})
}

func BenchmarkZapJSON(b *testing.B) {
	cfg := zap.NewProductionConfig()
	logger := zap.New(zapcore.NewCore(zapcore.NewJSONEncoder(cfg.EncoderConfig), zapcore.AddSync(&MockWriter{}), cfg.Level))
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info("", zap.String("message", testJSON.Message),
				zap.Int("number", testJSON.Number),
				zap.Float64("float", testJSON.Float))
			logger.Info("", zap.String("message", testJSON.Message),
				zap.Int("number", testJSON.Number),
				zap.Float64("float", testJSON.Float))
			logger.Info("",zap.String("message", testJSON.Message),
				zap.Int("number", testJSON.Number),
				zap.Float64("float", testJSON.Float))
			logger.Info("", zap.String("message", testJSON.Message),
				zap.Int("number", testJSON.Number),
				zap.Float64("float", testJSON.Float))
			logger.Info("", zap.String("message", testJSON.Message),
				zap.Int("number", testJSON.Number),
				zap.Float64("float", testJSON.Float))
		}
	})
}

func BenchmarkZapSugarJSON(b *testing.B) {
	cfg := zap.NewProductionConfig()
	logger := zap.New(zapcore.NewCore(zapcore.NewJSONEncoder(cfg.EncoderConfig), zapcore.AddSync(&MockWriter{}), cfg.Level)).Sugar()
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info(testJSON)
			logger.Info(testJSON)
			logger.Info(testJSON)
			logger.Info(testJSON)
			logger.Info(testJSON)
		}
	})
}
