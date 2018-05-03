// Copyright (c) 2018, David Url
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"go/build"
	"log"
	"net/http"
	"os"

	"github.com/urld/devdashboard"
)

const basePkg = "github.com/urld/devdashboard/cmd/devdashboard"

var (
	httpAddr = flag.String("http", "127.0.0.1:8080", "HTTP Service address (e.g., '127.0.0.1:8080')")
	basePath = flag.String("base", "", "base path for html templates and static resources")

	corpus *devdashboard.Corpus
)

func main() {
	flag.Parse()

	if *basePath == "" {
		p, err := build.Default.Import(basePkg, "", build.FindOnly)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Couldn't locate template files: %v\n", err)
			fmt.Fprintf(os.Stderr, basePathMessage, basePkg)
			os.Exit(1)
		}
		*basePath = p.Dir
	}

	initTemplates(*basePath, true)
	go watchTemplates(*basePath)

	http.HandleFunc("/static/", fileServer(*basePath))
	http.HandleFunc("/release/", releaseHandler)

	log.Fatal(http.ListenAndServe(*httpAddr, nil))
}

const basePathMessage = `
By default, devdashboard locates the html template files and associated
static content by looking for a %q package
in your Go workspaces (GOPATH).

You may use the -base flag to specify an alternate location.
`
