package artifacts

import (
	"context"
	"crypto/tls"
	"reflect"
	"testing"
)

func TestGitHubService_New(t *testing.T) {
	type args struct {
		ctx     context.Context
		url     string
		token   string
		owner   string
		name    string
		options *tls.Config
	}
	tests := []struct {
		name    string
		g       *GitHubService
		args    args
		want    Connection
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &GitHubService{}
			got, err := g.New(tt.args.ctx, tt.args.url, tt.args.token, tt.args.owner, tt.args.name, tt.args.options)
			if (err != nil) != tt.wantErr {
				t.Errorf("GitHubService.New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GitHubService.New() = %v, want %v", got, tt.want)
			}
		})
	}
}
