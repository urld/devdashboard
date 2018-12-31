// Copyright 2018 David Url.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/urld/devdashboard"
	"github.com/urld/devdashboard/devdashdata"
	"github.com/urld/devdashboard/devdashpb"
)

var (
	dataPath = flag.String("data", "", "data path ")
	logger   *devdashboard.DiskMutationLogger
)

func main() {
	flag.Parse()
	dir := targetDir()
	os.RemoveAll(dir)
	os.MkdirAll(dir, 0700)

	fmt.Println("logging fixtures to selected dir...", dir)
	logger = devdashboard.NewDiskMutationLogger(dir)

	log(&devdashpb.Mutation{
		Project: &devdashpb.ProjectMutation{
			Id:   "ABC",
			Name: "Alpha Bravo Charlie",
			Milestones: []*devdashpb.TrackerMilestone{
				{
					Id:          "abc201902",
					Project:     "ABC",
					Name:        "v0.1.0",
					Description: "Release 2019.02?",
				},
				{
					Id:          "abc201906",
					Project:     "ABC",
					Name:        "v1.0.0",
					Description: "Release 2019.06",
				},
			},
		},
	})
	log(&devdashpb.Mutation{
		Project: &devdashpb.ProjectMutation{
			Id:   "DEF",
			Name: "Another project",
			Milestones: []*devdashpb.TrackerMilestone{
				{
					Id:          "def201901",
					Project:     "DEF",
					Name:        "v0.1.0",
					Description: "Release 2019.01",
				},
				{
					Id:          "def201902",
					Project:     "DEF",
					Name:        "v0.2.0",
					Description: "Release 2019.02",
				},
				{
					Id:          "def201906",
					Project:     "DEF",
					Name:        "v1.0.0",
					Description: "Release 2019.06",
				},
			},
		},
	})
	log(&devdashpb.Mutation{
		Issue: &devdashpb.IssueMutation{
			Id:        "i1",
			Project:   "ABC",
			IssueKey:  "ABC-1",
			Title:     "Setup project",
			Body:      "* create git repo\n* write readme\n* configure ci build",
			Status:    "Done",
			Closed:    boolChange(true),
			ClosedAt:  pbTimestamp("2018-12-27T04:13"),
			ClosedBy:  &devdashpb.TrackerUser{Id: "urld", Name: "David Url", Email: "david@urld.io"},
			Created:   pbTimestamp("2018-12-10T14:13"),
			Updated:   pbTimestamp("2018-12-24T21:51"),
			Assignees: []*devdashpb.TrackerUser{{Id: "urld", Name: "David Url", Email: "david@urld.io"}},
			Owner:     &devdashpb.TrackerUser{Id: "urld", Name: "David Url", Email: "david@urld.io"},
		},
	})
	log(&devdashpb.Mutation{
		Issue: &devdashpb.IssueMutation{
			Id:       "i2",
			Project:  "ABC",
			IssueKey: "ABC-2",
			Title:    "service specs",
			Body:     "REST service specification",
			Status:   "New",
			Created:  pbTimestamp("2018-12-10T14:13"),
			Updated:  pbTimestamp("2018-12-24T21:51"),
			Owner:    &devdashpb.TrackerUser{Id: "urld", Name: "David Url", Email: "david@urld.io"},
		},
	})
	log(&devdashpb.Mutation{
		Issue: &devdashpb.IssueMutation{
			Id:        "i3",
			Project:   "DEF",
			IssueKey:  "DEF-1",
			Title:     "client prototype",
			Body:      "prototype of a http client for service x",
			Status:    "In Progress",
			Created:   pbTimestamp("2018-12-11T11:13"),
			Updated:   pbTimestamp("2018-12-26T19:21"),
			Assignees: []*devdashpb.TrackerUser{{Id: "urld", Name: "David Url", Email: "david@urld.io"}},
			Owner:     &devdashpb.TrackerUser{Id: "urld", Name: "David Url", Email: "david@urld.io"},
		},
	})
	log(&devdashpb.Mutation{
		Release: &devdashpb.ReleaseMutation{
			Id:          "r1",
			Name:        "Release 2019.01",
			Description: "Initial Release of 2019 (January)",
			FreezeDate:  pbTimestamp("2018-12-26T12:00"),
			ReleaseDate: pbTimestamp("2019-01-10T18:00"),
			Milestones:  []*devdashpb.TrackerMilestone{{Id: "def201901"}},
		},
	})
	log(&devdashpb.Mutation{
		Release: &devdashpb.ReleaseMutation{
			Id:          "r2",
			Name:        "Release 2019.02",
			Description: "Maintenance Release Feb 2019",
			FreezeDate:  pbTimestamp("2019-02-13T18:00"),
			ReleaseDate: pbTimestamp("2019-02-20T18:00"),
			Milestones:  []*devdashpb.TrackerMilestone{{Id: "def201902"}, {Id: "abc201902"}},
		},
	})
	log(&devdashpb.Mutation{
		Release: &devdashpb.ReleaseMutation{
			Id:          "r3",
			Name:        "Release 2019.06",
			Description: "Final Release Jun 2019",
			FreezeDate:  pbTimestamp("2019-06-20T12:00"),
			ReleaseDate: pbTimestamp("2019-06-27T18:00"),
			Milestones:  []*devdashpb.TrackerMilestone{{Id: "def201906"}, {Id: "abc201906"}},
		},
	})
	log(&devdashpb.Mutation{
		Issue: &devdashpb.IssueMutation{
			Id:         "i1",
			Closed:     boolChange(true),
			Milestones: []*devdashpb.TrackerMilestone{{Id: "abc201902"}},
		},
	})
	log(&devdashpb.Mutation{
		Issue: &devdashpb.IssueMutation{
			Id:         "i2",
			Milestones: []*devdashpb.TrackerMilestone{{Id: "abc201902"}},
		},
	})
	log(&devdashpb.Mutation{
		Issue: &devdashpb.IssueMutation{
			Id:         "i3",
			Milestones: []*devdashpb.TrackerMilestone{{Id: "def201901"}},
		},
	})
	log(&devdashpb.Mutation{
		Issue: &devdashpb.IssueMutation{
			Id:         "i4",
			Project:    "DEF",
			IssueKey:   "DEF-2",
			Title:      "Improve client API",
			Body:       "the client api needs some improvement after the initial prototype is hard to use (see DEF-1)",
			Status:     "New",
			Created:    pbTimestamp("2018-12-11T11:13"),
			Updated:    pbTimestamp("2018-12-26T19:21"),
			Owner:      &devdashpb.TrackerUser{Id: "urld", Name: "David Url", Email: "david@urld.io"},
			Milestones: []*devdashpb.TrackerMilestone{{Id: "def201902"}},
		},
	})
	log(&devdashpb.Mutation{
		Issue: &devdashpb.IssueMutation{
			Id:         "i5",
			Project:    "DEF",
			IssueKey:   "DEF-3",
			Title:      "Null pointer if network is down",
			Body:       "add null checks and recover from network connectivity issues",
			Status:     "New",
			Created:    pbTimestamp("2018-12-18T11:02"),
			Updated:    pbTimestamp("2018-12-18T13:14"),
			Owner:      &devdashpb.TrackerUser{Id: "urld", Name: "David Url", Email: "david@urld.io"},
			Milestones: []*devdashpb.TrackerMilestone{{Id: "def201902"}},
		},
	})
}

func pbTimestamp(s string) *timestamp.Timestamp {
	const timefmt = "2006-01-02T15:04"
	t, _ := time.Parse(timefmt, s)
	ts, _ := ptypes.TimestampProto(t)
	return ts
}

func targetDir() string {
	dir := *dataPath
	if dir == "" {
		dir = devdashdata.DefaultDir()
	}
	return dir
}

func log(m *devdashpb.Mutation) {
	err := logger.Log(m)
	if err != nil {
		panic(err)
	}
}

func boolChange(val bool) *devdashpb.BoolChange {
	return &devdashpb.BoolChange{Val: val}
}
