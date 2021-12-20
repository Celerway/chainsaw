package chainsaw

import (
	"bytes"
	"strings"
	"sync"
)

// Various helper stuff for the tests.

// stringLogger implements io.Writer interface and is useful for testing
// things that write to io.Writers. Each write is stored in a string.
type stringLogger struct {
	loglines []string
	mu       sync.Mutex
}

func (s *stringLogger) Write(p []byte) (n int, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.loglines == nil {
		s.loglines = make([]string, 0)
	}
	s.loglines = append(s.loglines, string(p))
	return len(p), nil
}

func (s *stringLogger) contains(substring string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, line := range s.loglines {
		if strings.Contains(line, substring) {
			return true
		}
	}
	return false
}

type SafeInt struct {
	value int
	m     sync.RWMutex
}

func (i *SafeInt) Inc() {
	i.m.Lock()
	defer i.m.Unlock()
	i.value++
}
func (i *SafeInt) Get() int {
	i.m.RLock()
	defer i.m.RUnlock()
	return i.value
}

type SafeBuffer struct {
	b bytes.Buffer
	m sync.Mutex
}

func (b *SafeBuffer) Read(p []byte) (n int, err error) {
	b.m.Lock()
	defer b.m.Unlock()
	return b.b.Read(p)
}
func (b *SafeBuffer) Write(p []byte) (n int, err error) {
	b.m.Lock()
	defer b.m.Unlock()
	return b.b.Write(p)
}
func (b *SafeBuffer) String() string {
	b.m.Lock()
	defer b.m.Unlock()
	return b.b.String()
}

func (b *SafeBuffer) Bytes() []byte {
	b.m.Lock()
	defer b.m.Unlock()
	return b.b.Bytes()
}
