// Original work: Copyright 2017 The Go Authors.
// Modified work: Copyright 2018 David Url.
// Use of this source code is governed by a BSD-style
// license that can be found in the go.LICENSE file.

package devdashdata

import (
	"context"
	"sync"
	"testing"

	"github.com/urld/devdashboard"
)

func BenchmarkGet(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := Get(context.Background(), DefaultDir())
		if err != nil {
			b.Fatal(err)
		}
	}
}

var (
	corpusMu    sync.Mutex
	corpusCache *devdashboard.Corpus
)

func getData(tb testing.TB) *devdashboard.Corpus {
	if testing.Short() {
		tb.Skip("not running tests requiring large download in short mode")
	}
	corpusMu.Lock()
	defer corpusMu.Unlock()
	if corpusCache != nil {
		return corpusCache
	}
	var err error
	corpusCache, err = Get(context.Background(), DefaultDir())
	if err != nil {
		tb.Fatalf("getting corpus: %v", err)
	}
	return corpusCache
}

func TestCorpusCheck(t *testing.T) {
	c := getData(t)
	if err := c.Check(); err != nil {
		t.Fatal(err)
	}
}
