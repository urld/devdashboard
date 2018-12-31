// Copyright 2018 David Url.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package devdashboard

import (
	"time"

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

	Labels  map[string]struct{}
	Commits map[string]*GitCommit

	URL string
}

type IssueTrackerUser struct {
	ID    string
	Name  string
	Email string
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
	if rm.Closed != nil {
		r.Closed = rm.Closed.Val
	}
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
	if im.Closed != nil {
		i.Closed = im.Closed.Val
	}
	if im.ClosedAt != nil {
		i.ClosedAt = pbTime(im.ClosedAt)
	}
	if im.ClosedBy != nil {
		i.ClosedBy = c.processTrackerUserMutation(im.ClosedBy)
	}
	for _, l := range im.Labels {
		if i.Labels == nil {
			i.Labels = make(map[string]struct{})
		}
		i.Labels[l.Name] = struct{}{}
	}
	for _, l := range im.DeletedLabels {
		delete(i.Labels, l)
	}
	if im.Url != "" {
		i.URL = im.Url
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
	if mm.Closed != nil {
		m.Closed = mm.Closed.Val
	}
	return m
}

var emptyProject = &Project{}

func (a *Project) GenMutationDiff(b *Project) *devdashpb.ProjectMutation {
	var ret *devdashpb.ProjectMutation // lazily initialized by diff
	diff := func() *devdashpb.ProjectMutation {
		if ret == nil {
			ret = &devdashpb.ProjectMutation{Id: b.ID}
		}
		return ret
	}
	if a == nil {
		a = emptyProject
	}
	if a.Name != b.Name {
		diff().Name = b.Name
	}
	if a.Description != b.Description {
		diff().Description = b.Description
	}
	milestones, deletedMilestones := genMilestoneDiffs(a.Milestones, b.Milestones)
	diff().Milestones = milestones
	diff().DeletedMilestones = deletedMilestones
	return ret
}

var emptyRelease = &Release{}

func (a *Release) GenMutationDiff(b *Release) *devdashpb.ReleaseMutation {
	var ret *devdashpb.ReleaseMutation
	diff := func() *devdashpb.ReleaseMutation {
		if ret == nil {
			ret = &devdashpb.ReleaseMutation{Id: b.ID}
		}
		return ret
	}
	if a == nil {
		a = emptyRelease
	}
	if a.Name != b.Name {
		diff().Name = b.Name
	}
	if a.Description != b.Description {
		diff().Description = b.Description
	}
	if a.FreezeDate != b.FreezeDate {
		diff().FreezeDate = pbTimestamp(b.FreezeDate)
	}
	if a.ReleaseDate != b.ReleaseDate {
		diff().ReleaseDate = pbTimestamp(b.ReleaseDate)
	}
	milestones, deletedMilestones := genMilestoneDiffs(a.Milestones, b.Milestones)
	diff().Milestones = milestones
	diff().DeletedMilestones = deletedMilestones
	if a.Closed != b.Closed {
		diff().Closed = pbBool(b.Closed)
	}
	return ret
}

func genMilestoneDiffs(a, b map[string]*Milestone) (milestones []*devdashpb.TrackerMilestone, deletedMilestones []string) {
	processed := newSet()
	for id, ma := range a {
		mb, ok := b[id]
		if ok {
			milestoneDiff := ma.GenMutationDiff(mb)
			if milestoneDiff != nil {
				milestones = append(milestones, milestoneDiff)
			}
		} else {
			deletedMilestones = append(deletedMilestones, id)
		}
		processed.put(id)
	}
	for id, mb := range b {
		if processed.has(id) {
			continue
		}
		var ma Milestone
		milestoneDiff := ma.GenMutationDiff(mb)
		milestones = append(milestones, milestoneDiff)
	}
	return
}

var emptyMilestone = &Milestone{p: &Project{}}

func (a *Milestone) GenMutationDiff(b *Milestone) *devdashpb.TrackerMilestone {
	var ret *devdashpb.TrackerMilestone // lazily initialized by diff
	diff := func() *devdashpb.TrackerMilestone {
		if ret == nil {
			ret = &devdashpb.TrackerMilestone{Id: b.ID}
		}
		return ret
	}
	if a == nil {
		a = emptyMilestone
	}
	if a.Name != b.Name {
		diff().Name = b.Name
	}
	if a.Description != b.Description {
		diff().Description = b.Description
	}
	if a.p.ID != b.p.ID {
		diff().Project = b.p.ID
	}
	if a.Closed != b.Closed {
		diff().Closed = pbBool(b.Closed)
	}

	return ret
}

func genIssueDiffs(a, b map[string]*Issue) (issues []*devdashpb.IssueMutation, deletedIssues []string) {
	processed := newSet()
	for id, ia := range a {
		ib, ok := b[id]
		if ok {
			issueDiff := ia.GenMutationDiff(ib)
			if issueDiff != nil {
				issues = append(issues, issueDiff)
			}
		} else {
			deletedIssues = append(deletedIssues, id)
		}
		processed.put(id)
	}
	for id, ib := range b {
		if processed.has(id) {
			continue
		}
		var ia *Issue
		issueDiff := ia.GenMutationDiff(ib)
		issues = append(issues, issueDiff)
	}
	return
}

var emptyIssue = &Issue{}

func (a *Issue) GenMutationDiff(b *Issue) *devdashpb.IssueMutation {
	var ret *devdashpb.IssueMutation
	diff := func() *devdashpb.IssueMutation {
		if ret == nil {
			ret = &devdashpb.IssueMutation{Id: b.ID}
		}
		return ret
	}
	if a == nil {
		a = emptyIssue
	}
	if a.p.ID != b.p.ID {
		diff().Project = b.p.ID
	}
	if a.Created != b.Created {
		diff().Created = pbTimestamp(b.Created)
	}
	if a.Updated != b.Updated {
		diff().Updated = pbTimestamp(b.Updated)
	}
	if a.IssueKey != b.IssueKey {
		diff().IssueKey = b.IssueKey
	}
	if a.Title != b.Title {
		diff().Title = b.Title
	}
	if a.Body != b.Body {
		diff().Body = b.Body
	}
	if a.Owner != b.Owner {
		diff().Owner = a.Owner.GenMutationDiff(b.Owner)

	}
	if a.Status != b.Status {
		diff().Status = b.Status
	}
	if a.Closed != b.Closed {
		diff().Closed = pbBool(b.Closed)
	}
	if a.ClosedAt != b.ClosedAt {
		diff().ClosedAt = pbTimestamp(b.ClosedAt)
	}
	if *a.ClosedBy != *b.ClosedBy {
		diff().ClosedBy = a.ClosedBy.GenMutationDiff(b.ClosedBy)
	}
	if a.URL != b.URL {
		diff().Url = b.URL
	}

	assignees, deletedAssignees := genTrackerUserDiffs(a.Assignees, b.Assignees)
	diff().Assignees = assignees
	diff().DeletedAssignees = deletedAssignees

	milestones, deletedMilestones := genMilestoneDiffs(a.Milestones, b.Milestones)
	diff().Milestones = milestones
	diff().DeletedMilestones = deletedMilestones

	labels, deletedLabels := genTrackerLabelDiffs(a.Labels, b.Labels)
	diff().Labels = labels
	diff().DeletedLabels = deletedLabels

	// TODO commits

	return ret
}

var emptyIssueTrackerUser = &IssueTrackerUser{}

func (a *IssueTrackerUser) GenMutationDiff(b *IssueTrackerUser) *devdashpb.TrackerUser {
	var ret *devdashpb.TrackerUser
	diff := func() *devdashpb.TrackerUser {
		if ret == nil {
			ret = &devdashpb.TrackerUser{Id: b.ID}
		}
		return ret
	}
	if a == nil {
		a = emptyIssueTrackerUser
	}
	if a.Name != b.Name {
		diff().Name = b.Name
	}
	if a.Email != b.Email {
		diff().Email = b.Email
	}
	return ret
}

func genTrackerUserDiffs(a, b map[string]*IssueTrackerUser) (users []*devdashpb.TrackerUser, deletedUsers []string) {
	processed := newSet()
	for id, ua := range a {
		ub, ok := b[id]
		if ok {
			userDiff := ua.GenMutationDiff(ub)
			if userDiff != nil {
				users = append(users, userDiff)
			}
		} else {
			deletedUsers = append(deletedUsers, id)
		}
		processed.put(id)
	}
	for id, ub := range b {
		if processed.has(id) {
			continue
		}
		var ua IssueTrackerUser
		userDiff := ua.GenMutationDiff(ub)
		users = append(users, userDiff)
	}
	return
}

func genTrackerLabelDiffs(a, b map[string]struct{}) (labels []*devdashpb.TrackerLabel, deletedLabels []string) {
	processed := newSet()
	for id := range a {
		_, ok := b[id]
		if !ok {
			deletedLabels = append(deletedLabels, id)
		}
		processed.put(id)
	}
	for id := range b {
		if processed.has(id) {
			continue
		}
		labels = append(labels, &devdashpb.TrackerLabel{Name: id})
	}
	return
}
