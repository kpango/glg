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

type levelMap struct {
	mu     sync.Mutex
	read   atomic.Value
	dirty  map[string]*entryLevelMap
	misses int
}

type readOnlyLevelMap struct {
	m       map[string]*entryLevelMap
	amended bool
}

var expungedLevelMap = unsafe.Pointer(new(LEVEL))

type entryLevelMap struct {
	p unsafe.Pointer
}

func newEntryLevelMap(i LEVEL) *entryLevelMap {
	return &entryLevelMap{p: unsafe.Pointer(&i)}
}

func (m *levelMap) Load(key string) (value LEVEL, ok bool) {
	read, _ := m.read.Load().(readOnlyLevelMap)
	e, ok := read.m[key]
	if !ok && read.amended {
		m.mu.Lock()
		read, _ = m.read.Load().(readOnlyLevelMap)
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

func (e *entryLevelMap) load() (value LEVEL, ok bool) {
	p := atomic.LoadPointer(&e.p)
	if p == nil || p == expungedLevelMap {
		return value, false
	}
	return *(*LEVEL)(p), true
}

func (m *levelMap) Store(key string, value LEVEL) {
	read, _ := m.read.Load().(readOnlyLevelMap)
	if e, ok := read.m[key]; ok && e.tryStore(&value) {
		return
	}

	m.mu.Lock()
	read, _ = m.read.Load().(readOnlyLevelMap)
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
			m.read.Store(readOnlyLevelMap{m: read.m, amended: true})
		}
		m.dirty[key] = newEntryLevelMap(value)
	}
	m.mu.Unlock()
}

func (e *entryLevelMap) tryStore(i *LEVEL) bool {
	for {
		p := atomic.LoadPointer(&e.p)
		if p == expungedLevelMap {
			return false
		}
		if atomic.CompareAndSwapPointer(&e.p, p, unsafe.Pointer(i)) {
			return true
		}
	}
}

func (e *entryLevelMap) unexpungeLocked() (wasExpunged bool) {
	return atomic.CompareAndSwapPointer(&e.p, expungedLevelMap, nil)
}

func (e *entryLevelMap) storeLocked(i *LEVEL) {
	atomic.StorePointer(&e.p, unsafe.Pointer(i))
}

func (m *levelMap) missLocked() {
	m.misses++
	if m.misses < len(m.dirty) {
		return
	}
	m.read.Store(readOnlyLevelMap{m: m.dirty})
	m.dirty = nil
	m.misses = 0
}

func (m *levelMap) dirtyLocked() {
	if m.dirty != nil {
		return
	}

	read, _ := m.read.Load().(readOnlyLevelMap)
	m.dirty = make(map[string]*entryLevelMap, len(read.m))
	for k, e := range read.m {
		if !e.tryExpungeLocked() {
			m.dirty[k] = e
		}
	}
}

func (e *entryLevelMap) tryExpungeLocked() (isExpunged bool) {
	p := atomic.LoadPointer(&e.p)
	for p == nil {
		if atomic.CompareAndSwapPointer(&e.p, nil, expungedLevelMap) {
			return true
		}
		p = atomic.LoadPointer(&e.p)
	}
	return p == expungedLevelMap
}
