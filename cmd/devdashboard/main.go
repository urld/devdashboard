// Copyright 2018 David Url.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"go/build"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"

	"github.com/bradleyjkemp/memviz"
)

const basePkg = "github.com/urld/devdashboard/cmd/devdashboard"

const basePathMessage = `
By default, devdashboard locates the html template files and associated
static content by looking for a %q package
in your Go workspaces (GOPATH).

You may use the -base flag to specify an alternate location.
`

var (
	httpAddr = flag.String("http", "127.0.0.1:8080", "HTTP Service address (e.g., '127.0.0.1:8080')")
	basePath = flag.String("base", "", "base path for html templates and static resources")
	dataPath = flag.String("data", "", "data path ")
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

	initServer()
	go initCorpus()

	log.Fatal(http.ListenAndServe(*httpAddr, nil))
}

func initServer() {
	initTemplates(*basePath, true)
	go watchTemplates(*basePath)

	http.HandleFunc("/static/", fileServer(*basePath))
	http.HandleFunc("/release/", releaseHandler)
	http.HandleFunc("/corpusviz/", corpusvizHandler)
}

func fileServer(root string) func(http.ResponseWriter, *http.Request) {
	fs := http.FileServer(http.Dir(root))
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=7776000")
		fs.ServeHTTP(w, r)
	}
}

func corpusvizHandler(w http.ResponseWriter, r *http.Request) {
	//TODO: not usefull for real data sets. needs to be removed soon
	cmd := exec.Command("dot", "-Tsvg")
	in, err := cmd.StdinPipe()
	internalServerError(w, err)
	defer in.Close()

	cmd.Stderr = os.Stderr
	out, err := cmd.StdoutPipe()
	internalServerError(w, err)

	err = cmd.Start()
	internalServerError(w, err)

	memviz.Map(in, corpus)
	in.Close()
	_, err = io.Copy(w, out)
	internalServerError(w, err)

	err = cmd.Wait()
	internalServerError(w, err)

}

func internalServerError(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}
	log.Println(err)
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

func serviceUnavailable(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}
	log.Println(err)
	http.Error(w, err.Error(), http.StatusServiceUnavailable)
}
