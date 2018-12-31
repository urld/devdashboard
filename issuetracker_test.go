// Copyright 2018 David Url.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package devdashboard

import (
	"context"
	"testing"

	"github.com/golang/protobuf/ptypes"
	"github.com/urld/devdashboard/devdashpb"
)

type mockLogger struct {
	ch chan MutationStreamEvent
}

func newLogger() *mockLogger {
	return &mockLogger{
		ch: make(chan MutationStreamEvent, 50),
	}
}

func (l *mockLogger) Log(m *devdashpb.Mutation) error {
	l.ch <- MutationStreamEvent{Mutation: m}
	return nil
}

func (l *mockLogger) GetMutations(context.Context) <-chan MutationStreamEvent {
	return l.ch
}

func (l *mockLogger) end() {
	l.ch <- MutationStreamEvent{End: true}
}

func (l *mockLogger) err(err error) {
	l.ch <- MutationStreamEvent{Err: err}
}

func TestSepratateIssueMutations(t *testing.T) {
	l := newLogger()
	c := &Corpus{}

	checkErr(t, l.Log(&devdashpb.Mutation{
		Issue: &devdashpb.IssueMutation{
			Id:        "i1",
			Project:   "ABC",
			IssueKey:  "ABC-1",
			Title:     "Setup project",
			Body:      "* create git repo\n* write readme\n* configure ci build",
			Status:    "Done",
			Closed:    pbBool(true),
			ClosedAt:  ptypes.TimestampNow(),
			ClosedBy:  &devdashpb.TrackerUser{Id: "urld", Name: "David Url", Email: "david@urld.io"},
			Created:   ptypes.TimestampNow(),
			Updated:   ptypes.TimestampNow(),
			Assignees: []*devdashpb.TrackerUser{{Id: "urld", Name: "David Url", Email: "david@urld.io"}},
			Owner:     &devdashpb.TrackerUser{Id: "urld", Name: "David Url", Email: "david@urld.io"},
		},
	}))
	checkErr(t, l.Log(&devdashpb.Mutation{
		Issue: &devdashpb.IssueMutation{
			Id:        "i2",
			Project:   "ABC",
			IssueKey:  "ABC-2",
			Title:     "service specs",
			Body:      "REST service specification",
			Status:    "New",
			Created:   ptypes.TimestampNow(),
			Updated:   ptypes.TimestampNow(),
			Assignees: []*devdashpb.TrackerUser{{Id: "urld", Name: "David Url", Email: "david@urld.io"}},
			Owner:     &devdashpb.TrackerUser{Id: "urld", Name: "David Url", Email: "david@urld.io"},
		},
	}))
	checkErr(t, l.Log(&devdashpb.Mutation{
		Issue: &devdashpb.IssueMutation{
			Id:        "i3",
			Project:   "DEF",
			IssueKey:  "DEF-1",
			Title:     "client prototype",
			Body:      "prototype of a http client for service x",
			Status:    "In Progress",
			Created:   ptypes.TimestampNow(),
			Updated:   ptypes.TimestampNow(),
			Assignees: []*devdashpb.TrackerUser{{Id: "urld", Name: "David Url", Email: "david@urld.io"}},
			Owner:     &devdashpb.TrackerUser{Id: "urld", Name: "David Url", Email: "david@urld.io"},
		},
	}))

	l.end()
	checkErr(t, c.Initialize(context.Background(), l))

	abc, ok := c.Projects["ABC"]
	if !ok {
		t.Fatal("Project ABC should have been created")
	}
	pi1, ok := abc.Issues["i1"]
	if !ok {
		t.Fatal("Project ABC should have issue i1")
	}
	if pi1.p != abc {
		t.Error("Issue i1 should point to project ABC")
	}
	pi2, ok := abc.Issues["i2"]
	if !ok {
		t.Fatal("Project ABC should have issue i2")
	}
	if pi2.p != abc {
		t.Error("Issue i2 should point to project ABC")
	}

	def, ok := c.Projects["DEF"]
	if !ok {
		t.Fatal("Project DEF should have been created")
	}
	pi3, ok := def.Issues["i3"]
	if !ok {
		t.Fatal("Project DEF should have issue i3")
	}
	if pi3.p != def {
		t.Error("Issue i3 should point to project DEF")
	}
}

