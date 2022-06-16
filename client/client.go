package client

import (
	"bytes"
	"crypto/tls"
	"docker-registry-cleaner/config"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

const (
	ApiGetCatalog     = "/v2/_catalog"
	ApiListTags       = "/v2/%s/tags/list"
	ApiGetManifests   = "/v2/%s/manifests/%s"
	ApiDeleteManifest = "/v2/%s/manifests/%s"
	ApiGetBlobs       = "/v2/%s/blobs/%s"

	HttpHeaderAccept = "application/vnd.docker.distribution.manifest.v2+json"
)

type Client struct {
	conf   config.DockerRegistry
	client *http.Client
}

func NewClient(conf config.DockerRegistry) *Client {
	client := &Client{
		conf: conf,
		client: &http.Client{Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}},
	}
	return client
}

func (a Client) ListCatalog() (*RawCatalog, error) {
	catalog := &RawCatalog{}
	err := a.do("GET", ApiGetCatalog, catalog)
	if err != nil {
		return nil, err
	}
	return catalog, nil
}

func (a Client) GetImage(name string) (*Image, error) {
	image := NewImage(name)
	tags, err := a.listRawTags(name)
	if err != nil {
		return nil, err
	}
	for _, tagName := range tags.Tags {
		manifest, err := a.getRawManifests(name, tagName)
		if err != nil {
			return nil, err
		}
		digest, err := a.getTagDigest(name, tagName)
		if err != nil {
			return nil, err
		}
		blob, err := a.getRawBlob(name, manifest.Config.Digest)
		if err != nil {
			return nil, err
		}
		tag := Tag{
			Name:          tagName,
			ContentDigest: digest,
			Created:       blob.Created,
		}
		image.Tags = append(image.Tags, tag)
	}
	image.SortTagByCreatedDesc()
	return image, nil
}

func (a Client) DeleteImageTag(image Image, tag string) error {
	for i, t := range image.Tags {
		if t.Name == tag {
			fmt.Printf("deleting image [%s:%s]\n", image.Name, image.Tags[i].Name)
			return a.deleteImageTag(image.Name, image.Tags[i].ContentDigest)
		}
	}
	return fmt.Errorf("image [%s:%s] doesn't exist", image.Name, tag)
}

func (a Client) listRawTags(name string) (*RawTags, error) {
	val := &RawTags{}
	err := a.do("GET", fmt.Sprintf(ApiListTags, name), val)
	if err != nil {
		return nil, err
	}
	return val, nil
}

func (a Client) getRawManifests(name, tag string) (*RawManifests, error) {
	val := &RawManifests{}
	err := a.do("GET", fmt.Sprintf(ApiGetManifests, name, tag), val)
	if err != nil {
		return nil, err
	}
	return val, nil
}

func (a Client) getTagDigest(name, tag string) (string, error) {
	req, err := a.request("GET", fmt.Sprintf(ApiGetManifests, name, tag), nil)
	if err != nil {
		return "", err
	}
	resp, err := a.client.Do(req)
	if err != nil {
		return "", err
	}
	digest := resp.Header.Get("Docker-Content-Digest")
	if digest == "" {
		return "", fmt.Errorf("header Docker-Content-Digest does not exist")
	}
	return digest, nil
}

func (a Client) getRawBlob(name, digest string) (*RawBlob, error) {
	val := &RawBlob{}
	err := a.do("GET", fmt.Sprintf(ApiGetBlobs, name, digest), val)
	if err != nil {
		return nil, err
	}
	return val, nil
}

func (a Client) deleteImageTag(name, digest string) error {
	req, err := a.request("DELETE", fmt.Sprintf(ApiDeleteManifest, name, digest), nil)
	if err != nil {
		return err
	}
	_, err = a.client.Do(req)
	return err
}

func (a Client) do(method, url string, data interface{}) error {
	req, err := a.request(method, url, nil)
	if err != nil {
		return err
	}
	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf(resp.Status)
	}
	return a.unmarshalResponseBody(resp, data)
}

func (a Client) request(method, url string, input []byte) (*http.Request, error) {
	url = fmt.Sprintf("%s%s", a.conf.URL, url)
	req, err := http.NewRequest(method, url, bytes.NewBuffer(input))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", HttpHeaderAccept)
	req.SetBasicAuth(a.conf.Username, a.conf.Password)
	return req, nil
}

func (a Client) unmarshalResponseBody(resp *http.Response, data interface{}) error {
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, data)
}
