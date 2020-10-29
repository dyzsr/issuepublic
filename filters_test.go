package main

import "testing"

func TestFilterOptions(t *testing.T) {
	isPR := false
	linkedPR := false
	filter := &filterOptions{
		withLabel: include,
		orLabel:   sigLabels,
		noLabel:   exclude,
		isPR:      &isPR,
		linkedPR:  &linkedPR,
	}
	t.Log(filter.queryString())
}
