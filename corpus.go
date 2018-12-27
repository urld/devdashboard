// Copyright 2018 David Url.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package devdashboard

import (
	"context"
	"log"
	"sync"

	"github.com/urld/devdashboard/devdashpb"
)

type Corpus struct {
	mutationSource MutationSource
	mutationLogger MutationLogger
	verbose        bool

	mu sync.RWMutex // guards all following fields
	// state:
	didInit bool // true after Initialize completes successfully

	// issue tracker data:
	Projects     map[string]*Project
	TrackerUsers map[string]*IssueTrackerUser
	Milestones   map[string]*Milestone
	Releases     map[string]*Release
	Issues       map[string]*Issue

	// source data:
	GitRepos map[string]*GitRepo
}

// RLock grabs the corpus's read lock. Grabbing the read lock prevents
// any concurrent writes from mutating the corpus. This is only
// necessary if the application is querying the corpus and calling its
// Update method concurrently.
func (c *Corpus) RLock() { c.mu.RLock() }

// RUnlock unlocks the corpus's read lock.
func (c *Corpus) RUnlock() { c.mu.RUnlock() }

// Initialize populates the corpus using the data from the
// MutationSource. It returns once it's up-to-date. To incrementally
// update it later, use the Update method.
func (c *Corpus) Initialize(ctx context.Context, src MutationSource) error {
	if c.mutationSource != nil {
		panic("duplicate call to Initialize")
	}
	c.mutationSource = src

	c.Projects = make(map[string]*Project)
	c.TrackerUsers = make(map[string]*IssueTrackerUser)
	c.Milestones = make(map[string]*Milestone)
	c.Releases = make(map[string]*Release)
	c.Issues = make(map[string]*Issue)

	c.GitRepos = make(map[string]*GitRepo)

	log.Printf("Loading data from log %T ...", src)
	return c.update(ctx, nil)
}

// Update incrementally updates the corpus from its current state to
// the latest state from the MutationSource passed earlier to
// Initialize. It does not return until there's either a new change or
// the context expires.
//
// Update must not be called concurrently with any other Update calls. If
// reading the corpus concurrently while the corpus is updating, you must hold
// the read lock using Corpus.RLock.
func (c *Corpus) Update(ctx context.Context) error {
	if c.mutationSource == nil {
		panic("Update called without call to Initialize")
	}
	log.Printf("Updating data from log %T ...", c.mutationSource)
	return c.update(ctx, nil)
}

type noopLocker struct{}

func (noopLocker) Lock()   {}
func (noopLocker) Unlock() {}

// lk optionally specifies a locker to use while processing mutations.
func (c *Corpus) update(ctx context.Context, lk sync.Locker) error {
	src := c.mutationSource
	mutations := src.GetMutations(ctx)
	done := ctx.Done()
	c.mu.Lock()
	defer c.mu.Unlock()
	if lk == nil {
		lk = noopLocker{}
	}
	for {
		select {
		case <-done:
			err := ctx.Err()
			log.Printf("Context expired while loading data from log %T: %v", src, err)
			return err
		case e := <-mutations:
			if e.Err != nil {
				log.Printf("Corpus GetMutations: %v", e.Err)
				return e.Err
			}
			if e.End {
				c.didInit = true
				log.Printf("Reloaded data from log %T.", src)
				return nil
			}
			lk.Lock()
			c.processMutationLocked(e.Mutation)
			lk.Unlock()
		}
	}
}

// addMutation adds a mutation to the log and immediately processes it.
func (c *Corpus) addMutation(m *devdashpb.Mutation) {
	if c.verbose {
		log.Printf("mutation: %v", m)
	}
	c.mu.Lock()
	c.processMutationLocked(m)
	c.mu.Unlock()

	if c.mutationLogger == nil {
		return
	}
	err := c.mutationLogger.Log(m)
	if err != nil {
		log.Fatalf("could not log mutation %v: %v\n", m, err)
	}
}

// c.mu must be held.
func (c *Corpus) processMutationLocked(m *devdashpb.Mutation) {
	if pm := m.Project; pm != nil {
		c.processProjectMutation(pm)
	}
	if rm := m.Release; rm != nil {
		c.processReleaseMutation(rm)
	}
	if im := m.Issue; im != nil {
		c.processIssueMutation(im)
	}
	if gm := m.Git; gm != nil {
		c.processGitMutation(gm)
	}
}

// Check verifies the internal structure of the Corpus data structures.
// It is intended for tests and debugging.
func (c *Corpus) Check() error {
	// TODO
	return nil
}
