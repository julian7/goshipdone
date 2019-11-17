package githubclient

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
	"golang.org/x/oauth2"
)

type (
	GitHubClient struct {
		*github.Client
		context.Context
		Owner string
		Name  string
	}

	Release struct {
		Conn *GitHubClient
		ID   int64
		Tag  string
		Ref  string
		Ver  string
	}
)

func New(ctx context.Context, url, token, owner, name string, options *tls.Config) (*GitHubClient, error) {
	client := &GitHubClient{Context: ctx, Name: name, Owner: owner}
	client.connection(ctx, token, options)

	return client, client.setURLs(url)
}

func (c *GitHubClient) NewReleaser(tag, ref, version string) *Release {
	return &Release{
		Conn: c,
		Tag:  tag,
		Ref:  ref,
		Ver:  version,
	}
}

func (c *GitHubClient) connection(ctx context.Context, token string, options *tls.Config) {
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

	c.Client = github.NewClient(oauth2.NewClient(ctx, ts))
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

func (rel *Release) Release(name, notes string) error {
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

func (rel *Release) getReleaseData(name, notes string) *github.RepositoryRelease {
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

func (rel *Release) Upload(filename, location string) error {
	if rel.ID == 0 {
		return errors.New("no release selected")
	}

	file, err := os.Open(location)
	if err != nil {
		return fmt.Errorf("opening file %s for uploading: %w", location, err)
	}

	if _, _, err = rel.Conn.Client.Repositories.UploadReleaseAsset(
		rel.Conn.Context,
		rel.Conn.Owner,
		rel.Conn.Name,
		rel.ID,
		&github.UploadOptions{
			Name: filename,
		},
		file,
	); err != nil {
		return fmt.Errorf("uploading file %s into %v: %w", location, rel, err)
	}

	return nil
}

func (rel *Release) String() string {
	return fmt.Sprintf("%s/%s #%d", rel.Conn.Owner, rel.Conn.Name, rel.ID)
}
