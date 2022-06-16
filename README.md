# A private Docker Registry cleaner

A CLI tool for cleanup private docker registry. The most common usage is to retain the latest X tags for each repository or delete repositories older than Y days. It also supports applying different cleanup rules for some specific repositories or tags by using the RegExp matcher.

The tool is compatible with registries that implement docker's [HTTP API V2](https://docs.docker.com/registry/spec/api/). Tested on Sonatype Nexus Repository Manager.

## Usage

Make a copy of the sample config file under the `example` folder. Edit the file and then run the command below to start a cleanup.

```bash
docker-registry-clean -c config.yaml

```

## Configurations

### Scope

#### Default

The default retention policy applies to all repositories in the registry.

#### Exceptions

Advanced retention policy lets users use different cleanup policies for various repositories or even on some specific tags. Users can set regular expressions on NameMatcher and TagMatcher. Exceptional retention policies are evaluated one by one, it stops on the first successful match. A default retention policy will be used when no one matches. When TagMatcher is defined, retention policy only applies to those matched tags. The rest of the tags will be retained and marked as `excluded`` in the command output.

### Retention Policy

#### ImagesToKeep

The maximum number of tags to keep for each repository. Tags will be cleaned by creation time order.

#### DaysToKeep

Tags created earlier than the DaysToKeep will be removed.

#### KeepLatest

Always keep the tag that is named `latest`.
