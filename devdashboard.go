// Copyright (c) 2018, David Url
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package devdashboard

type Corpus struct {
	dataDir string

	releases map[string]*Release
	issues   *IssueTracker
}

func NewCorpus(path string) *Corpus {
	return &Corpus{dataDir: path}
}
