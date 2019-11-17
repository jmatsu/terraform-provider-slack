#!/usr/bin/env bash

set -eu

readonly RELEASE_URL="https://github.com/jmatsu/terraform-provider-slack/releases"

die() {
    echo "$@" 1>&2
    exit 1
}

mktemp() {
    command mktemp 2>/dev/null || command mktemp -t tmp
}

last_version() {
  curl -sL -o /dev/null -w %{url_effective} "$RELEASE_URL/latest" |
    rev |
    cut -f1 -d'/'|
    rev
}

is_windows() {
    [[ "$(uname -s)" == "Windows" ]]
}

download() {
    local -r version="$1" temp_file="$TEMP_FILE"

    if is_windows; then
        curl -sL -o "$temp_file" \
            "$RELEASE_URL/download/$VERSION/terraform-provider-slack_$(uname -s)_$(uname -m).zip"
    else
        curl -sL -o "$temp_file" \
            "$RELEASE_URL/download/$VERSION/terraform-provider-slack_$(uname -s)_$(uname -m).tar.gz"
    fi
}

: "${VERSION:=$(last_version)}"

if [[ -z "$VERSION" ]]; then
    die "Unable to get version. Please retry with VERSION=<version>"
fi

readonly TEMP_FILE=$(mktemp)
readonly TEMP_DIR=$(command mktemp -d)

trap "rm ${TEMP_FILE} || : ; rm -fr ${TEMP_DIR} || : ; exit 1"  1 2 3 15

download "$VERSION"

mkdir -p "$TEMP_DIR"

if is_windows; then
    unzip -d "$TEMP_DIR" "$TEMP_FILE"
else
    tar -xf "$TEMP_FILE" -C "$TEMP_DIR"
fi

cp "$TEMP_DIR/terraform-provider-slack" ./terraform-provider-slack_${VERSION}

rm -fr "$TEMP_DIR"