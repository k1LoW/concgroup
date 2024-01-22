package concgroup_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/k1LoW/concgroup"
)

//nolint:gosec
// Original example is https://pkg.go.dev/golang.org/x/sync/errgroup#example-Group-JustErrors
func ExampleGroup() {
	cg := new(concgroup.Group)
	var urlgroups = map[string][]string{
		"go": {
			"https://go.dev/",
			"https://go.dev/dl/",
		},
		"google": {
			"http://www.google.com/",
		},
	}
	for key, ug := range urlgroups {
		key := key
		for _, url := range ug {
			url := url // https://golang.org/doc/faq#closures_and_goroutines
			cg.Go(key, func() error {
				// Fetch URL sequentially by key
				resp, err := http.Get(url)
				if err == nil {
					resp.Body.Close()
				}
				return err
			})
		}
	}
	// Wait for all HTTP fetches to complete.
	if err := cg.Wait(); err == nil {
		fmt.Println("Successfully fetched all URLs.")
	}
}

func TestSimple(t *testing.T) {
	t.Parallel()
	cg := &concgroup.Group{}
	cg.Go("one", func() error {
		return nil
	})
	cg.Go("two", func() error {
		return nil
	})
	if err := cg.Wait(); err != nil {
		t.Error(err)
	}
}

func TestConcurrencyGroup(t *testing.T) {
	t.Parallel()
	cg := new(concgroup.Group)
	mu := sync.Mutex{}
	for i := 0; i < 10; i++ {
		cg.Go("samegroup", func() error {
			if !mu.TryLock() {
				return errors.New("violate group concurrency")
			}
			defer mu.Unlock()
			time.Sleep(50 * time.Millisecond)
			return nil
		})
	}
	if err := cg.Wait(); err != nil {
		t.Error(err)
	}
}

func TestConcurrencyGroupWithContext(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	cg, _ := concgroup.WithContext(ctx)
	mu := sync.Mutex{}
	for i := 0; i < 10; i++ {
		cg.Go("samegroup", func() error {
			if !mu.TryLock() {
				return errors.New("violate group concurrency")
			}
			defer mu.Unlock()
			time.Sleep(50 * time.Millisecond)
			return nil
		})
	}
	if err := cg.Wait(); err != nil {
		t.Error(err)
	}
}

func TestConcurrencyGroupWithSetLimit(t *testing.T) {
	t.Parallel()
	const loop = 10
	cg := new(concgroup.Group)
	cg.SetLimit(1)
	mu := sync.Mutex{}
	call := 0
	for i := 0; i < loop; i++ {
		cg.Go("samegroup", func() error {
			if !mu.TryLock() {
				return errors.New("violate group concurrency")
			}
			call++
			defer mu.Unlock()
			return nil
		})
	}
	if err := cg.Wait(); err != nil {
		t.Error(err)
	}
	if call != loop {
		t.Error("Failed to Go")
	}
}

func TestConcurrencyGroupWithTryGo(t *testing.T) {
	t.Parallel()
	const loop = 10
	cg := new(concgroup.Group)
	cg.SetLimit(1)
	mu := sync.Mutex{}
	call := 0
	for i := 0; i < loop; i++ {
		cg.TryGo("samegroup", func() error {
			if !mu.TryLock() {
				return errors.New("violate group concurrency")
			}
			call++
			defer mu.Unlock()
			time.Sleep(50 * time.Millisecond)
			return nil
		})
	}
	if err := cg.Wait(); err != nil {
		t.Error(err)
	}
	if call == loop {
		t.Error("Failed to skip by TryGo")
	}
}

func TestConcurrencyGroupMulti(t *testing.T) {
	t.Parallel()
	cg := new(concgroup.Group)
	mu := sync.Mutex{}
	for i := 0; i < 10; i++ {
		keys := []string{"samegroup", fmt.Sprintf("group-%d", i)}
		cg.GoMulti(keys, func() error {
			if !mu.TryLock() {
				return errors.New("violate group concurrency")
			}
			defer mu.Unlock()
			time.Sleep(50 * time.Millisecond)
			return nil
		})
	}
	if err := cg.Wait(); err != nil {
		t.Error(err)
	}
}

func TestConcurrencyGroupWithTryGoMulti(t *testing.T) {
	t.Parallel()
	const loop = 10
	cg := new(concgroup.Group)
	cg.SetLimit(1)
	mu := sync.Mutex{}
	call := 0
	for i := 0; i < loop; i++ {
		keys := []string{"samegroup", fmt.Sprintf("group-%d", i)}
		cg.TryGoMulti(keys, func() error {
			if !mu.TryLock() {
				return errors.New("violate group concurrency")
			}
			call++
			defer mu.Unlock()
			time.Sleep(50 * time.Millisecond)
			return nil
		})
	}
	if err := cg.Wait(); err != nil {
		t.Error(err)
	}
	if call == loop {
		t.Error("Failed to skip by TryGoMulti")
	}
}
