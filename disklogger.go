// Original work: Copyright 2017 The Go Authors.
// Modified work: Copyright 2018 David Url.
// Use of this source code is governed by a BSD-style
// license that can be found in the go.LICENSE file.

package devdashboard

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/urld/devdashboard/devdashpb"
	"github.com/urld/devdashboard/reclog"
)

// DiskMutationLogger logs mutations to disk.
type DiskMutationLogger struct {
	directory string

	mu   sync.Mutex
	done bool // true after first GetMutations
}

// NewDiskMutationLogger creates a new DiskMutationLogger, which will create
// mutations in the given directory.
func NewDiskMutationLogger(directory string) *DiskMutationLogger {
	if directory == "" {
		panic("empty directory")
	}
	return &DiskMutationLogger{directory: directory}
}

// filename returns the filename to write to. The oldest filename must come
// first in lexical order.
func (d *DiskMutationLogger) filename() string {
	now := time.Now().UTC()
	name := fmt.Sprintf("devdashboard-%s.mutlog", now.Format("2006-01-02"))
	return filepath.Join(d.directory, name)
}

// Log will write m to disk. If a mutation file does not exist for the current
// day, it will be created.
func (d *DiskMutationLogger) Log(m *devdashpb.Mutation) error {
	data, err := proto.Marshal(m)
	if err != nil {
		return err
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	return reclog.AppendRecordToFile(d.filename(), data)
}

func (d *DiskMutationLogger) ForeachFile(fn func(fullPath string, fi os.FileInfo) error) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.directory == "" {
		panic("empty directory")
	}
	// Walk guarantees that files are walked in lexical order, which we depend on.
	return filepath.Walk(d.directory, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if fi.IsDir() && path != filepath.Clean(d.directory) {
			return filepath.SkipDir
		}
		if !strings.HasPrefix(fi.Name(), "devdashboard-") {
			return nil
		}
		if !strings.HasSuffix(fi.Name(), ".mutlog") {
			return nil
		}
		return fn(path, fi)
	})
}

func (d *DiskMutationLogger) GetMutations(ctx context.Context) <-chan MutationStreamEvent {
	ch := make(chan MutationStreamEvent, 50)
	go func() {
		err := d.sendMutations(ctx, ch)
		final := MutationStreamEvent{Err: err}
		if err == nil {
			final.End = true
		}
		select {
		case ch <- final:
		case <-ctx.Done():
		}
	}()
	return ch
}

func (d *DiskMutationLogger) sendMutations(ctx context.Context, ch chan<- MutationStreamEvent) error {
	return d.ForeachFile(func(fullPath string, fi os.FileInfo) error {
		return reclog.ForeachFileRecord(fullPath, func(off int64, hdr, rec []byte) error {
			m := new(devdashpb.Mutation)
			if err := proto.Unmarshal(rec, m); err != nil {
				return err
			}
			select {
			case ch <- MutationStreamEvent{Mutation: m}:
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		})
	})
}
