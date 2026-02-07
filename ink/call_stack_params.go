package ink

// CanPopThread returns true if there is more than one thread.
func (cs *CallStack) CanPopThread() bool {
	return len(cs.Threads) > 1
}

// PopThread pops the current thread.
func (cs *CallStack) PopThread() {
	if cs.CanPopThread() {
		cs.Threads = cs.Threads[:len(cs.Threads)-1]
	}
}
