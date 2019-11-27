package artifacts

import (
	"context"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/julian7/goshipdone/ctx"
	"gopkg.in/yaml.v3"
)

type (
	Service interface {
		DefaultTokenEnv() string
		DefaultTokenFile() string
		New(ctx context.Context, url, token, owner, name string, opts *tls.Config) (Connection, error)
	}

	Connection interface {
		NewReleaser(tag, ref, version string) (Releaser, error)
	}

	Releaser interface {
		Release(name, notes string) error
		Upload(*ctx.Artifact) error
	}

	Storage struct {
		Service
	}
)

func (s *Storage) GetToken(cx context.Context, tokenEnv, tokenFile string) string {
	context, err := ctx.GetShipContext(cx)
	if err != nil {
		return ""
	}

	for _, envName := range []string{tokenEnv, s.Service.DefaultTokenEnv()} {
		if envName == "" {
			continue
		}

		if token, ok := context.Env.Get(envName); ok {
			return token
		}
	}

	for _, fileName := range []string{tokenFile, s.Service.DefaultTokenFile()} {
		if fileName == "" {
			continue
		}

		if _, err := os.Stat(fileName); err != nil {
			log.Printf("cannot stat tokenfile `%s`: %v", fileName, err)
			break
		}

		data, err := ioutil.ReadFile(tokenFile)
		if err != nil {
			log.Printf("tokenfile read error: %v", err)
			break
		}

		return strings.TrimSpace(string(data))
	}

	return ""
}

func (s *Storage) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind != yaml.ScalarNode {
		return fmt.Errorf("storage is `%v`, not scalar", node.Kind)
	}

	var storageName string
	if err := node.Decode(&storageName); err != nil {
		return fmt.Errorf("storage cannot be decoded: %w", err)
	}

	service, err := s.Load(storageName)
	if err != nil {
		return err
	}

	s.Service = service

	return nil
}

func (s *Storage) Load(name string) (Service, error) {
	switch name {
	case "", "github", "GitHub", "Github":
		return &GitHubService{}, nil
	case "gitlab", "GitLab", "Gitlab":
		return &GitLabService{}, nil
	}

	return nil, fmt.Errorf("invalid storage: `%s`", name)
}

func New(name string) (*Storage, error) {
	s := &Storage{}

	service, err := s.Load(name)
	if err != nil {
		return nil, err
	}

	s.Service = service

	return s, nil
}
