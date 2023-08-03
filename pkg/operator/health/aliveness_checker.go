package health

import (
	"runtime"
	"sync"

	"k8s.io/klog/v2"
)

type AlivenessChecker interface {
	Alive() bool
}

type MultiAlivenessChecker struct {
	m sync.Mutex
	// name -> checker
	checkerMap map[string]AlivenessChecker
}

func (r *MultiAlivenessChecker) Add(name string, c AlivenessChecker) {
	r.m.Lock()
	defer r.m.Unlock()

	r.checkerMap[name] = c
}

func (r *MultiAlivenessChecker) Alive() bool {
	r.m.Lock()
	defer r.m.Unlock()

	for s, checker := range r.checkerMap {
		if !checker.Alive() {
			klog.Warningf("Controller [%s] didn't sync for a long time, declaring unhealthy and dumping stack", s)
			// 12 mb should be enough for a full goroutine dump
			buf := make([]byte, 1024*1024*12)
			n := runtime.Stack(buf, true)
			klog.Warningf("%s", buf[:n])

			return false
		}
	}

	return true
}

func NewMultiAlivenessChecker() *MultiAlivenessChecker {
	return &MultiAlivenessChecker{
		m:          sync.Mutex{},
		checkerMap: make(map[string]AlivenessChecker),
	}
}