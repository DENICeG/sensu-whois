# sensu-whois Asset and Check

Creating releases for sensu is handled by GitHub Actions.

- run `./publish_release.sh v1.x.y` (and commit interactively)
- apply `sensu/asset.yaml` and `sensu/check.yaml` via sensuctl
