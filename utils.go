package main

import (
	"fmt"

	"github.com/google/go-github/v32/github"
)

func fromBool(v bool) *bool {
	return &v
}

func fromString(v string) *string {
	return &v
}

func printIssue(issue *github.Issue) {
	typ := "issue"
	if issue.IsPullRequest() {
		typ = "pull"
	}
	fmt.Printf("#%d, %s, title: %s\n", *issue.Number, typ, *issue.Title)
	fmt.Printf("\tstate: %s\n", *issue.State)
	fmt.Printf("\turl: %s\n", *issue.HTMLURL)
	fmt.Printf("\tlabels: %s\n", labelNames(issue.Labels))
}

func issueNumbers(issues []*github.Issue) []int {
	ret := make([]int, 0, len(issues))
	for _, issue := range issues {
		ret = append(ret, *issue.Number)
	}
	return ret
}

func labelNames(labels []*github.Label) []string {
	ret := make([]string, 0, len(labels))
	for _, label := range labels {
		ret = append(ret, *label.Name)
	}
	return ret
}

func appendLabels(labels []*string, newLabels ...string) []*string {
	for _, label := range newLabels {
		labels = append(labels, &label)
	}
	return labels
}

func findLabel(labels []string, target string) bool {
	for _, label := range labels {
		if target == label {
			return true
		}
	}
	return false
}
