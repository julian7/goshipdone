package artifacts

import (
	"context"
	"testing"

	"github.com/xanzy/go-gitlab"
)

func TestGitLabClient_ProjectID(t *testing.T) {
	ctx := context.Background()
	client, _ := gitlab.NewClient("")
	tests := []struct {
		name      string
		namespace string
		project   string
		want      string
	}{
		{name: "nominal case", namespace: "abc", project: "def", want: "abc%2Fdef"},
		{name: "has a dot", namespace: "abc", project: "de.f", want: "abc%2Fde%2Ef"},
		{name: "ns has a dot", namespace: "ab.c", project: "de.f", want: "ab%2Ec%2Fde%2Ef"},
		{name: "subgroup", namespace: "a/b/c", project: "def", want: "a%2Fb%2Fc%2Fdef"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			c := &GitLabClient{
				Client:    client,
				Context:   ctx,
				Namespace: tt.namespace,
				Name:      tt.project,
			}
			if got := c.ProjectID(); got != tt.want {
				t.Errorf("GitLabClient.ProjectID() = %v, want %v", got, tt.want)
			}
		})
	}
}
