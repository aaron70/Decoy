package decoy

import "sync/atomic"

func (d *Decoy) CurrentIncrementalInt(id string, defaultValue int64) int64 {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	value, exists := d.intIncrementals[id]
	if !exists {
		return defaultValue
	}
	return value.Load()
}

func (d *Decoy) NextIncrementalInt(id string, start, step int64) int64 {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	value, exists := d.intIncrementals[id]
	if !exists {
		a := &atomic.Int64{}
		a.Store(start)
		d.intIncrementals[id] = a
		return start
	}
	return value.Add(step)
}
