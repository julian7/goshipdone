package modules

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/julian7/goshipdone/ctx"
	"github.com/julian7/goshipdone/internal/artifacts"
	"github.com/julian7/goshipdone/modules"
)

type GitLab struct {
	// Builds specifies which build names should be uploaded to the
	// gitlab release.
	Builds []string
	// Name specifies the repository's name. No default, no detectio (yet).
	// Required.
	Name string
	// Namespace specifies the repository's owning group or username. No default,
	// no detection (yet). Required.
	Namespace string
	// ReleaseName specifies the release's name, using modules.TemplateData.
	// Default: "{{.Version}}"
	ReleaseName string `yaml:"release_name,omitempty"`
	// ReleaseNotes selects the artifact to be used for release notes.
	// It must select a single artifact.
	ReleaseNotes string `yaml:"release_notes"`
	// SkipTLSVerify allows connecting to servers with invalid TLS certs.
	// default: false
	SkipTLSVerify bool `yaml:"skip_tls_verify"`
	// TokenEnv specifies which environment variable the module should look
	// for for GitLab private token. Default: "GITLAB_TOKEN".
	TokenEnv string `yaml:"token_env"`
	// TokenFile specifies which file the module should look for for
	// GitLab private token. Variable expansion is available. Default:
	// "$XDG_CONFIG_HOME/goshipdone/gitlab_token".
	TokenFile string `yaml:"token_file"`
	// URL base URL for gitlab. Provide this only for on-premise GitLab.
	// default: ""
	URL string
}

// nolint: gochecknoinits
func init() {
	modules.RegisterModule(&modules.ModuleRegistration{
		Stage:   "publish",
		Type:    "gitlab",
		Factory: NewGitLab,
	})
}

func NewGitLab() modules.Pluggable {
	return &GitLab{
		ReleaseName:   "{{.Version}}",
		SkipTLSVerify: false,
		TokenEnv:      "GITLAB_TOKEN",
		TokenFile:     "$XDG_CONFIG_HOME/goshipdone/gitlab_token",
		URL:           "",
	}
}

func (mod *GitLab) Run(context *ctx.Context) error {
	var notes string

	relNotes := []*ctx.Artifact(*context.Artifacts.ByID(mod.ReleaseNotes))

	switch len(relNotes) {
	case 0:
		return errors.New("release notes not found")
	case 1:
		content, err := ioutil.ReadFile(relNotes[0].Location)
		if err != nil {
			return fmt.Errorf("reading release notes: %w", err)
		}

		notes = string(content)
	default:
		return errors.New("multiple release notes found")
	}

	client, err := artifacts.NewGitLabClient(
		context.Context,
		mod.URL,
		mod.getToken(context),
		mod.Namespace,
		mod.Name,
		mod.getTLSConfig(),
	)
	if err != nil {
		return err
	}

	td := modules.NewTemplate(context)

	name, err := td.Parse("release-name", mod.ReleaseName)
	if err != nil {
		return fmt.Errorf("parsing release name: %w", err)
	}

	releaser, err := client.NewReleaser(context.Git.Tag, context.Git.Ref, context.Version)
	if err != nil {
		return fmt.Errorf("setting up releaser: %w", err)
	}

	if err := releaser.Release(name, notes); err != nil {
		return fmt.Errorf("releasing: %w", err)
	}

	for _, build := range context.Artifacts.OsArchByIDs(mod.Builds, nil) {
		for _, item := range *build {
			if err := releaser.Upload(item.Filename, item.Location); err != nil {
				return fmt.Errorf("uploading file %s to release %v: %w", item.Location, releaser, err)
			}
		}
	}

	return nil
}

func (mod *GitLab) getToken(context *ctx.Context) string {
	if token, ok := context.Env.Get(mod.TokenEnv); ok {
		return token
	}

	tokenFile := context.Env.Expand(mod.TokenFile)

	if _, err := os.Stat(tokenFile); err != nil {
		log.Printf("tokenfile `%s` not found: %v", tokenFile, err)
		return ""
	}

	data, err := ioutil.ReadFile(tokenFile)
	if err != nil {
		log.Printf("tokenfile read error: %v", err)
	}

	return strings.TrimSpace(string(data))
}

func (mod *GitLab) getTLSConfig() *tls.Config {
	return &tls.Config{
		// nolint: gosec
		InsecureSkipVerify: mod.SkipTLSVerify,
		MinVersion:         tls.VersionTLS12,
	}
}
