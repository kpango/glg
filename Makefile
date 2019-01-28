GO_VERSION:=$(shell go version)

.PHONY: bench profile clean test

all: install

bench:
	GO111MODULE=on go test -count=5 -run=NONE -bench . -benchmem

profile:
	mkdir bench
	GO111MODULE=on go test -count=10 -run=NONE -bench . -benchmem -o pprof/test.bin -cpuprofile pprof/cpu.out -memprofile pprof/mem.out
	GO111MODULE=on go tool pprof --svg pprof/test.bin pprof/mem.out > bench/mem.svg
	GO111MODULE=on go tool pprof --svg pprof/test.bin pprof/cpu.out > bench/cpu.svg
	rm -rf pprof
	mkdir pprof
	GO111MODULE=on go test -count=10 -run=NONE -bench=BenchmarkGlg -benchmem -o pprof/test.bin -cpuprofile pprof/cpu-glg.out -memprofile pprof/mem-glg.out
	GO111MODULE=on go tool pprof --svg pprof/test.bin pprof/cpu-glg.out > bench/cpu-glg.svg
	GO111MODULE=on go tool pprof --svg pprof/test.bin pprof/mem-glg.out > bench/mem-glg.svg
	rm -rf pprof
	mkdir pprof
	GO111MODULE=on go test -count=10 -run=NONE -bench=BenchmarkDefaultLog -benchmem -o pprof/test.bin -cpuprofile pprof/cpu-def.out -memprofile pprof/mem-def.out
	GO111MODULE=on go tool pprof --svg pprof/test.bin pprof/mem-def.out > bench/mem-def.svg
	GO111MODULE=on go tool pprof --svg pprof/test.bin pprof/cpu-def.out > bench/cpu-def.svg
	rm -rf pprof

clean:
	rm -rf bench
	rm -rf pprof
	rm -rf ./*.svg
	rm -rf ./*.log

test:
	GO111MODULE=on go test --race ./...

