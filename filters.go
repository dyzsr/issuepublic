package main

import (
	"strings"
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
