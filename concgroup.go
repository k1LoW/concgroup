package concgroup

import (
	"context"
	"sync"

	"golang.org/x/sync/errgroup"
)

// Group is a collection of goroutines like errgroup.Group.
type Group struct {
	eg       *errgroup.Group
	mu       sync.Mutex
	locks    map[string]*sync.Mutex
	initOnce sync.Once
}

// WithContext returns a new Group and an associated Context like errgroup.Group.
func WithContext(ctx context.Context) (*Group, context.Context) {
	eg, ctx := errgroup.WithContext(ctx)
	return &Group{eg: eg}, ctx
}

// Go calls the given function in a new goroutine like errgroup.Group.
func (g *Group) Go(key string, f func() error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.init()
	mu, ok := g.locks[key]
	if !ok {
		mu = &sync.Mutex{}
		g.locks[key] = mu
	}
	g.eg.Go(func() error {
		mu.Lock()
		defer mu.Unlock()
		return f()
	})
}

// TryGo calls the given function only when the number of active goroutines is currently below the configured limit like errgroup.Group.
func (g *Group) TryGo(key string, f func() error) bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.init()
	mu, ok := g.locks[key]
	if !ok {
		mu = &sync.Mutex{}
		g.locks[key] = mu
	}
	if !g.eg.TryGo(func() error {
		mu.Lock()
		defer mu.Unlock()
		return f()
	}) {
		return false
	}
	return true
}

// SetLimit limits the number of active goroutines in this group to at most n like errgroup.Group.
func (g *Group) SetLimit(n int) {
	g.init()
	g.eg.SetLimit(n)
}

// Wait blocks until all function calls from the Go method have returned like errgroup.Group.
func (g *Group) Wait() error {
	return g.eg.Wait()
}

func (g *Group) init() {
	g.initOnce.Do(func() {
		if g.eg == nil {
			g.eg = &errgroup.Group{}
		}
		if g.locks == nil {
			g.locks = map[string]*sync.Mutex{}
		}
	})
}
