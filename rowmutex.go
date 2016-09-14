package rowmutex

import "sync"

type row struct {
	cond      *sync.Cond
	waitCount int
}

type Table struct {
	mu sync.Mutex
	m  map[string]*row
}

func (t *Table) Do(key string, fn func() error) error {
	t.mu.Lock()
	if t.m == nil {
		t.m = make(map[string]*row)
	}
	if r, ok := t.m[key]; ok {
		r.waitCount++
		r.cond.Wait()
		r.waitCount--
	} else {
		t.m[key] = &row{
			cond:      sync.NewCond(&t.mu),
			waitCount: 0,
		}
	}
	t.mu.Unlock()

	fnErr := fn()

	t.mu.Lock()
	r := t.m[key]
	if r.waitCount <= 0 {
		delete(t.m, key)
	} else {
		r.cond.Signal()
	}
	t.mu.Unlock()
	return fnErr
}