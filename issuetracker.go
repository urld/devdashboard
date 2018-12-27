// Copyright 2018 David Url.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package devdashboard

import (
	"log"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/urld/devdashboard/devdashpb"
)

type Project struct {
	c *Corpus

	ID          string
	Name        string
	Description string

	Issues     map[string]*Issue
	Milestones map[string]*Milestone
}

type Release struct {
	c *Corpus

	ID          string
	Name        string
	Description string

	FreezeDate  time.Time
	ReleaseDate time.Time
	Closed      bool

	Milestones map[string]*Milestone
}

func (r *Release) IsFrozen() bool {
	t := time.Now()
	return t.After(r.FreezeDate)
}

func (r *Release) IsReleased() bool {
	t := time.Now()
	return t.After(r.ReleaseDate)
}

func (c *Corpus) getOrCreateProject(id string) *Project {
	p, ok := c.Projects[id]
	if !ok {
		// new project
		p = &Project{
			ID:         id,
			Issues:     make(map[string]*Issue),
			Milestones: make(map[string]*Milestone),
		}
	}
	c.Projects[id] = p
	return p

}

type Milestone struct {
	p *Project

	ID string

	Name        string
	Description string
	Closed      bool

	Issues map[string]*Issue
}

func (m *Milestone) Project() *Project {
	return m.p
}

type Issue struct {
	p *Project

	ID       string
	IssueKey string

	Created time.Time
	Updated time.Time

	Title string
	Body  string

	Owner     *IssueTrackerUser
	Assignees map[string]*IssueTrackerUser

	Milestones map[string]*Milestone

	Status string
	Closed bool

	ClosedAt time.Time
	ClosedBy *IssueTrackerUser

	Labels  map[string]bool
	Commits map[string]*GitCommit

	//TODO:
	URL string
}

type IssueTrackerUser struct {
	ID    string
	Name  string
	Email string
}

func (c *Corpus) processProjectMutation(pm *devdashpb.ProjectMutation) {
	p := c.getOrCreateProject(pm.Id)
	if pm.Name != "" {
		p.Name = pm.Name
	}
	if pm.Description != "" {
		p.Description = pm.Description
	}
	for _, mm := range pm.Milestones {
		c.processMilestoneMutation(mm)
	}
	for _, id := range pm.DeletedMilestones {
		m, ok := p.Milestones[id]
		if ok {
			for _, i := range m.Issues {
				delete(i.Milestones, id)
			}
		}
		delete(p.Milestones, id)
		delete(c.Milestones, id)
	}
}

func (c *Corpus) processReleaseMutation(rm *devdashpb.ReleaseMutation) {
	r, ok := c.Releases[rm.Id]
	if !ok {
		// new release
		r = &Release{
			ID: rm.Id,
		}
		c.Releases[rm.Id] = r
	}
	if rm.Name != "" {
		r.Name = rm.Name
	}
	if rm.Description != "" {
		r.Description = rm.Description
	}
	if rm.FreezeDate != nil {
		r.FreezeDate = pbTime(rm.FreezeDate)
	}
	if rm.ReleaseDate != nil {
		r.ReleaseDate = pbTime(rm.ReleaseDate)
	}
	r.Closed = rm.Closed
	for _, mm := range rm.Milestones {
		m := c.processMilestoneMutation(mm)
		if r.Milestones == nil {
			r.Milestones = make(map[string]*Milestone)
		}
		r.Milestones[mm.Id] = m
	}
	for _, id := range rm.DeletedMilestones {
		delete(r.Milestones, id)
	}
}

func (c *Corpus) processIssueMutation(im *devdashpb.IssueMutation) {
	i, ok := c.Issues[im.Id]
	if !ok {
		// new issue
		i = &Issue{
			ID:    im.Id,
			Owner: c.processTrackerUserMutation(im.Owner),
		}
		c.Issues[im.Id] = i
	}
	// update issue
	if im.Project != "" {
		i.p = c.getOrCreateProject(im.Project)
		i.p.Issues[i.ID] = i
	}
	if im.IssueKey != "" {
		i.IssueKey = im.IssueKey
	}
	if im.Created != nil {
		i.Created = pbTime(im.Created)
	}
	if im.Updated != nil {
		i.Updated = pbTime(im.Updated)
	}
	if im.Title != "" {
		i.Title = im.Title
	}
	if im.Body != "" {
		i.Body = im.Body
	}
	for _, um := range im.Assignees {
		u := c.processTrackerUserMutation(um)
		if i.Assignees == nil {
			i.Assignees = make(map[string]*IssueTrackerUser, len(im.Assignees))
		}
		i.Assignees[u.ID] = u
	}
	for _, id := range im.DeletedAssignees {
		delete(i.Assignees, id)
	}
	for _, mm := range im.Milestones {
		m := c.processMilestoneMutation(mm)
		if m.Issues == nil {
			m.Issues = make(map[string]*Issue, len(im.Milestones))
		}
		m.Issues[i.ID] = i
		if i.Milestones == nil {
			i.Milestones = make(map[string]*Milestone, len(im.Milestones))
		}
		i.Milestones[m.ID] = m
	}
	for _, id := range im.DeletedMilestones {
		m, ok := i.Milestones[id]
		if ok {
			delete(m.Issues, i.ID)
		}
		delete(i.Milestones, id)
	}
	if im.Status != "" {
		i.Status = im.Status
	}
	i.Closed = im.Closed
	if im.ClosedAt != nil {
		i.ClosedAt = pbTime(im.ClosedAt)
	}
	if im.ClosedBy != nil {
		i.ClosedBy = c.processTrackerUserMutation(im.ClosedBy)
	}
	for _, l := range im.Labels {
		if i.Labels == nil {
			i.Labels = make(map[string]bool)
		}
		i.Labels[l.Name] = true
	}
	for _, l := range im.DeletedLabels {
		delete(i.Labels, l.Name)
	}
}

func (c *Corpus) processTrackerUserMutation(um *devdashpb.TrackerUser) *IssueTrackerUser {
	if um == nil {
		return nil
	}
	u, ok := c.TrackerUsers[um.Id]
	if !ok {
		// new user
		u = &IssueTrackerUser{
			ID: um.Id,
		}
		c.TrackerUsers[um.Id] = u
	}
	// update user
	if um.Name != "" {
		u.Name = um.Name
	}
	if um.Email != "" {
		u.Email = um.Email
	}
	return u
}

func (c *Corpus) processMilestoneMutation(mm *devdashpb.TrackerMilestone) *Milestone {
	if mm == nil {
		return nil
	}
	m, ok := c.Milestones[mm.Id]
	if !ok {
		// new milestone
		m = &Milestone{
			ID: mm.Id,
		}
		c.Milestones[mm.Id] = m
	}
	// update milestone
	if mm.Project != "" {
		m.p = c.Projects[mm.Project]
		m.p.Milestones[m.ID] = m
	}
	if mm.Name != "" {
		m.Name = mm.Name
	}
	if mm.Description != "" {
		m.Description = mm.Description
	}
	m.Closed = mm.Closed
	return m
}

func pbTime(ts *timestamp.Timestamp) time.Time {
	t, err := ptypes.Timestamp(ts)
	if err != nil {
		log.Printf("could not convert protobuf timestamp:  %v", err)
	}
	return t
}
