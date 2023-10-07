package tezos

// listener is the state we hold in memory for each individual listener that has been added
type listener struct {
	hwmBlock        int64
	catchupLoopDone chan struct{}
}
