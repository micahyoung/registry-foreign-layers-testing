#!/bin/bash
set -o errexit -o pipefail -o nounset

http_registry_host=localhost:5000
test_image=${http_registry_host}/test

# spin up docker-distribution registry that allows foreign layers
cat > config.yml <<EOF
version: 0.1
log:
  fields:
    service: registry
storage:
  cache:
    blobdescriptor: inmemory
  filesystem:
    rootdirectory: /var/lib/registry
http:
  addr: :5000
  headers:
    X-Content-Type-Options: [nosniff]
health:
  storagedriver:
    enabled: true
    interval: 10s
    threshold: 3
# Allow foreign layers
validation:
  manifests:
    urls:
      allow:
        - ^https?://
EOF
docker run -d -p5000:5000 -v $PWD/config.yml:/etc/docker/registry/config.yml --name reg registry:2
trap "docker rm -f reg >/dev/null" EXIT

# push image without layer blob
go run main.go $test_image false

# attempt to pull, see it fail
if docker pull $test_image; then
  echo invalid test, this should have failed
  exit 1
else
  echo failed as expected
fi

# push image with layer blob
go run main.go $test_image

# pull successfully
docker pull $test_image

echo success
