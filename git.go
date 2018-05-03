// Copyright (c) 2018, David Url
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package devdashboard

type Repo struct {
	c *Corpus

	commits map[string]*Commit
	refs    []string

	URL string
}

type Commit struct {
	Id  string
	Msg string
}
