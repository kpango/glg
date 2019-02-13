GO_VERSION:=$(shell go version)

.PHONY: all clean bench bench-all profile lint test contributors update install

all: clean install lint test bench

clean:
	go clean ./...
	rm -rf ./*.log
	rm -rf ./*.svg
	rm -rf ./go.*
	rm -rf bench
	rm -rf pprof
	rm -rf vendor


bench: clean init
	go test -count=5 -run=NONE -bench . -benchmem

init:
	GO111MODULE=on go mod init
	GO111MODULE=on go mod vendor

profile: clean init
	rm -rf bench
	mkdir bench
	mkdir pprof
	go test -count=10 -run=NONE -bench . -benchmem -o pprof/test.bin -cpuprofile pprof/cpu.out -memprofile pprof/mem.out
	go tool pprof --svg pprof/test.bin pprof/mem.out > bench/mem.svg
	go tool pprof --svg pprof/test.bin pprof/cpu.out > bench/cpu.svg
	rm -rf pprof
	mkdir pprof
	go test -count=10 -run=NONE -bench=BenchmarkGlg -benchmem -o pprof/test.bin -cpuprofile pprof/cpu-glg.out -memprofile pprof/mem-glg.out
	go tool pprof --svg pprof/test.bin pprof/cpu-glg.out > bench/cpu-glg.svg
	go tool pprof --svg pprof/test.bin pprof/mem-glg.out > bench/mem-glg.svg
	go-torch -f bench/cpu-glg-graph.svg pprof/test.bin pprof/cpu-glg.out
	go-torch --alloc_objects -f bench/mem-glg-graph.svg pprof/test.bin pprof/mem-glg.out
	rm -rf pprof
	mkdir pprof
	go test -count=10 -run=NONE -bench=BenchmarkDefaultLog -benchmem -o pprof/test.bin -cpuprofile pprof/cpu-def.out -memprofile pprof/mem-def.out
	go tool pprof --svg pprof/test.bin pprof/mem-def.out > bench/mem-def.svg
	go tool pprof --svg pprof/test.bin pprof/cpu-def.out > bench/cpu-def.svg
	go-torch -f bench/cpu-def-graph.svg pprof/test.bin pprof/cpu-def.out
	go-torch --alloc_objects -f bench/mem-def-graph.svg pprof/test.bin pprof/mem-def.out
	rm -rf pprof

test: clean init
	go test --race -v $(go list ./... | rg -v vendor)

contributors:
	git log --format='%aN <%aE>' | sort -fu > CONTRIBUTORS
