package githubapi

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/google/go-github/v28/github"
	"golang.org/x/oauth2"
)

type Github struct {
	Client *github.Client
	Token  string
	Owner  string
	Repo   string
}

func NewClient(owner, repo, gtoken string) *Github {
	gtoken = strings.TrimLeft(gtoken, "$")
	token := os.Getenv(gtoken)
	if token == "" {
		log.Fatal("Unauthorized: No token present")
	}
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	config := &Github{Client: client, Owner: owner, Repo: repo, Token: token}

	return config
}

func (g *Github) GetIssues() ([]*github.Issue, error) {
	events, _, err := g.Client.Issues.ListByRepo(context.Background(), g.Owner, g.Repo, nil)
	if err != nil {
		return nil, err
	}

	return events, nil
}

func (g *Github) DuplicateIssueTitle(compTitle string) (int, bool, error) {
	events, err := g.GetIssues()
	if err != nil {
		return 0, false, err
	}

	for _, event := range events {
		title := event.GetTitle()
		if title == compTitle {
			return event.GetNumber(), true, nil
		}
	}

	return 0, false, nil

}

func (g *Github) CreateIssue(title, body string, labels []string) error {
	issreq := &github.IssueRequest{Title: &title, Body: &body, Labels: &labels}
	_, _, err := g.Client.Issues.Create(context.Background(), g.Owner, g.Repo, issreq)
	if err != nil {
		return err
	}
	return nil
}

func (g *Github) CreateIssueComment(issueNum int, comment string) error {
	issueComment := &github.IssueComment{Body: &comment}
	_, _, err := g.Client.Issues.CreateComment(context.Background(), g.Owner, g.Repo, issueNum, issueComment)
	if err != nil {
		return err
	}
	return nil
}

func (g *Github) ReplaceLabel(issueNum int, labels []string) error {
	_, _, err := g.Client.Issues.ReplaceLabelsForIssue(context.Background(), g.Owner, g.Repo, issueNum, labels)
	if err != nil {
		return err
	}

	return nil
}
