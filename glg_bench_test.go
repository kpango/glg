package glg_test

import (
	"log"
	"testing"

	"github.com/kpango/glg"
)

var (
	testMsg = `benchmark sample message blow
MIT License

Copyright (c) 2017 Yusuke Kato

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
)

type MockWriter struct {
	msg string
}

func (m MockWriter) Write(b []byte) (int, error) {
	_ = string(b)
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
	glg.Get().SetMode(glg.WRITER).SetWriter(&MockWriter{})
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
	glg.Get().SetMode(glg.WRITER).SetWriter(&MockWriter{})
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
