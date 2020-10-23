#!/bin/bash

if [[ -z "$1" ]]; then 
  echo "need tag/version in format v1.x.y"
  exit 1
else
  TAG=$1
fi

CGO_ENABLED=0 go build -o bin/sensu-whois cmd/sensu-whois/main.go
tar czf sensu-whois_${TAG}_linux_amd64.tar.gz bin/

sha512sum sensu-whois_${TAG}_linux_amd64.tar.gz > sensu-whois_${TAG}_sha512_checksums.txt
SHA_HASH_ONLY=$(cut -d " " -f 1 sensu-whois_${TAG}_sha512_checksums.txt)

sed "s/__TAG__/${TAG}/g" sensu/asset_template.tpl > sensu/asset.yaml
sed -i "s/__SHA__/${SHA_HASH_ONLY}/g" sensu/asset.yaml

mkdir -p artifacts
rm -f artifacts/*
mv sensu-whois_${TAG}_linux_amd64.tar.gz sensu-whois_${TAG}_sha512_checksums.txt artifacts/

git add .
git commit -m $TAG
git tag $TAG
git push && git push --tags
