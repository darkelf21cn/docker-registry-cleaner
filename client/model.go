package client

import (
	"sort"
	"time"
)

type RawCatalog struct {
	Repositories []string `json:"repositories"`
}

type RawTags struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}

type RawManifests struct {
	SchemaVersion int    `json:"schemaVersion"`
	MediaType     string `json:"mediaType"`
	Config        struct {
		MediaType string `json:"mediaType"`
		Size      int    `json:"size"`
		Digest    string `json:"digest"`
	} `json:"config"`
	Layers []struct {
		MediaType string `json:"mediaType"`
		Size      int    `json:"size"`
		Digest    string `json:"digest"`
	} `json:"layers"`
}

type RawBlob struct {
	Name          string    `json:"name"`
	Tag           string    `json:"tag"`
	ContentDigest string    `json:"content_digest"`
	Architecture  string    `json:"architecture"`
	Container     string    `json:"container"`
	Created       time.Time `json:"created"`
	DockerVersion string    `json:"docker_version"`
	History       []struct {
		Created    time.Time `json:"created"`
		CreatedBy  string    `json:"created_by"`
		EmptyLayer bool      `json:"empty_layer,omitempty"`
	} `json:"history"`
	Os     string `json:"os"`
	Rootfs struct {
		Type    string   `json:"type"`
		DiffIds []string `json:"diff_ids"`
	} `json:"rootfs"`
}

type Image struct {
	Name string
	Tags []Tag
}

type Tag struct {
	Name          string
	ContentDigest string
	Created       time.Time
}

func NewImage(name string) *Image {
	return &Image{
		Name: name,
		Tags: make([]Tag, 0),
	}
}

func (a Image) Len() int {
	return len(a.Tags)
}

func (a Image) Less(i, j int) bool {
	return a.Tags[i].Created.Before(a.Tags[j].Created)
}

func (a Image) Swap(i, j int) {
	a.Tags[i], a.Tags[j] = a.Tags[j], a.Tags[i]
}

func (a Image) SortTagByCreatedDesc() {
	sort.Sort(sort.Reverse(a))
}
