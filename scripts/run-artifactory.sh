#!/usr/bin/env sh

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" > /dev/null && pwd )"

set -euf

docker run -i -t -d --rm -v "${SCRIPT_DIR}/artifactory.lic:/artifactory_extra_conf/artifactory.lic:ro" \
  -p8081:8081 -p8082:8082 -p8080:8080 releases-docker.jfrog.io/PatchSimple/artifactory-pro:7.27.10

echo "Waiting for Artifactory to start"
until curl -sf -u admin:password http://localhost:8081/artifactory/api/system/licenses/; do
    printf '.'
    sleep 4
done
echo ""
# Use decrypted passwords
curl -u admin:password  --output /dev/null --silent --fail localhost:8080/projects/api/system/decrypt -X POST
