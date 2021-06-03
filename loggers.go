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

// Package glg can quickly output that are colored and leveled logs with simple syntax
package glg

import (
	"sync"
	"sync/atomic"
	"unsafe"
)

type loggers struct {
	mu     sync.Mutex
	read   atomic.Value
	dirty  map[LEVEL]*entryLoggers
	misses int
}

type readOnlyLoggers struct {
	m       map[LEVEL]*entryLoggers
	amended bool
}

var expungedLoggers = unsafe.Pointer(new(*logger))

type entryLoggers struct {
	p unsafe.Pointer
}

func newEntryLoggers(i *logger) *entryLoggers {
	return &entryLoggers{p: unsafe.Pointer(&i)}
}

func (m *loggers) Load(key LEVEL) (value *logger, ok bool) {
	read, _ := m.read.Load().(readOnlyLoggers)
	e, ok := read.m[key]
	if !ok && read.amended {
		m.mu.Lock()
		read, _ = m.read.Load().(readOnlyLoggers)
		e, ok = read.m[key]
		if !ok && read.amended {
			e, ok = m.dirty[key]
			m.missLocked()
		}
		m.mu.Unlock()
	}
	if !ok {
		return value, false
	}
	return e.load()
}

func (e *entryLoggers) load() (value *logger, ok bool) {
	p := atomic.LoadPointer(&e.p)
	if p == nil || p == expungedLoggers {
		return value, false
	}
	return *(**logger)(p), true
}

func (m *loggers) Store(key LEVEL, value *logger) {
	read, _ := m.read.Load().(readOnlyLoggers)
	if e, ok := read.m[key]; ok && e.tryStore(&value) {
		return
	}

	m.mu.Lock()
	read, _ = m.read.Load().(readOnlyLoggers)
	if e, ok := read.m[key]; ok {
		if e.unexpungeLocked() {
			m.dirty[key] = e
		}
		e.storeLocked(&value)
	} else if e, ok := m.dirty[key]; ok {
		e.storeLocked(&value)
	} else {
		if !read.amended {
			m.dirtyLocked()
			m.read.Store(readOnlyLoggers{m: read.m, amended: true})
		}
		m.dirty[key] = newEntryLoggers(value)
	}
	m.mu.Unlock()
}

func (e *entryLoggers) tryStore(i **logger) bool {
	for {
		p := atomic.LoadPointer(&e.p)
		if p == expungedLoggers {
			return false
		}
		if atomic.CompareAndSwapPointer(&e.p, p, unsafe.Pointer(i)) {
			return true
		}
	}
}

func (e *entryLoggers) unexpungeLocked() (wasExpunged bool) {
	return atomic.CompareAndSwapPointer(&e.p, expungedLoggers, nil)
}

func (e *entryLoggers) storeLocked(i **logger) {
	atomic.StorePointer(&e.p, unsafe.Pointer(i))
}

func (m *loggers) Range(f func(key LEVEL, value *logger) bool) {
	read, _ := m.read.Load().(readOnlyLoggers)
	if read.amended {
		m.mu.Lock()
		read, _ = m.read.Load().(readOnlyLoggers)
		if read.amended {
			read = readOnlyLoggers{m: m.dirty}
			m.read.Store(read)
			m.dirty = nil
			m.misses = 0
		}
		m.mu.Unlock()
	}

	for k, e := range read.m {
		v, ok := e.load()
		if !ok {
			continue
		}
		if !f(k, v) {
			break
		}
	}
}

func (m *loggers) missLocked() {
	m.misses++
	if m.misses < len(m.dirty) {
		return
	}
	m.read.Store(readOnlyLoggers{m: m.dirty})
	m.dirty = nil
	m.misses = 0
}

func (m *loggers) dirtyLocked() {
	if m.dirty != nil {
		return
	}

	read, _ := m.read.Load().(readOnlyLoggers)
	m.dirty = make(map[LEVEL]*entryLoggers, len(read.m))
	for k, e := range read.m {
		if !e.tryExpungeLocked() {
			m.dirty[k] = e
		}
	}
}

func (e *entryLoggers) tryExpungeLocked() (isExpunged bool) {
	p := atomic.LoadPointer(&e.p)
	for p == nil {
		if atomic.CompareAndSwapPointer(&e.p, nil, expungedLoggers) {
			return true
		}
		p = atomic.LoadPointer(&e.p)
	}
	return p == expungedLoggers
}
