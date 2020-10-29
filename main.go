package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"regexp"
	"time"

	"github.com/google/go-github/v32/github"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
)

var (
	inspect bool
	token   string
	owner   string
	repo    string
	mentor  string
	limit   int

	ctx = context.Background()
	cli *github.Client
)

func init() {
	rand.Seed(time.Now().Unix())
	flag.BoolVar(&inspect, "inspect", false, "Inspect issues")
	flag.StringVar(&token, "token", "", "GitHub personal access token")
	flag.StringVar(&owner, "owner", "dyzsr", "Owner of repo")
	flag.StringVar(&repo, "repo", "issuepublic", "Repo name")
	flag.StringVar(&mentor, "mentor", "lzmhhh123", "Mentor of the issue")
	flag.IntVar(&limit, "limit", 10, "Number of issues to be processed")
}

func initCli() {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	cli = github.NewClient(tc)
}

func editDesc(desc string, slackChannel string) (string, error) {
	matched, err := regexp.MatchString("(?m)^## Description", desc)
	if err != nil {
		return "", errors.WithStack(err)
	}
	ret := desc
	if !matched {
		ret = "## Description\n" + ret
	}
	ret += fmt.Sprintf(`
## SIG slack channel

 %s

## Score

300

## Mentor

- @%s
`, slackChannel, mentor)

	return ret, nil
}

var (
	include     = []string{"type/bug"}
	exclude     = []string{"high-performance", "challenge-program"}
	addToIssues = []string{"challenge-program", "status/help-wanted"}
	sigLabels   = []string{"sig/execution", "sig/planner"}
	sigs        = map[string]string{
		"sig/execution": "[#sig-exec](https://slack.tidb.io/invite?team=tidb-community&channel=sig-exec&ref=high-performance)",
		"sig/planner":   "[#sig-planner](https://slack.tidb.io/invite?team=tidb-community&channel=sig-planner&ref=high-performance)",
	}
)

func defaultEditIssue(issue *github.Issue) (err error) {
	var desc string
	if issue.Body != nil {
		desc = *issue.Body
	}

	newDesc := desc
	labels := labelNames(issue.Labels)
	for label, sig := range sigs {
		if findLabel(labels, label) {
			newDesc, err = editDesc(desc, sig)
			if err != nil {
				return errors.WithStack(err)
			}
			break
		}
	}

	labels = append(labels, addToIssues...)

	cli.Issues.Edit(ctx, owner, repo, *issue.Number, &github.IssueRequest{
		Title:  issue.Title,
		Body:   &newDesc,
		Labels: &labels,
	})
	return nil
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

		for _, issue := range issues {
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
				allIssues = append(allIssues, issue)
			}
			if limit > 0 && len(allIssues) >= limit {
				break
			}
		}
		if limit > 0 && len(allIssues) >= limit {
			break
		}

		nextPage = resp.NextPage
	}

	return allIssues, nil
}

type editOption struct {
	filterOptions
	editIssue func(*github.Issue) error
}

func editIssues(opt *editOption) error {
	log.Printf("start editIssues")
	defer log.Printf("end editIssues")
	issues, err := getIssuesByFilter(owner, repo, &opt.filterOptions)
	if err != nil {
		return errors.WithStack(err)
	}

	log.Printf("actually got: %v", issueNumbers(issues))

	for _, issue := range issues {
		printIssue(issue)
		if inspect {
			continue
		}
		if err := opt.editIssue(issue); err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

func main() {
	flag.Parse()
	initCli()

	editOpt := &editOption{
		filterOptions: filterOptions{
			withLabel: include,
			orLabel:   sigLabels,
			noLabel:   exclude,
			isPR:      fromBool(false),
			isOpen:    fromBool(true),
			linkedPR:  fromBool(false),
		},
		editIssue: defaultEditIssue,
	}
	err := editIssues(editOpt)
	if err != nil {
		log.Fatal(err)
	}
}
