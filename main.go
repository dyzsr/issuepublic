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

type labelOption struct {
	and []string
	or  []string
	nor []string
}

func getIssuesByLabels(owner string, repo string, opt *labelOption) ([]*github.Issue, error) {
	log.Printf("start getIssuesByLabels")
	defer log.Printf("end getIssuesByLabels")

	if opt == nil {
		opt = &labelOption{}
	}

	var allIssues []*github.Issue
	nextPage := 1
	for nextPage > 0 {
		listOpt := &github.IssueListByRepoOptions{
			Labels: opt.and,
		}
		listOpt.Page = nextPage
		issues, resp, err := cli.Issues.ListByRepo(ctx, owner, repo, listOpt)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		log.Printf("got: %v\n", issueNumbers(issues))
		allIssues = append(allIssues, issues...)
		nextPage = resp.NextPage
	}

	var ret []*github.Issue
	for _, issue := range allIssues {
		if issue.IsPullRequest() {
			continue
		}
		var include, exclude bool
		for _, lb := range opt.or {
			for _, label := range issue.Labels {
				if lb == *label.Name {
					include = true
					break
				}
			}
		}
		if len(opt.or) == 0 {
			include = true
		}
		for _, lb := range opt.nor {
			for _, label := range issue.Labels {
				if lb == *label.Name {
					exclude = true
					break
				}
			}
		}
		if include && !exclude {
			ret = append(ret, issue)
		}
	}

	return ret, nil
}

func printIssue(issue *github.Issue) {
	fmt.Printf("#%d:\n", *issue.Number)
	fmt.Printf("\tname: %s\n", *issue.Title)
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

func editDesc(desc string, slackChannel string) (string, error) {
	matched, err := regexp.MatchString("(?m)^## Description", desc)
	if err != nil {
		return "", errors.WithStack(err)
	}
	ret := desc
	if !matched {
		ret = "## Description\n" + ret
	}
	//matched, err = regexp.MatchString("(?s)(?m)^## SIG slack channel.*^## Score.*^## Mentor", desc)
	//if err != nil {
	//	return "", errors.WithStack(err)
	//}
	//if !matched {
	ret += fmt.Sprintf(`
## SIG slack channel

 %s

## Score

300

## Mentor

- @%s
`, slackChannel, mentor)
	//}

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

type editOption struct {
	labelOption
	editIssue func(*github.Issue) error
}

func editIssues(opt *editOption) error {
	log.Printf("start editIssues")
	defer log.Printf("end editIssues")
	issues, err := getIssuesByLabels(owner, repo, &opt.labelOption)
	if err != nil {
		return errors.WithStack(err)
	}

	if limit == 0 {
		limit = len(issues)
	}
	for i, issue := range issues {
		if i == limit {
			log.Printf("reached limit")
			break
		}

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
		labelOption: labelOption{
			and: include,
			or:  sigLabels,
			nor: exclude,
		},
		editIssue: defaultEditIssue,
	}
	err := editIssues(editOpt)
	if err != nil {
		log.Fatal(err)
	}
}
