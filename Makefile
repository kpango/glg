GO_VERSION:=$(shell go version)

.PHONY: bench profile

all: install

bench:
	go test -count=5 -run=NONE -bench . -benchmem

profile:
	mkdir bench
	go test -count=10 -run=NONE -bench . -benchmem -o pprof/test.bin -cpuprofile pprof/cpu.out -memprofile pprof/mem.out
	go tool pprof --svg pprof/test.bin pprof/mem.out > mem.svg
	go tool pprof --svg pprof/test.bin pprof/cpu.out > cpu.svg
	rm -rf pprof
	go test -count=10 -run=NONE -bench=BenchmarkGlg -benchmem -o pprof/test.bin -cpuprofile pprof/cpu-glg.out -memprofile pprof/mem-glg.out
	go tool pprof --svg pprof/test.bin pprof/cpu-glg.out > cpu-glg.svg
	go tool pprof --svg pprof/test.bin pprof/mem-glg.out > mem-glg.svg
	rm -rf pprof
	go test -count=10 -run=NONE -bench=BenchmarkDefaultLog -benchmem -o pprof/test.bin -cpuprofile pprof/cpu-def.out -memprofile pprof/mem-def.out
	go tool pprof --svg pprof/test.bin pprof/mem-def.out > mem-def.svg
	go tool pprof --svg pprof/test.bin pprof/cpu-def.out > cpu-def.svg
	rm -rf pprof
	mv ./*.svg bench/
