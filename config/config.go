package config

import (
	"fmt"
	"strings"
)

type Config struct {
	DryRun          bool            `yaml:"DryRun"`
	DockerRegistry  DockerRegistry  `yaml:"DockerRegistry"`
	RetentionPolicy RetentionPolicy `yaml:"RetentionPolicy"`
}

func (a *Config) Validate() error {
	a.DockerRegistry.URL = strings.TrimRight(a.DockerRegistry.URL, "/")
	if err := validateRetentionPolicy(a.RetentionPolicy.Default); err != nil {
		return err
	}
	for _, exp := range a.RetentionPolicy.Exceptions {
		if err := validateRetentionPolicy(exp); err != nil {
			return err
		}
	}
	return nil
}

type DockerRegistry struct {
	URL      string `yaml:"URL"`
	Username string `yaml:"Username"`
	Password string `yaml:"Password"`
}

type RetentionPolicy struct {
	Default    DefaultRetentionPolicy   `yaml:"Default"`
	Exceptions []AdvanceRetentionPolicy `yaml:"Exceptions"`
}

type IRetentionPolicy interface {
	GetTagsToKeep() int
	GetDaysToKeep() int
	GetKeepLatest() bool
}

func validateRetentionPolicy(rp IRetentionPolicy) error {
	if rp.GetTagsToKeep() < 0 {
		return fmt.Errorf("invalid config: TagsToKeep can not be a negative number")
	}
	if rp.GetDaysToKeep() < 0 {
		return fmt.Errorf("invalid config: DaysToKeep can not be a negative number")
	}
	if rp.GetTagsToKeep() == 0 && rp.GetDaysToKeep() == 0 {
		return fmt.Errorf("invalid config: TagsToKeep and DaysToKeep are both empty")
	}
	return nil
}

type DefaultRetentionPolicy struct {
	TagsToKeep int  `yaml:"TagsToKeep"`
	DaysToKeep int  `yaml:"DaysToKeep"`
	KeepLatest bool `yaml:"KeepLatest"`
}

func (a DefaultRetentionPolicy) GetTagsToKeep() int  { return a.TagsToKeep }
func (a DefaultRetentionPolicy) GetDaysToKeep() int  { return a.DaysToKeep }
func (a DefaultRetentionPolicy) GetKeepLatest() bool { return a.KeepLatest }

type AdvanceRetentionPolicy struct {
	NameMatcher string `yaml:"NameMatcher"`
	TagMatcher  string `yaml:"TagMatcher"`
	TagsToKeep  int    `yaml:"TagsToKeep"`
	DaysToKeep  int    `yaml:"DaysToKeep"`
	KeepLatest  bool   `yaml:"KeepLatest"`
}

func (a AdvanceRetentionPolicy) GetTagsToKeep() int  { return a.TagsToKeep }
func (a AdvanceRetentionPolicy) GetDaysToKeep() int  { return a.DaysToKeep }
func (a AdvanceRetentionPolicy) GetKeepLatest() bool { return a.KeepLatest }
