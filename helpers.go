// Copyright 2018 David Url.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package devdashboard

import (
	"log"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/urld/devdashboard/devdashpb"
)

func pbTimestamp(t time.Time) *timestamp.Timestamp {
	ts, err := ptypes.TimestampProto(t)
	if err != nil {
		log.Printf("could not convert to protobuf timestamp:  %v", err)
	}
	return ts
}

func pbTime(ts *timestamp.Timestamp) time.Time {
	t, err := ptypes.Timestamp(ts)
	if err != nil {
		log.Printf("could not convert protobuf timestamp:  %v", err)
	}
	return t
}

func pbBool(val bool) *devdashpb.BoolChange {
	return &devdashpb.BoolChange{Val: val}
}

type set struct {
	m map[string]struct{}
}

func newSet() set {
	return set{m: make(map[string]struct{})}
}

func (s *set) put(id string) {
	s.m[id] = struct{}{}
}

func (s *set) has(id string) bool {
	_, ok := s.m[id]
	return ok
}
