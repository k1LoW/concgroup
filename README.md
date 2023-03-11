# concgroup

`concgroup` provides almost the same features to goroutine groups as `golang.org/x/sync/errgroup` but ensures sequential run of goroutines with the same group key.

## Usage

``` go
package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/k1LoW/concgroup"
)

func main() {
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
```
