package decoy

import "sync/atomic"

func (d *Decoy) NextIncrementalInt(id string, start, step int64) int64 {
	d.incrementalMu.RLock()
	counter, exists := d.intIncrementals[id]
	d.incrementalMu.RUnlock()
	if exists {
		return counter.Add(step)
	}

	d.incrementalMu.Lock()
	defer d.incrementalMu.Unlock()

	d.ensureIncrementalsLocked()
	counter, exists = d.intIncrementals[id]
	if exists {
		return counter.Add(step)
	}

	counter = &atomic.Int64{}
	counter.Store(start)
	d.intIncrementals[id] = counter
	return start
}

func (d *Decoy) SetIncrementalInt(id string, value int64) int64 {
	d.incrementalMu.Lock()
	defer d.incrementalMu.Unlock()

	d.ensureIncrementalsLocked()
	counter, exists := d.intIncrementals[id]
	if !exists {
		counter = &atomic.Int64{}
		d.intIncrementals[id] = counter
	}
	counter.Store(value)
	return value
}

func (d *Decoy) DeleteIncrementalInt(id string) bool {
	d.incrementalMu.Lock()
	defer d.incrementalMu.Unlock()

	if _, exists := d.intIncrementals[id]; !exists {
		return false
	}
	delete(d.intIncrementals, id)
	return true
}