func TestIssueMutation(t *testing.T) {
	l := newLogger()
	c := &Corpus{}

	checkErr(t, l.Log(&devdashpb.Mutation{
		Issue: &devdashpb.IssueMutation{
			Id:       "i1",
			Project:  "ABC",
			IssueKey: "ABC-1",
			Title:    "Setup project",
			Assignees: []*devdashpb.TrackerUser{
				{Id: "ass1", Name: "Assignee 1", Email: "ass1@example.com"},
				{Id: "ass2", Name: "Assignee 2", Email: "ass2@example.com"},
			},
		},
	}))
	checkErr(t, l.Log(&devdashpb.Mutation{
		Issue: &devdashpb.IssueMutation{
			Id:               "i1",
			Project:          "DEF",
			IssueKey:         "DEF-1",
			Title:            "initial project setup",
			Assignees:        []*devdashpb.TrackerUser{{Id: "ass2", Name: "Assignee Two"}},
			DeletedAssignees: []string{"ass1"},
		},
	}))

	l.end()
	checkErr(t, c.Initialize(context.Background(), l))

	i1, ok := c.Issues["i1"]
	if !ok {
		t.Fatal("Issue i1 should exist")
	}
	if i1.Title != "initial project setup" {
		t.Errorf("Issue i1 should have updated title. %s != %s", i1.Title, "initial project setup")
	}
	if i1.p.ID != "DEF" {
		t.Errorf("Project should have been reassigned. %s != %s", i1.p.ID, "DEF")
	}
	if i1.Assignees["ass2"].Name != "Assignee Two" {
		t.Errorf("assignee name should have been updated. %s != %s", i1.Assignees["ass2"].Name, "Assignee Two")
	}
}

func TestMilestoneMutation(t *testing.T) {
	l := newLogger()
	c := &Corpus{}

	checkErr(t, l.Log(&devdashpb.Mutation{
		Project: &devdashpb.ProjectMutation{
			Id: "ABC",
			Milestones: []*devdashpb.TrackerMilestone{
				{Id: "m1", Project: "ABC", Name: "1.0.0", Description: "Release 2019.02"},
				{Id: "m2", Project: "ABC", Name: "1.1.0", Description: "Release 2019.03"},
			},
		},
	}))
	checkErr(t, l.Log(&devdashpb.Mutation{
		Issue: &devdashpb.IssueMutation{
			Id:         "i1",
			Project:    "ABC",
			IssueKey:   "ABC-1",
			Milestones: []*devdashpb.TrackerMilestone{{Id: "m1"}, {Id: "m2"}},
		},
	}))

	l.end()
	checkErr(t, c.Initialize(context.Background(), l))

	i1, ok := c.Issues["i1"]
	if !ok {
		t.Fatal("Issue i1 should exist")
	}

	m1, ok := i1.Milestones["m1"]
	if !ok {
		t.Fatal("Issue i1 should have milestone m1")
	}
	if m1.Name != "1.0.0" {
		t.Errorf("Milestone m1 should have the name, defined with the previous mutation. %s != %s", m1.Name, "1.0.0")
	}
	if m1.Issues["i1"] != i1 {
		t.Error("Milestone m1 should have issue i1")
	}
	if m1.Closed {
		t.Error("Milestone m1 should not be closed")
	}

	m2, ok := i1.Milestones["m2"]
	if !ok {
		t.Fatal("Issue i1 should have milestone m2")
	}
	if m2.Name != "1.1.0" {
		t.Errorf("Milestone m2 should have the name, defined with the previous mutation. %s != %s", m1.Name, "1.1.0")
	}
	if m2.Issues["i1"] != i1 {
		t.Error("Milestone m2 should have issue i1")
	}

	checkErr(t, l.Log(&devdashpb.Mutation{
		Issue: &devdashpb.IssueMutation{
			Id:                "i1",
			Project:           "ABC",
			IssueKey:          "ABC-1",
			DeletedMilestones: []string{"m2"},
		},
	}))
	checkErr(t, l.Log(&devdashpb.Mutation{
		Project: &devdashpb.ProjectMutation{
			Id:         "ABC",
			Milestones: []*devdashpb.TrackerMilestone{{Id: "m1", Closed: pbBool(true)}},
		},
	}))

	l.end()
	checkErr(t, c.Update(context.Background()))

	_, ok = i1.Milestones["m2"]
	if ok {
		t.Error("Milestone m2 should have been removed from issue i1")
	}
	_, ok = m2.Issues["i1"]
	if ok {
		t.Error("Issue i1 should have been removed from milestone m1")
	}

	if !i1.Milestones["m1"].Closed {
		t.Error("Milestone m1 should be closed")
	}

	checkErr(t, l.Log(&devdashpb.Mutation{
		Project: &devdashpb.ProjectMutation{
			Id:                "ABC",
			DeletedMilestones: []string{"m1"},
		},
	}))

	l.end()
	checkErr(t, c.Update(context.Background()))

	_, ok = c.Milestones["m1"]
	if ok {
		t.Error("Milestone m1 should have been removed from corpus")
	}
	_, ok = i1.Milestones["m1"]
	if ok {
		t.Error("Milestone m1 should have been removed from issue i1")
	}

}

