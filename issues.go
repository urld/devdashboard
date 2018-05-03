// Copyright (c) 2018, David Url
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package devdashboard

type IssueTracker struct {
	c *Corpus

	projects map[string]*Project
	labels   []string
}

type Project struct {
	it *IssueTracker

	id       string
	issues   map[string]*Issue
	versions map[string]*Version
}

type Version struct {
	p *Project

	ProjectId string
	Name      string
	URL       string
}

type Issue struct {
	p *Project

	Id    string
	Title string
	Kind  IssueKind
	State IssueState
	URL   string
}

type IssueKind string

const (
	Story IssueKind = "story"
	Bug   IssueKind = "bug"
)

type IssueState string

const (
	Open   IssueState = "open"
	Closed IssueState = "closed"
)

func (i *Issue) IsOpen() bool {
	return i.State == Open
}

func (i *Issue) IsClosed() bool {
	return i.State == Closed
}

func (i *Issue) Commits() []*Commit {
	return []*Commit{}
}

func (i *Issue) HasUnmergedCommits() bool {
	return true
}
