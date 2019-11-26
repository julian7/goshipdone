package artifacts

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"

	"github.com/blang/semver"
	"github.com/google/go-github/v28/github"
	"github.com/julian7/goshipdone/ctx"
	"golang.org/x/oauth2"
)

type GitHubService struct{}

type GitHubRelease struct {
	Conn *GitHubClient
	ID   int64
	Tag  string
	Ref  string
	Ver  string
}

func (*GitHubService) DefaultTokenEnv() string {
	return "GITHUB_TOKEN"
}

func (*GitHubService) DefaultTokenFile() string {
	return "$XDG_CONFIG_HOME/goshipdone/github_token"
}

func (*GitHubService) New(
	ctx context.Context,
	url, token, owner, name string,
	options *tls.Config,
) (Connection, error) {
	conn := &GitHubClient{Context: ctx, Name: name, Owner: owner}
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})

	if options != nil {
		ctx = context.WithValue(
			ctx,
			oauth2.HTTPClient,
			&http.Client{
				Transport: &http.Transport{TLSClientConfig: options},
			},
		)
	}

	conn.Client = github.NewClient(oauth2.NewClient(ctx, ts))

	return conn, conn.setURLs(url)
}

type GitHubClient struct {
	*github.Client
	context.Context
	Owner string
	Name  string
}

func (c *GitHubClient) NewReleaser(tag, ref, version string) (Releaser, error) {
	return &GitHubRelease{
		Conn: c,
		Tag:  tag,
		Ref:  ref,
		Ver:  version,
	}, nil
}

func (c *GitHubClient) setURLs(baseURL string) error {
	if baseURL == "" {
		return nil
	}

	serverURL, err := url.Parse(baseURL)
	if err != nil {
		return fmt.Errorf("cannot parse github server URL: %w", err)
	}

	apiPortion, err := url.Parse("api/v3/")
	if err != nil {
		return fmt.Errorf("cannot parse api sub-URL: %w", err)
	}

	uploadPortion, err := url.Parse("api/uploads/")
	if err != nil {
		return fmt.Errorf("cannot parse upload api sub-URL: %w", err)
	}

	c.Client.BaseURL = serverURL.ResolveReference(apiPortion)
	c.Client.UploadURL = serverURL.ResolveReference(uploadPortion)

	return nil
}

func (rel *GitHubRelease) Release(name, notes string) error {
	data := rel.getReleaseData(name, notes)

	release, _, err := rel.Conn.Client.Repositories.GetReleaseByTag(
		rel.Conn.Context,
		rel.Conn.Owner,
		rel.Conn.Name,
		data.GetTagName(),
	)
	if err != nil {
		var resp *github.Response

		release, resp, err = rel.Conn.Client.Repositories.CreateRelease(
			rel.Conn.Context,
			rel.Conn.Owner,
			rel.Conn.Name,
			data,
		)
		if err != nil {
			returned, _ := ioutil.ReadAll(resp.Response.Body)
			return fmt.Errorf("creating release %s: %w (%s)", rel.Ver, err, string(returned))
		}
	} else {
		relID := release.GetID()
		if release.GetBody() != "" {
			data.Body = release.Body
		}

		release, _, err = rel.Conn.Client.Repositories.EditRelease(
			rel.Conn.Context,
			rel.Conn.Owner,
			rel.Conn.Name,
			relID,
			data,
		)
		if err != nil {
			return fmt.Errorf("editing release %d: %w", relID, err)
		}
	}

	rel.ID = release.GetID()

	return nil
}

func (rel *GitHubRelease) getReleaseData(name, notes string) *github.RepositoryRelease {
	var prerelease bool

	tag := rel.Tag
	if tag == "" {
		tag = rel.Ver
	}

	if ver, err := semver.Parse(tag); err == nil {
		prerelease = len(ver.Pre) > 0
	}

	return &github.RepositoryRelease{
		Name:       github.String(name),
		TagName:    github.String(tag),
		Body:       github.String(notes),
		Draft:      github.Bool(rel.Tag == ""),
		Prerelease: github.Bool(prerelease),
	}
}

func (rel *GitHubRelease) Upload(art *ctx.Artifact) error {
	if rel.ID == 0 {
		return errors.New("no release selected")
	}

	file, err := os.Open(art.Location)
	if err != nil {
		return fmt.Errorf("opening file %s for uploading: %w", art.Location, err)
	}

	if _, _, err = rel.Conn.Client.Repositories.UploadReleaseAsset(
		rel.Conn.Context,
		rel.Conn.Owner,
		rel.Conn.Name,
		rel.ID,
		&github.UploadOptions{
			Name: art.Filename,
		},
		file,
	); err != nil {
		return fmt.Errorf("uploading file %s into %v: %w", art.Location, rel, err)
	}

	return nil
}

func (rel *GitHubRelease) String() string {
	return fmt.Sprintf("%s/%s #%d", rel.Conn.Owner, rel.Conn.Name, rel.ID)
}
