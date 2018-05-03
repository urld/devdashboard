// Copyright (c) 2018, David Url
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/urld/devdashboard"
)

func releaseHandler(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/release/")

	data := release(name)

	err := renderHtml(w, "release", &data)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}

func release(name string) devdashboard.Release {
	return devdashboard.Release{
		Name:        name,
		FreezeDate:  time.Date(2018, 5, 15, 18, 0, 0, 0, time.UTC),
		ReleaseDate: time.Date(2018, 5, 23, 23, 30, 0, 0, time.UTC),
	}
}