func TestReleaseMutation(t *testing.T) {
	l := newLogger()
	c := &Corpus{}

	checkErr(t, l.Log(&devdashpb.Mutation{
		Project: &devdashpb.ProjectMutation{
			Id: "ABC",
			Milestones: []*devdashpb.TrackerMilestone{
				{Id: "m1", Project: "ABC", Name: "1.0.0", Description: "Release 2019.02"},
				{Id: "m2", Project: "ABC", Name: "1.1.0", Description: "Release 2019.03"},
			},
		},
	}))
	checkErr(t, l.Log(&devdashpb.Mutation{
		Project: &devdashpb.ProjectMutation{
			Id: "DEF",
			Milestones: []*devdashpb.TrackerMilestone{
				{Id: "m3", Project: "ABC", Name: "3.0.0", Description: "Release 2019.02"},
				{Id: "m4", Project: "ABC", Name: "3.1.0", Description: "Release 2019.03"},
			},
		},
	}))
	checkErr(t, l.Log(&devdashpb.Mutation{
		Issue: &devdashpb.IssueMutation{
			Id:         "i1",
			Project:    "ABC",
			IssueKey:   "ABC-1",
			Milestones: []*devdashpb.TrackerMilestone{{Id: "m1"}},
		},
	}))
	checkErr(t, l.Log(&devdashpb.Mutation{
		Issue: &devdashpb.IssueMutation{
			Id:         "i2",
			Project:    "DEF",
			IssueKey:   "DEF-1",
			Milestones: []*devdashpb.TrackerMilestone{{Id: "m3"}},
		},
	}))
	checkErr(t, l.Log(&devdashpb.Mutation{
		Release: &devdashpb.ReleaseMutation{
			Id:          "r1",
			Name:        "Release 2019.02",
			ReleaseDate: ptypes.TimestampNow(),
			Milestones:  []*devdashpb.TrackerMilestone{{Id: "m1"}, {Id: "m2"}},
		},
	}))

	l.end()
	checkErr(t, c.Initialize(context.Background(), l))

	r1, ok := c.Releases["r1"]
	if !ok {
		t.Fatal("Release r1 should exist")
	}

	m1, ok := c.Milestones["m1"]
	if !ok {
		t.Fatal("Milestone m1 should exist")
	}
	m2, ok := c.Milestones["m2"]
	if !ok {
		t.Fatal("Milestone m2 should exist")
	}

	if r1.Milestones["m1"] != m1 {
		t.Error("Release r1 should have milestone m1.")
	}

	if r1.Milestones["m2"] != m2 {
		t.Error("Release r1 should have milestone m2.")
	}

}

func checkErr(t *testing.T, err error) {
	if err != nil {
		t.Fatalf("unexpected error: %s", err.Error())
	}
}
