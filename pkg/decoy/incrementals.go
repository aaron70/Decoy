package decoy

import "sync/atomic"

func (d *Decoy) NextIncrementalInt(id string, start, step int) int64 {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	value, exists := d.intIncrementals[id]
	if !exists {
		a := &atomic.Int64{}
		a.Store(int64(start))
		d.intIncrementals[id] = a
		return int64(start)
	}
	return value.Add(int64(step))
}

func (d *Decoy) SetIncrementalInt(id string, value int64) int64 {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	incremental, exists := d.intIncrementals[id]
	if !exists {
		incremental = &atomic.Int64{}
		d.intIncrementals[id] = incremental
	}
	incremental.Store(value)
	return value
}

func (d *Decoy) DeleteIncrementalInt(id string) any {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	delete(d.intIncrementals, id)
	return nil
}
