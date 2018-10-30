package main

import (
	"sync"

	"github.com/kpango/glg"
)

func main() {
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		glg.Info("test1")
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		glg.Info("test2")
	}()

	wg.Wait()
}
