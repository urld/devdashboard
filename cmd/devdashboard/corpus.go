// Copyright 2018 David Url.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/urld/devdashboard"
	"github.com/urld/devdashboard/devdashdata"
)

var (
	corpus *devdashboard.Corpus
)

func initCorpus() {
	targetDir := *dataPath
	if targetDir == "" {
		targetDir = devdashdata.DefaultDir()
	}
	log.Printf("initializing corpus from %s...", targetDir)
	c, err := devdashdata.Get(context.Background(), targetDir)
	if err != nil {
		log.Fatalf("unable to initialize corpus: %v", err)
	}
	corpus = c
}

func checkReady(w http.ResponseWriter) bool {
	if corpus == nil {
		serviceUnavailable(w, errors.New("devdashboards corpus is still initializing..."))
		return false
	}
	return true
}

func releaseHandler(w http.ResponseWriter, r *http.Request) {
	if !checkReady(w) {
		return
	}
	name := strings.TrimPrefix(r.URL.Path, "/release/")

	corpus.RLock()
	defer corpus.RUnlock()

	data := make([]*devdashboard.Release, 0)
	for _, release := range corpus.Releases {
		if name == "" || release.Name == name {
			data = append(data, release)
		}
	}

	err := renderHTML(w, "release", data)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}
