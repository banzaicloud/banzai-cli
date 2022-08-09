#!/bin/bash

set -euf

RELEASE_TAG="$1"
OWNER='banzaicloud'
REPO='banzaicloud.github.io'
WORKFLOW='cli-docgen.yml'

function main()
{
    curl \
      -X POST \
      -H "Accept: application/vnd.github+json" \
      -H "Authorization: token ${GITHUB_TOKEN}" \
      "https://api.github.com/repos/${OWNER}/${REPO}/actions/workflows/${WORKFLOW}/dispatches" \
      -d "{\"ref\":\"gh-pages\",\"inputs\":{\"cli\":\"banzai-cli\", \"cli-release-tag\": \"${RELEASE_TAG}\", \"cli-base-path\":\"/docs/pipeline/cli/reference/\"}}"
}

main "$@"
