# Pipeline

setup:

- [x] project settings
- [x] git tag detection
- [x] environment variables
- [x] skip publish

build:

- [x] go
- [ ] go run
- [ ] script
- [x] upx
- [ ] gzip
- [x] tar
- [ ] zip
- [x] cut changelog (to create release notes for tag)
- [ ] git-chglog
- [ ] script
- [ ] go run
- [x] checksum
- [ ] GPG sign

publish:

- [ ] S3
- [x] SCP
- [ ] HTTP PUT
- [x] GitHub Releases API
- [ ] GitLab Releases API
- [ ] Artifactory
