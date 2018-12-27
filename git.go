// Copyright (c) 2018, David Url
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package devdashboard

import "github.com/urld/devdashboard/devdashpb"

type GitRepo struct {
	c *Corpus

	URL     string
	commits map[string]*GitCommit
	refs    []GitRef
}

type GitRef struct {
	r *GitRepo

	Ref  string
	Sha1 string
}

type GitCommit struct {
	r *GitRepo

	Sha1     string
	Raw      string
	DiffTree map[string]*GitDiffTreeFile
}

type GitDiffTreeFile struct {
	c *GitCommit

	file    string
	added   int64
	deleted int64
	binary  bool
}

func (c *Corpus) processGitMutation(gm *devdashpb.GitMutation) {
	// TODO
}
