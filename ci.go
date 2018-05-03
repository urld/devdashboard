// Copyright (c) 2018, David Url
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package devdashboard

type CiBuild struct {
	Id     string
	Status CiStatus

	URL string
}

type CiStatus int

const (
	Success CiStatus = iota
	Failed
	Unknown
)
