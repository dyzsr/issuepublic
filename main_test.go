package main

import (
	"flag"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/google/go-github/v32/github"
)

var (
	writeDesc bool
)

func TestCreateIssues(t *testing.T) {
	title := "Issue on " + time.Now().String()
	var body string
	if writeDesc {
		body = `
## Description
...

## Bug Report

These sql results are inconsistent with MySQL

### 1. Minimal reproduce step (Required)
...

### 2. What did you expect to see? (Required)
...

### 3. What did you see instead (Required)
...

### 4. What is your TiDB version? (Required)
...
	`
	}
	req := &github.IssueRequest{
		Title: &title,
		Body:  &body,
	}
	issue, _, err := cli.Issues.Create(ctx, "dyzsr", "issuepublic", req)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("#%d, title:\"%s\"\n", *issue.Number, *issue.Title)
}

func TestAddLabels(t *testing.T) {
	editOpt := &editOption{
		labelOption: labelOption{},
		editIssue: func(issue *github.Issue) error {
			var lb string
			switch rand.Intn(5) {
			case 0:
				lb = "type/bug"
			case 1:
				lb = "high-performance"
			case 2:
				lb = "challenge-program"
			case 3:
				lb = "sig/execution"
			case 4:
				lb = "sig/planner"
			}
			_, _, err := cli.Issues.AddLabelsToIssue(ctx, owner, repo, *issue.Number, []string{lb})
			if err != nil {
				return err
			}
			return nil
		},
	}
	err := editIssues(editOpt)
	if err != nil {
		t.Error(err)
	}
}

func TestRemoveLabels(t *testing.T) {
	editOpt := &editOption{
		labelOption: labelOption{},
		editIssue: func(issue *github.Issue) error {
			_, err := cli.Issues.RemoveLabelsForIssue(ctx, owner, repo, *issue.Number)
			return err
		},
	}
	err := editIssues(editOpt)
	if err != nil {
		t.Error(err)
	}
}

func TestCloseIssues(t *testing.T) {
	editOpt := &editOption{
		labelOption: labelOption{},
		editIssue: func(issue *github.Issue) error {
			state := "close"
			_, _, err := cli.Issues.Edit(ctx, owner, repo, *issue.Number, &github.IssueRequest{State: &state})
			return err
		},
	}
	err := editIssues(editOpt)
	if err != nil {
		t.Error(err)
	}
}

func TestMain(m *testing.M) {
	flag.BoolVar(&writeDesc, "desc", false, "Write descriptions for issue?")
	flag.Parse()
	initCli()
	m.Run()
}
