package spam

import (
	"strings"
	"time"
)

type item struct {
	expiry int64
	key    string
}

type Filter struct {
	window int64
	seen   map[string]int64
	q      []item
	head   int
}

func New(window time.Duration) *Filter {
	return &Filter{
		window: window.Nanoseconds(),
		seen:   make(map[string]int64),
		q:      make([]item, 0, 1024),
	}
}

func (f *Filter) Allow(now time.Time, msg string) bool {
	t := now.UnixNano()

	for f.head < len(f.q) && f.q[f.head].expiry <= t {
		it := f.q[f.head]
		f.head++

		if cur, ok := f.seen[it.key]; ok && cur == it.expiry {
			delete(f.seen, it.key)
		}
	}

	if f.head > 1024 && f.head*2 >= len(f.q) {
		copy(f.q, f.q[f.head:])
		f.q = f.q[:len(f.q)-f.head]
		f.head = 0
	}

	key := strings.Clone(msg)
	if exp, ok := f.seen[key]; ok && exp > t {
		return false
	}

	exp := t + f.window
	f.seen[key] = exp
	f.q = append(f.q, item{expiry: exp, key: key})
	return true
}
