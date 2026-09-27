#!/bin/bash
#
# package.sh - package the Breadcrumb SDD kit for release.
#
# Writes one release asset, OUTDIR/breadcrumb-kit-<version>.zip, where
# <version> is the content of tools/VERSION. Extracted at the root of a
# repository, it installs the kit:
#
#   .breadcrumb-kit/
#     breadcrumb-sdd.md   the specification, with the install header prepended
#     MANIFEST            sha256sum lines for the tooling files, sorted by path
#     tools/
#       <tool>.sh         every tools/*.sh, copied unchanged
#       VERSION           the package version, copied unchanged
#
# Files ending in _test.sh are tests, not tooling: they are never packaged.
#
# The install header is two lines: a do-not-edit notice that carries the
# SHA-256 of breadcrumb-sdd.md, then a blank line. Anyone can check it with
#   tail -n +3 .breadcrumb-kit/breadcrumb-sdd.md | sha256sum
# The MANIFEST covers tooling only, so no seal spans the specification and
# the tooling. Anyone can check it, from .breadcrumb-kit/, with
#   sha256sum -c MANIFEST
#
# Usage: package.sh [--tag TAG] [OUTDIR]
#
# OUTDIR defaults to dist. It must not exist or must be empty, so that no
# stale file is ever published.
# With --tag, package.sh refuses to package unless TAG, without its leading
# "v", equals the content of tools/VERSION.
#
# Runs from anywhere: the sources are read from the repository that holds
# this script.
#
# Exit status: 0 packaged, 1 the sources are not fit to package, 2 usage error.
#
# Needs Bash 3.2 or later, cp, find, mktemp, sort, tail, zip, and sha256sum
# or shasum.

LC_ALL=C
export LC_ALL

NOTICE_HEAD='<!-- DO NOT EDIT. Installed by Breadcrumb SDD packaging. SHA-256 of the source specification:'

usage_error() {
  printf 'package.sh: %s\n' "$1" >&2
  exit 2
}

fail() {
  printf 'package.sh: %s\n' "$1" >&2
  exit 1
}

sha256() { # hashes standard input; prints the hash alone
  local out
  if command -v sha256sum >/dev/null 2>&1; then
    out=$(sha256sum) || return 1
  elif command -v shasum >/dev/null 2>&1; then
    out=$(shasum -a 256) || return 1
  else
    fail "no sha256sum or shasum found"
  fi
  printf '%s' "${out%% *}"
}

main() {
  local tag='' outdir='' here root spec version hash name body tools stage kit zipfile

  while [ "$#" -gt 0 ]; do
    case $1 in
      -h | --help)
        sed -n '3,37s/^# \{0,1\}//p' "$0"
        exit 0
        ;;
      --tag)
        [ "$#" -ge 2 ] || usage_error "--tag needs a value"
        tag=$2
        shift
        ;;
      --)
        shift
        break
        ;;
      -?*) usage_error "unknown option: $1 (usage: package.sh [--tag TAG] [OUTDIR])" ;;
      *)
        [ -z "$outdir" ] || usage_error "more than one OUTDIR: $outdir, $1"
        outdir=$1
        ;;
    esac
    shift
  done
  if [ "$#" -gt 0 ]; then # what follows "--"
    if [ -n "$outdir" ] || [ "$#" -ne 1 ]; then
      usage_error "usage: package.sh [--tag TAG] [OUTDIR]"
    fi
    outdir=$1
  fi
  [ -n "$outdir" ] || outdir=dist

  case $0 in
    */*) here=$(cd "${0%/*}" && pwd) ;;
    *) here=$PWD ;;
  esac
  root=${here%/*}
  spec=$root/breadcrumb-sdd.md

  [ -f "$spec" ] || fail "no specification at $spec"
  [ -f "$root/tools/VERSION" ] || fail "no package version at $root/tools/VERSION"

  # VERSION is one line: the package version, and nothing else.
  version=$(<"$root/tools/VERSION")
  case $version in
    '' | *[!0-9A-Za-z._+-]*) fail "tools/VERSION does not hold a single version: '$version'" ;;
  esac
  [ "$(wc -l <"$root/tools/VERSION" | tr -d ' ')" -le 1 ] || fail "tools/VERSION has more than one line"

  if [ -n "$tag" ] && [ "${tag#v}" != "$version" ]; then
    fail "tag $tag does not match tools/VERSION ($version)"
  fi

  # A source that already carries the header would be sealed twice.
  IFS= read -r body <"$spec"
  case $body in
    "$NOTICE_HEAD"*) fail "$spec already carries an install header: package the source, not an installed copy" ;;
  esac

  # Tooling: every tools/*.sh that is not a test, in byte order.
  tools=$(cd "$root/tools" && find . -maxdepth 1 -type f -name '*.sh' ! -name '*_test.sh' | sed 's|^\./||' | sort)
  [ -n "$tools" ] || fail "no tooling found in $root/tools"

  command -v zip >/dev/null 2>&1 || fail "no zip found"

  if [ -e "$outdir" ]; then
    [ -d "$outdir" ] || fail "$outdir exists and is not a directory"
    [ -z "$(ls -A "$outdir")" ] || fail "$outdir is not empty"
  else
    mkdir -p "$outdir" || fail "cannot create $outdir"
  fi
  outdir=$(cd "$outdir" && pwd) || fail "cannot enter $outdir"
  zipfile=$outdir/breadcrumb-kit-$version.zip

  # The kit is laid out in a staging directory, then zipped from there.
  stage=$(mktemp -d) || fail "cannot create a staging directory"
  # shellcheck disable=SC2064 # expand stage now: it is local to main
  trap "rm -rf '$stage'" EXIT
  kit=$stage/.breadcrumb-kit
  mkdir -p "$kit/tools" || fail "cannot create $kit/tools"

  # The specification: install header, then the source, byte for byte.
  hash=$(sha256 <"$spec") || fail "cannot hash $spec"
  {
    printf '%s %s -->\n\n' "$NOTICE_HEAD" "$hash"
    cat "$spec"
  } >"$kit/breadcrumb-sdd.md" || fail "cannot write breadcrumb-sdd.md"
  [ "$(tail -n +3 "$kit/breadcrumb-sdd.md" | sha256)" = "$hash" ] ||
    fail "the packaged specification does not hash back to its source"

  # The tooling, unchanged, and its seal. MANIFEST paths are relative to
  # .breadcrumb-kit/.
  cp "$root/tools/VERSION" "$kit/tools/VERSION" || fail "cannot copy VERSION"
  while IFS= read -r name; do
    cp "$root/tools/$name" "$kit/tools/$name" || fail "cannot copy $name"
  done <<EOF
$tools
EOF
  {
    while IFS= read -r name; do
      printf '%s  %s\n' "$(sha256 <"$kit/tools/$name")" "tools/$name"
    done <<EOF
$tools
VERSION
EOF
  } | sort -k2 >"$kit/MANIFEST" || fail "cannot write MANIFEST"

  (cd "$stage" && zip -q -X -r "$zipfile" .breadcrumb-kit) || fail "cannot write $zipfile"

  printf 'package.sh: packaged version %s into %s\n' "$version" "$zipfile"
  (cd "$stage" && find .breadcrumb-kit -type f | sort) | sed 's/^/  /'
}

main "$@"
