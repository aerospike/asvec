#!/bin/bash


set -eo pipefail
if [ -n "$DEBUG" ]; then set -x; fi
trap 'echo "Error: $? at line $LINENO" >&2' ERR

# step 0: Set the environment variables that should be arguments to the script or github action
export VERSION_ARG=0.0.1-dev
export JFROG_CLI_BUILD_NAME=asvec
export JFROG_CLI_BUILD_NUMBER=1  # Or any build number you prefer
export REPO=artifact.aerospike.io/ecosystem-container-dev-local

# Step 1: Build the Docker image and push it to the registry
GIT_ROOT=$(git rev-parse --show-toplevel)
jf  docker buildx bake -f bake.hcl\
                      --set asvec.context="$GIT_ROOT" \
                      --set asvec.dockerfile="$GIT_ROOT/Dockerfile" \
                      --file "$GIT_ROOT/docker/asvec.docker/bake.hcl" \
                      --push \
                      --metadata-file=build-metadata

# Step 2: Extract the first image name and SHA digest
jq -r '.[] | {digest: .["containerimage.digest"], names: .["image.name"] | split(",")} | "\(.names[0])@\(.digest)"' build-metadata > meta-info


# Step 3: Use jf rt build-docker-create to create Docker build information
jf rt build-docker-create \
                    --build-name "$JFROG_CLI_BUILD_NAME" \
                    --build-number "$JFROG_CLI_BUILD_NUMBER" \
                    --image-file ./meta-info\
                    --project ecosystem\
                    ecosystem-container-dev-local

# Step 4: Publish the build information to JFrog Artifactory
jf rt build-publish --detailed-summary --project ecosystem "$JFROG_CLI_BUILD_NAME" "$JFROG_CLI_BUILD_NUMBER"


echo "Docker build information created and published successfully."
