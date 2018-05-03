// Copyright (c) 2018, David Url
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package devdashboard

import "time"

type Release struct {
	Name        string
	FreezeDate  time.Time
	ReleaseDate time.Time
}

func (r *Release) Components() []*ComponentVersion {
	// TODO
	c1 := ComponentVersion{ComponentId: "FOO", ComponentName: "Foo Library", Version: "4.8.0", ComponentURL: "https://example.com/foolib"}
	c2 := ComponentVersion{ComponentId: "BAR", ComponentName: "Bar Component", Version: "1.7.3", ComponentURL: "https://example.com/barcomp"}
	return []*ComponentVersion{&c1, &c2}
}

func (r *Release) IsFrozen() bool {
	t := time.Now()
	return t.After(r.FreezeDate)
}

func (r *Release) IsReleased() bool {
	t := time.Now()
	return t.After(r.ReleaseDate)
}

type ComponentVersion struct {
	ComponentId   string
	ComponentName string
	Version       string

	ComponentURL string
	VersionURL   string
}

func (cv *ComponentVersion) Issues() []*Issue {
	// TODO
	i1 := Issue{Id: "ABC-42", Title: "fatal crash due to overheating", Kind: Bug, State: Open, URL: "https://example.com/issues/ABC-42"}
	i2 := Issue{Id: "ABC-39", Title: "add important hello button", Kind: Story, State: Closed, URL: "https://example.com/issues/ABC-39"}
	return []*Issue{&i1, &i2}
}
