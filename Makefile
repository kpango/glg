GO_VERSION:=$(shell go version)

.PHONY: all clean bench bench-all profile lint test contributors update install

all: clean install lint test bench

clean:
	go clean ./...
	rm -rf ./*.log
	rm -rf ./*.svg
	rm -rf ./go.mod
	rm -rf ./go.sum
	rm -rf bench
	rm -rf pprof
	rm -rf vendor


bench: clean init
	go test -count=5 -run=NONE -bench . -benchmem

init:
	go mod init
	go mod tidy
	sleep 3

profile: clean init
	rm -rf bench
	mkdir bench
	mkdir pprof
	\
	go test -count=10 -run=NONE -bench=BenchmarkGlg -benchmem -o pprof/glg-test.bin -cpuprofile pprof/cpu-glg.out -memprofile pprof/mem-glg.out
	go tool pprof --svg pprof/glg-test.bin pprof/cpu-glg.out > cpu-glg.svg
	go tool pprof --svg pprof/glg-test.bin pprof/mem-glg.out > mem-glg.svg
	\
	go test -count=10 -run=NONE -bench=BenchmarkDefaultLog -benchmem -o pprof/default-test.bin -cpuprofile pprof/cpu-default.out -memprofile pprof/mem-default.out
	go tool pprof --svg pprof/default-test.bin pprof/mem-default.out > mem-default.svg
	go tool pprof --svg pprof/default-test.bin pprof/cpu-default.out > cpu-default.svg
	\
	mv ./*.svg bench/

profile-web:
	go tool pprof -http=":6061" \
		pprof/glg-test.bin \
		pprof/cpu-glg.out &
	go tool pprof -http=":6062" \
		pprof/glg-test.bin \
		pprof/mem-glg.out

cpu:
	go tool pprof pprof/glg-test.bin pprof/cpu-glg.out

mem:
	go tool pprof --alloc_space pprof/glg-test.bin pprof/mem-glg.out

lint:
	gometalinter --enable-all . | rg -v comment

test: clean init
	GO111MODULE=on go test --race -timeout 1h -v $(go list ./... | rg -v vendor)

contributors:
	git log --format='%aN <%aE>' | sort -fu > CONTRIBUTORS
