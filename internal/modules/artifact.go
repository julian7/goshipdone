package modules

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/julian7/goshipdone/ctx"
	"github.com/julian7/goshipdone/internal/artifacts"
	"github.com/julian7/goshipdone/modules"
)

// Artifact is a publish module for artifact storage servers like GitHub, or GitLab.
type Artifact struct {
	// Builds specifies which build names should be uploaded to the
	// github release.
	Builds []string
	// Name specifies the repository's name. No default, no detection (yet).
	// Required.
	Name string
	// Owner specifies the repository's owning organization. No default,
	// no detection (yet). Required.
	Owner string
	// ReleaseName specifies the release's name, using modules.TemplateData.
	// Default: "{{.Version}}"
	ReleaseName string `yaml:"release_name,omitempty"`
	// ReleaseNotes selects the artifact to be used for release notes.
	// It must select a single artifact.
	ReleaseNotes string `yaml:"release_notes"`
	// SkipTLSVerify allows connecting to servers with invalid TLS certs.
	// default: false
	SkipTLSVerify bool `yaml:"skip_tls_verify"`
	// Service specifies which artifact service we are using. Default: "github".
	Storage *artifacts.Storage
	// TokenEnv specifies which environment variable the module should look
	// for for server token. It is discovered from artifacts.Storage if not set.
	// Example: GITHUB_TOKEN.
	TokenEnv string `yaml:"token_env"`
	// TokenFile specifies which file the module should look for artifact storage
	// token. Variable expansion is available. It is discovered
	// from artifacts.Storage if not set. Example:
	// "$XDG_CONFIG_HOME/goshipdone/github_token".
	TokenFile string `yaml:"token_file"`
	// URL base URL for the artifact storage. Provide this only for on-premises services.
	URL string
}

// nolint: gochecknoinits
func init() {
	modules.RegisterModule(&modules.ModuleRegistration{
		Stage:   "publish",
		Type:    "artifact",
		Factory: NewArtifact,
	})
}

// NewArtifact is a factory method for Artifact module
func NewArtifact() modules.Pluggable {
	storage, _ := artifacts.New("github")

	return &Artifact{
		ReleaseName:   "{{.Version}}",
		SkipTLSVerify: false,
		Storage:       storage,
	}
}

// Run uploads previously created artifact into artifact storage provided
// by artifacts.Storage.
func (mod *Artifact) Run(cx context.Context) error {
	context, err := ctx.GetShipContext(cx)
	if err != nil {
		return err
	}

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

	client, err := mod.NewClient(cx)
	if err != nil {
		return err
	}

	td, err := modules.NewTemplate(context)
	if err != nil {
		return err
	}

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
			if err := releaser.Upload(item); err != nil {
				return fmt.Errorf("uploading file %s to release %v: %w", item.Location, releaser, err)
			}
		}
	}

	return nil
}

// NewClient returns a new Storage connection
func (mod *Artifact) NewClient(cx context.Context) (artifacts.Connection, error) {
	return mod.Storage.New(
		cx,
		mod.URL,
		mod.Storage.GetToken(cx, mod.TokenEnv, mod.TokenFile),
		mod.Owner,
		mod.Name,
		mod.getTLSConfig(),
	)
}

func (mod *Artifact) getTLSConfig() *tls.Config {
	return &tls.Config{
		// nolint: gosec
		InsecureSkipVerify: mod.SkipTLSVerify,
		MinVersion:         tls.VersionTLS12,
	}
}
