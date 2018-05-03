// Copyright (c) 2018, David Url
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"net/http"

	"github.com/xi2/httpgzip"
)

func fileServer(root string) func(http.ResponseWriter, *http.Request) {
	fs := httpgzip.NewHandler(http.FileServer(http.Dir(root)), nil)
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=7776000")
		fs.ServeHTTP(w, r)
	}
}
