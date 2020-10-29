package main

import (
	"log"
	"strings"

	"github.com/google/go-github/v32/github"
	"github.com/pkg/errors"
)

type filterOptions struct {
	withLabel []string
	orLabel   []string
	noLabel   []string
	assignees []string
	isOpen    *bool
	isPR      *bool
	linkedPR  *bool
}

var defaultFilter = filterOptions{
	isOpen: fromBool(true),
	isPR:   fromBool(false),
}

func (f *filterOptions) queryString() string {
	var qs []string
	qs = append(qs, "repo:"+owner+"/"+repo)
	for _, lb := range f.withLabel {
		qs = append(qs, "label:"+lb)
	}
	for _, lb := range f.noLabel {
		qs = append(qs, "-label:"+lb)
	}
	if f.assignees == nil {
		qs = append(qs, "no:assignee")
	} else {
		for _, a := range f.assignees {
			qs = append(qs, "assignee:"+a)
		}
	}
	if f.isOpen != nil {
		if *f.isOpen {
			qs = append(qs, "is:open")
		} else {
			qs = append(qs, "is:close")
		}
	}
	if f.isPR != nil {
		if *f.isPR {
			qs = append(qs, "is:pr")
		} else {
			qs = append(qs, "is:issue")
		}
	}
	if f.linkedPR != nil {
		if *f.linkedPR {
			qs = append(qs, "linked:pr")
		} else {
			qs = append(qs, "-linked:pr")
		}
	}
	q := strings.Join(qs, " ")
	return q
}

func getIssuesByFilter(owner string, repo string, filters *filterOptions) ([]*github.Issue, error) {
	log.Printf("start getIssuesByFilter")
	defer log.Printf("end getIssuesByFilter")

	if filters == nil {
		filters = &filterOptions{}
	}
	query := filters.queryString()

	var allIssues []*github.Issue
	nextPage := 1
	for nextPage > 0 {
		searchOpts := &github.SearchOptions{
			Sort: "created",
		}
		searchOpts.Page = nextPage

		issuesResult, resp, err := cli.Search.Issues(ctx, query, searchOpts)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		issues := issuesResult.Issues

		log.Printf("got: %v\n", issueNumbers(issues))
		allIssues = append(allIssues, issues...)
		nextPage = resp.NextPage
	}

	var ret []*github.Issue
	for _, issue := range allIssues {
		var include bool
		for _, lb := range filters.orLabel {
			for _, label := range issue.Labels {
				if lb == *label.Name {
					include = true
					break
				}
			}
		}
		if len(filters.orLabel) == 0 {
			include = true
		}

		if include {
			ret = append(ret, issue)
		}
		if limit > 0 && len(ret) >= limit {
			break
		}
	}

	return ret, nil
}
