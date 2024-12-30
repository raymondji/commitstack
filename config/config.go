package config

import "github.com/raymondji/git-stack/githost"

type Config struct {
	PrimaryColor   string `json:"primaryColor"`
	SecondaryColor string `json:"secondaryColor"`
	Repositories   []RepoConfig
}

type RepoConfig struct {
	DefaultBranch string
	GitHost       githost.Kind
	Gitlab        struct {
		PersonalAccessToken string
	}
}
