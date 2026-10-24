package decoy

func (d *Decoy) SetVariable(n int, ngrams map[string][]string) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	d.n = n
	d.ngrams = ngrams
}
