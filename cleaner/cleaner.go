package cleaner

import (
	"docker-registry-cleaner/client"
	"docker-registry-cleaner/config"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/fatih/color"
	"gopkg.in/yaml.v3"
)

type Cleaner struct {
	imageMatchers []ImageMatcher
	defaultPolicy config.DefaultRetentionPolicy
	client        *client.Client
	dryrun        bool
}

type ImageMatcher struct {
	NameMatcher     *regexp.Regexp
	TagMatcher      *regexp.Regexp
	retentionPolicy config.AdvanceRetentionPolicy
}

func NewCleaner(conf config.Config) *Cleaner {
	b, _ := yaml.Marshal(conf)
	fmt.Println("initializing docker-registry-cleaner using configuration:")
	fmt.Println(string(b))
	cleaner := &Cleaner{
		imageMatchers: make([]ImageMatcher, 0),
		defaultPolicy: conf.RetentionPolicy.Default,
		client:        client.NewClient(conf.DockerRegistry),
		dryrun:        conf.DryRun,
	}
	for _, arp := range conf.RetentionPolicy.Exceptions {
		mMatcher, err := regexp.Compile(arp.NameMatcher)
		if err != nil {
			panic(fmt.Errorf("invalid image name matcher: %s", err.Error()))
		}
		if arp.TagMatcher == "" {
			arp.TagMatcher = ".*"
		}
		tMatcher, err := regexp.Compile(arp.TagMatcher)
		if err != nil {
			panic(fmt.Errorf("invalid image tag matcher: %s", err.Error()))
		}
		cleaner.imageMatchers = append(cleaner.imageMatchers, ImageMatcher{
			NameMatcher:     mMatcher,
			TagMatcher:      tMatcher,
			retentionPolicy: arp,
		})
	}
	return cleaner
}

func (a Cleaner) Run() error {
	// Scan registry to get a list of images to be deleted
	fmt.Println("scanning the registry")
	imgs, err := a.client.ListCatalog()
	if err != nil {
		return err
	}
	imagesToDelete := make(map[*client.Image][]string, 0)
	for _, imgName := range imgs.Repositories {
		img, err := a.client.GetImage(imgName)
		if err != nil {
			return err
		}
		tagsToDelete, err := a.scanImage(*img)
		if err != nil {
			return err
		}
		if len(tagsToDelete) > 0 {
			imagesToDelete[img] = tagsToDelete
		}
	}
	fmt.Println("scan completed")

	// Delete images
	if !a.dryrun {
		fmt.Println("cleaning up images")
		for img, tags := range imagesToDelete {
			for _, tag := range tags {
				if err := a.client.DeleteImageTag(*img, tag); err != nil {
					return err
				}
			}
		}
		fmt.Println("cleanup completed")
	}
	return nil
}

func (a Cleaner) scanImage(image client.Image) ([]string, error) {
	var rp config.IRetentionPolicy
	rpName := "default"
	rp = a.defaultPolicy

	// Set to retention policy to AdvanceRetentionPolicy when image name matches
	for _, m := range a.imageMatchers {
		if !m.NameMatcher.MatchString(image.Name) {
			continue
		}

		// Exclude tags that do not match with tag matcher
		// So these tags will NOT be cleaned up by neither AdvanceRetentionPolicy nor DefaultRetentionPolicy
		rpName = m.NameMatcher.String()
		rp = m.retentionPolicy
		for i := len(image.Tags) - 1; i >= 0; i-- {
			tag := image.Tags[i]
			if !m.TagMatcher.MatchString(tag.Name) {
				fmt.Printf("  %s%s\n", PadRight(tag.Name, ' ', 40), color.BlueString("excluded"))
				image.Tags = append(image.Tags[:i], image.Tags[i+1:]...)
			}
		}
	}

	// Run retention policy to get tags to be cleaned
	fmt.Printf("marking image [%s] using [%s] retention policy\n", image.Name, rpName)
	return getTagsForCleanup(rp, image)
}

func getTagsForCleanup(conf config.IRetentionPolicy, image client.Image) ([]string, error) {
	// create a bitmap to store intermediate results
	type Bitmap struct {
		CleanByCount bool
		CleanByDate  bool
	}
	bitmap := make(map[string]Bitmap)
	for _, t := range image.Tags {
		byCount := false
		byDate := false
		if conf.GetTagsToKeep() == 0 {
			byCount = true
		}
		if conf.GetDaysToKeep() == 0 {
			byDate = true
		}
		bitmap[t.Name] = Bitmap{
			CleanByCount: byCount,
			CleanByDate:  byDate,
		}
	}

	// Mark tags for delete by count
	if conf.GetTagsToKeep() > 0 {
		for i := conf.GetTagsToKeep(); i < len(image.Tags); i++ {
			if image.Tags[i].Name == "latest" && conf.GetKeepLatest() {
				continue
			}
			tmp := bitmap[image.Tags[i].Name]
			tmp.CleanByCount = true
			bitmap[image.Tags[i].Name] = tmp
		}
	}

	// Mark tags for delete by date
	if conf.GetDaysToKeep() > 0 {
		for _, t := range image.Tags {
			if t.Name == "latest" && conf.GetKeepLatest() {
				continue
			}
			if t.Created.Add(time.Duration(conf.GetDaysToKeep()) * 24 * time.Hour).Before(time.Now()) {
				tmp := bitmap[t.Name]
				tmp.CleanByDate = true
				bitmap[t.Name] = tmp
			}
		}
	}

	val := make([]string, 0)
	for tag, result := range bitmap {
		if result.CleanByCount && result.CleanByDate {
			val = append(val, tag)
			fmt.Printf("  %s%s\n", PadRight(tag, ' ', 40), color.RedString("delete"))
		} else {
			fmt.Printf("  %s%s\n", PadRight(tag, ' ', 40), color.GreenString("retain"))
		}
	}
	return val, nil
}

func PadRight(str string, pad rune, length int) string {
	if len(str) > length {
		return str + string(pad)
	} else {
		return str + strings.Repeat(string(pad), length-len(str))
	}
}
