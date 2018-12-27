// Copyright 2018 David Url.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package devdashboard

import (
	"context"

	"github.com/urld/devdashboard/devdashpb"
)

// A MutationSource yields a log of mutations that will catch a corpus
// back up to the present.
type MutationSource interface {
	// GetMutations returns a channel of mutations or related events.
	// The channel will never be closed.
	// All sends on the returned channel should select
	// on the provided context.
	GetMutations(context.Context) <-chan MutationStreamEvent
}

// MutationStreamEvent represents one of three possible events while
// reading mutations from disk. An event is either a mutation, an
// error, or reaching the current end of the log. Only one of the
// fields will be non-zero.
type MutationStreamEvent struct {
	Mutation *devdashpb.Mutation

	// Err is a fatal error reading the log. No other events will
	// follow an Err.
	Err error

	// End, if true, means that all mutations have been sent and
	// the next event might take some time to arrive (it might not
	// have occurred yet). The End event is not a terminal state
	// like Err. There may be multiple Ends.
	End bool
}

// A MutationLogger logs mutations.
type MutationLogger interface {
	Log(*devdashpb.Mutation) error
}
