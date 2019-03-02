package registry

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"strings"

	"github.com/blang/semver"
	"github.com/google/go-github/v24/github"
	"golang.org/x/oauth2"

	"github.com/connectordb/connectordb/src/assets"
)

type Github struct {
	ctx    context.Context
	client *github.Client
}

func NewGithubClient(key string) *Github {
	g := &Github{
		ctx: context.Background(),
	}
	if key != "" {
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: key},
		)
		tc := oauth2.NewClient(g.ctx, ts)

		g.client = github.NewClient(tc)
	} else {
		g.client = github.NewClient(nil)
	}
	return g
}

// Get uses a link to a repo, and extracts all the info necessary to add a plugin to the registry.
// It also validates several properties of the repository, such as a valid license, not archived,
// and that it includes the "connectordb" topic
func (g *Github) Get(link string) (*Plugin, error) {
	if !strings.HasPrefix(link, "http") {
		if !strings.HasPrefix(link, "github.com") {
			return nil, errors.New("Link must be to a github repository")
		}
		link = "https://" + link
	}
	u, err := url.Parse(link)
	if err != nil {
		return nil, err
	}
	if u.Host != "github.com" {
		return nil, errors.New("Link must be to github.com")
	}
	s := strings.Split(u.Path, "/")
	if len(s) < 3 || len(s[0]) > 0 {
		return nil, errors.New("Github url must be in form github.com/{user}/{repo}")
	}

	r, _, err := g.client.Repositories.Get(g.ctx, s[1], s[2])
	if err != nil {
		return nil, err
	}
	b, _ := json.MarshalIndent(r, "", "  ")
	fmt.Printf("Response: %s", string(b))

	rr, _, err := g.client.Repositories.GetLatestRelease(g.ctx, s[1], s[2])
	if err != nil {
		return nil, err
	}

	if rr.TagName == nil {
		return nil, errors.New("Could not find release tag")
	}

	versiontag := *rr.TagName

	if strings.HasPrefix(versiontag, "v") {
		versiontag = versiontag[1:]
	}

	pluginversion, err := semver.Parse(versiontag)
	if err != nil {
		return nil, err
	}

	b, _ = json.MarshalIndent(rr, "", "  ")
	fmt.Printf("\n\nResponse: %s", string(b))

	// And download the connectordb.conf file
	cdbloc := "connectordb.conf"
	conffile, err := g.client.Repositories.DownloadContents(g.ctx, s[1], s[2], cdbloc, &github.RepositoryContentGetOptions{
		Ref: *rr.TagName,
	})
	if err != nil {
		// The connectordb.conf file was not found there, so look inside assets folder
		cdbloc = "assets/connectordb.conf"
		conffile, err = g.client.Repositories.DownloadContents(g.ctx, s[1], s[2], cdbloc, &github.RepositoryContentGetOptions{
			Ref: *rr.TagName,
		})
		if err != nil {
			return nil, fmt.Errorf("Could not find connectordb.conf in root directory or assets folder of tag %s", *rr.TagName)
		}
	}

	// No reason for the config file to be larger than 1MB
	cf, err := ioutil.ReadAll(io.LimitReader(conffile, 1024*1024))
	conffile.Close()
	if err != nil {
		return nil, err
	}

	fmt.Printf("\nConfig file:\n-------------\n%s\n---------------\n\n", string(cf))

	cfg, err := assets.LoadConfigBytes(cf, cdbloc)
	if err != nil {
		return nil, err
	}

	b, _ = json.MarshalIndent(cfg, "", "  ")
	fmt.Printf("\n\nResponse:\n%s", string(b))

	if len(cfg.Plugins) < 0 || len(cfg.Plugins) > 1 {
		return nil, errors.New("There must be exactly one plugin defined in connectordb.conf")
	}

	p := &Plugin{
		Version: pluginversion,
	}

	if r.StargazersCount != nil {
		p.Stars = *r.StargazersCount
	}
	return p, nil
}
