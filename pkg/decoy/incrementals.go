package decoy

import "sync/atomic"

func (d *Decoy) NextIncrementalInt(id string, start, step int) int64 {
	d.mutex.RLock()
	counter, exists := d.intIncrementals[id]
	d.mutex.RUnlock()
	if exists {
		return counter.Add(int64(step))
	}

	d.mutex.Lock()
	defer d.mutex.Unlock()

	// Someone may have created it while we waited for the write lock.
	counter, exists = d.intIncrementals[id]
	if exists {
		return counter.Add(int64(step))
	}

	counter = &atomic.Int64{}
	counter.Store(int64(start))
	d.intIncrementals[id] = counter
	return int64(start)
}

func (d *Decoy) SetIncrementalInt(id string, value int) int {
	d.mutex.RLock()
	counter, exists := d.intIncrementals[id]
	d.mutex.RUnlock()
	if exists {
		counter.Store(int64(value))
		return value
	}

	d.mutex.Lock()
	defer d.mutex.Unlock()

	// Someone may have created it while we waited for the write lock.
	counter, exists = d.intIncrementals[id]
	if !exists {
		counter = &atomic.Int64{}
		d.intIncrementals[id] = counter
	}
	counter.Store(int64(value))
	return value
}

func (d *Decoy) DeleteIncrementalInt(id string) any {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	delete(d.intIncrementals, id)
	return nil
}
