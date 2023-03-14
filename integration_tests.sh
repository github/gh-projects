#!/bin/bash

# this is a helper script to perform integration tests of the extension commands
# some errors, such as missing or unused struct tags, can only be detected at runtime

# required environment variables:
# - ORG_NAME: the name of the organization to use for testing for which you are an admin.
# - ITEM_URL: the URL of an issue or pull request in the organization to add to a project owned by ORG_NAME

set -ex

if [[ -z $ORG_NAME ]]; then
    echo "ORG_NAME must be set"
    exit 1
fi

go build -o gh-projects projects.go

PROJECT_NUMBER=$(./gh-projects create --org $ORG_NAME --title clitest --format=json | jq '.number')
./gh-projects view $PROJECT_NUMBER --org $ORG_NAME --format=json  | jq .

./gh-projects list --org $ORG_NAME --format=json  | jq .

COPY_PROJECT_NUMBER=$(./gh-projects copy $PROJECT_NUMBER --source-org $ORG_NAME --target-org $ORG_NAME --title new-copy --format=json | jq '.number')
./gh-projects delete $COPY_PROJECT_NUMBER --org $ORG_NAME --format=json  | jq .

./gh-projects edit $PROJECT_NUMBER --org $ORG_NAME --title edited-clitest --format=json  | jq .
./gh-projects field-list $PROJECT_NUMBER --org $ORG_NAME --format=json  | jq .
FIELD_ID=$(./gh-projects field-create $PROJECT_NUMBER --org $ORG_NAME --data-type TEXT --name custom-text --format=json | jq '.id')
./gh-projects field-delete --id $FIELD_ID --format=json | jq .

if [[ -n $ITEM_URL ]]; then
    ./gh-projects item-add $PROJECT_NUMBER --org $ORG_NAME --url $ITEM_URL --format=json | jq .
fi

ITEM_ID=$(./gh-projects item-create $PROJECT_NUMBER --org $ORG_NAME --title 'draft issue' --format=json | jq '.id')
./gh-projects item-list $PROJECT_NUMBER --org $ORG_NAME --format=json | jq .

./gh-projects item-archive $PROJECT_NUMBER --org $ORG_NAME --id $ITEM_ID --format=json | jq .
./gh-projects item-delete $PROJECT_NUMBER --id $ITEM_ID --org $ORG_NAME --format=json  | jq .
./gh-projects delete $PROJECT_NUMBER --org $ORG_NAME --format=json | jq .