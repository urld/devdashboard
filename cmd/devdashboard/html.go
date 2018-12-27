// Copyright 2018 David Url.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"errors"
	"html/template"
	"io"
	"log"
	"path/filepath"
	"sync"
	"time"

	humanize "github.com/dustin/go-humanize"
	"github.com/fsnotify/fsnotify"
)

var (
	tmplMap  map[string]*template.Template
	tmplLock sync.RWMutex
)

func initTemplates(basePath string, failOnErr bool) {
	if tmplMap == nil {
		tmplMap = make(map[string]*template.Template)
	}

	rootTmpl := filepath.Join(basePath, "templates/root.tmpl")

	for name, contentTmpl := range map[string]string{
		"release": "release.tmpl",
	} {
		contentTmpl = filepath.Join(basePath, "templates", contentTmpl)

		tmpl := template.New("root")
		tmpl.Funcs(template.FuncMap{
			"fmtDate":     fmtDate,
			"fmtDateTime": fmtDateTime,
			"fmtRelTime":  fmtRelTime,
		})
		tmpl, err := tmpl.ParseFiles(rootTmpl, contentTmpl)
		if err != nil {
			if failOnErr {
				log.Fatal(err)
			} else {
				log.Println(err)
				continue
			}
		}
		tmplLock.Lock()
		tmplMap[name] = tmpl
		tmplLock.Unlock()
	}
}

func watchTemplates(basePath string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	tmplPath := filepath.Join(basePath, "templates")
	err = watcher.Add(tmplPath)
	if err != nil {
		log.Fatal(err)
	}

	for {
		select {
		case event := <-watcher.Events:
			if event.Op&fsnotify.Write == fsnotify.Write {
				log.Println("detected change in templates. reloading...")
				initTemplates(basePath, false)
			}
		case err := <-watcher.Errors:
			log.Println(err)
		}
	}
}

func renderHTML(w io.Writer, name string, data interface{}) error {
	tmplLock.RLock()
	tmpl, ok := tmplMap[name]
	tmplLock.RUnlock()
	if !ok {
		return errors.New("could not find template for " + name)
	}
	return tmpl.ExecuteTemplate(w, "root", data)
}

func fmtDate(t time.Time) string {
	return t.Format("2006-01-02")
}

func fmtDateTime(t time.Time) string {
	//return t.Format("2006-01-02 15:04")
	return t.Format(time.RFC1123Z)
}

func fmtRelTime(t time.Time) string {
	return humanize.Time(t)
}
