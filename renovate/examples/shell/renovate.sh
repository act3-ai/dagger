#!/bin/bash

# Call renovate to check for updates on a github repo and create PRs if found

dagger -m github.com/act3-ai/dagger/renovate call \
--platform=github \
--endpoint-url=https://api.github.com \
--project=act3-ai/dagger \
--token=env:GITHUB_TOKEN \
  update

# Call renovate to check for updates on a github repo using only a custom.regex manager to find updates.
# Also creates PRs using signed commits from author provided.
dagger -m github.com/act3-ai/dagger/renovate call \
--platform=github \
--endpoint-url=https://api.github.com \
--project=act3-ai/dagger \
--author="$GITHUB_USER" \
--email="$GITHUB_EMAIL" \
--token=env:GITHUB_TOKEN \
--git-private-key=env:GITHUB_PRIVATE_KEY \
--enabled-managers="custom.regex" \
  update
