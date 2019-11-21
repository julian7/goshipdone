package artifacts

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/julian7/goshipdone/ctx"
	"github.com/xanzy/go-gitlab"
)

type GitLabService struct{}

type GitLabClient struct {
	*gitlab.Client
	context.Context
	Namespace string
	Name      string
}

type GitLabRelease struct {
	Base string
	Conn *GitLabClient
	ID   string
	Ref  string
	Tag  string
	Ver  string
}

func (*GitLabService) DefaultTokenEnv() string {
	return "GITLAB_TOKEN"
}

func (*GitLabService) DefaultTokenFile() string {
	return "$XDG_CONFIG_HOME/goshipdone/gitlab_token"
}

func (*GitLabService) New(ctx context.Context, url, token, namespace, name string, options *tls.Config) (Connection, error) {
	client := &GitLabClient{Context: ctx, Name: name, Namespace: namespace}
	client.connection(ctx, token, options)

	if url != "" {
		if err := client.SetBaseURL(url); err != nil {
			return nil, err
		}
	}
	return client, nil
}

func (c *GitLabClient) NewReleaser(tag, ref, version string) (Releaser, error) {
	proj, _, err := c.Projects.GetProject(c.ProjectPath(), nil)
	if err != nil {
		return nil, fmt.Errorf("getting project info: %w", err)
	}

	return &GitLabRelease{
		Base: proj.WebURL,
		Conn: c,
		Tag:  tag,
		Ref:  ref,
		Ver:  version,
	}, nil
}

func (c *GitLabClient) ProjectPath() string {
	return fmt.Sprintf("%s/%s", c.Namespace, c.Name)
}

func (c *GitLabClient) ProjectID() string {
	return strings.Replace(url.PathEscape(c.ProjectPath()), ".", "%2E", -1)
}

func (c *GitLabClient) connection(ctx context.Context, token string, options *tls.Config) {
	httpClient := &http.Client{}

	if options != nil {
		httpClient.Transport = &http.Transport{TLSClientConfig: options}
	}

	c.Client = gitlab.NewClient(httpClient, token)
}

func (rel *GitLabRelease) Release(name, notes string) error {
	var release *gitlab.Release

	tag := rel.Tag
	if tag == "" {
		tag = rel.Ver
	}

	projectPath := rel.Conn.ProjectPath()

	_, resp, err := rel.Conn.Client.Releases.GetRelease(rel.Conn.ProjectID(), tag)
	if err != nil {
		if resp.StatusCode != 404 {
			return fmt.Errorf("searching existing release %s: %w", tag, err)
		}

		release, _, err = rel.Conn.Client.Releases.CreateRelease(
			projectPath,
			&gitlab.CreateReleaseOptions{
				Name:        &name,
				Description: &notes,
				Ref:         &rel.Ref,
				TagName:     &tag,
			},
		)
		if err != nil {
			returned, _ := ioutil.ReadAll(resp.Response.Body)
			return fmt.Errorf("creating release %s: %w (%s)", rel.Ver, err, string(returned))
		}
	} else {
		release, _, err = rel.Conn.Client.Releases.UpdateRelease(
			projectPath,
			tag,
			&gitlab.UpdateReleaseOptions{
				Name:        &name,
				Description: &notes,
			},
		)
		if err != nil {
			return fmt.Errorf("editing release %s: %w", tag, err)
		}
	}

	rel.ID = release.Name

	return nil
}

func (rel *GitLabRelease) uploadFile(filename, location string) (*gitlab.ProjectFile, error) {
	file, err := os.Open(location)
	if err != nil {
		return nil, fmt.Errorf("opening file %s for uploading: %w", location, err)
	}
	defer file.Close()

	u := fmt.Sprintf("projects/%s/uploads", rel.Conn.ProjectID())

	b := &bytes.Buffer{}
	w := multipart.NewWriter(b)

	fw, err := w.CreateFormFile("file", filename)
	if err != nil {
		return nil, fmt.Errorf("building file upload form for %s: %w", filename, err)
	}

	_, err = io.Copy(fw, file)
	if err != nil {
		return nil, fmt.Errorf("loading file %s to upload form: %w", filename, err)
	}
	_ = w.Close()

	req, err := rel.Conn.NewRequest("", u, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("setting up new request to upload %s: %w", filename, err)
	}

	req.Body = ioutil.NopCloser(b)
	req.ContentLength = int64(b.Len())
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Method = "POST"
	projFile := &gitlab.ProjectFile{}

	_, err = rel.Conn.Do(req, projFile)
	if err != nil {
		return nil, fmt.Errorf("uploading file %s: %w", filename, err)
	}

	return projFile, nil
}

func (rel *GitLabRelease) Upload(art *ctx.Artifact) error {
	if rel.ID == "" {
		return errors.New("no release selected")
	}

	projectFile, err := rel.uploadFile(
		art.Filename,
		art.Location,
	)
	if err != nil {
		return err
	}

	fileURL := rel.Base + projectFile.URL
	relLink, _, err := rel.Conn.ReleaseLinks.CreateReleaseLink(
		rel.Conn.ProjectPath(),
		rel.ID,
		&gitlab.CreateReleaseLinkOptions{
			Name: &art.Filename,
			URL:  &fileURL,
		},
	)
	if err != nil {
		return fmt.Errorf("uploading file %s into %v: %w", art.Location, rel, err)
	}

	_ = relLink
	return nil
}

func (rel *GitLabRelease) String() string {
	return fmt.Sprintf("%s/%s release %s", rel.Conn.Namespace, rel.Conn.Name, rel.ID)
}
