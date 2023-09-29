package internal

import (
	"fmt"
	"io/fs"
	"os"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/go-git/go-git/v5/storage/memory"
	billymem "github.com/go-git/go-billy/v5/memfs"
	"github.com/nicjohnson145/skeley/config"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
	"github.com/forensicanalysis/gitfs"
)


func InputFSFromEnv(logger zerolog.Logger, template string) (fs.FS, error) {
	inputType, err := config.ParseSourceType(viper.GetString(config.InputType))
	if err != nil {
		return nil, err
	}

	switch inputType {
	case config.SourceTypeGit:
		return fsFromGit(logger, template)
	case config.SourceTypeLocal:
		return fs.Sub(os.DirFS(viper.GetString(config.TemplateDir)), template)
	default:
		return nil, fmt.Errorf("unhandled input type %v", inputType)
	}
}

func fsFromGit(logger zerolog.Logger, template string) (fs.FS, error) {
	auth, err := authFromEnv(logger)
	if err != nil {
		return nil, err
	}

	workTree := billymem.New()
	_, err = git.Clone(memory.NewStorage(), workTree, &git.CloneOptions{
		URL: viper.GetString(config.TemplateDir),
		Auth: auth,
	})
	if err != nil {
		return nil, fmt.Errorf("error cloning repo: %w", err)
	}

	ioFS := &gitfs.GitFS{FS: workTree}
	outFS, err := fs.Sub(ioFS, template)
	if err != nil {
		return nil, err
	}

	return outFS, nil
}

func authFromEnv(logger zerolog.Logger) (transport.AuthMethod, error) {
	if viper.GetString(config.Token) != "" {
		logger.Debug().Msg("using token auth")
		user := viper.GetString(config.TokenUser)
		if user == "" {
			user = "_token"
		}

		return &http.BasicAuth{
			Username: user,
			Password: viper.GetString(config.Token),
		}, nil
	}
	if viper.GetString(config.KeyPath) != "" {
		logger.Debug().Msg("using SSH key auth")
		keyBytes, err := os.ReadFile(viper.GetString(config.KeyPath))
		if err != nil {
			return nil, fmt.Errorf("error reading ssh key: %w", err)
		}

		publicKeys, err := ssh.NewPublicKeys("git", keyBytes, "")
		if err != nil {
			return nil, err
		}

		return publicKeys, nil
	}
	logger.Warn().Msg("cloning repo unauthenticated")
	return nil, nil
}
