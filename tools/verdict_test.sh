#!/bin/bash
#
# Tests for tools/verdict.sh. No dependencies beyond the tool's own.
#
# Usage: tests/verdict_test.sh [name-filter]
#   VERDICT_BASH=/bin/bash tests/verdict_test.sh   # pick the shell under test
#
# Each test builds a throwaway repository, runs the tool from its root, and
# checks the exit status and output. Tests are grouped by the section of the
# specification they exercise. The format itself is tested in task_test.sh;
# here, only what verdict.sh does with a format violation.

LC_ALL=C
export LC_ALL

HERE=$(cd "$(dirname "$0")" && pwd)
VERDICT="$HERE/../tools/verdict.sh"
SH=$(command -v "${VERDICT_BASH:-bash}") || {
  echo "shell not found: ${VERDICT_BASH:-bash}" >&2
  exit 2
}
FILTER=${1:-}
SCRATCH=$(mktemp -d "${TMPDIR:-/tmp}/verdict-test.XXXXXX")
trap 'rm -rf "$SCRATCH"' EXIT

# sha256 of the four bytes "test", as in the specification's example Task.
HASH_TEST=9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08

PASSED=0
FAILED=0
SKIPPED=0
CURRENT=
N=0

# --- fixtures ---------------------------------------------------------------

new_repo() {
  N=$((N + 1))
  REPO=$SCRATCH/repo$N
  mkdir -p "$REPO/breadcrumbs"
}

# put PATH CONTENT: write CONTENT exactly, no newline added.
put() {
  mkdir -p "$REPO/$(dirname "$1")"
  printf '%s' "$2" >"$REPO/$1"
}

# task PATH: write stdin to PATH (relative to the repository root).
task() {
  mkdir -p "$REPO/$(dirname "$1")"
  cat >"$REPO/$1"
}

# V: the front matter lines that make specification 0.1.0 verify a Task.
V='status: IMPLEMENTED
spec_version: "0.1.0"'

# verified_task PATH: a verified Task whose claims come from stdin.
verified_task() {
  { printf -- '---\n%s\nclaims:\n' "$V"; cat; printf -- '---\n\n# A change\n'; } | task "$1"
}

# front PATH LINES: a Task whose front matter is LINES (printf escapes apply).
front() {
  # shellcheck disable=SC2059
  { printf -- '---\n'; printf -- "$2"; printf -- '---\n'; } | task "$1"
}

# run ARGS...: run the tool from the repository root (or $RUN_DIR).
run() {
  (cd "${RUN_DIR:-$REPO}" && PATH=${RUN_PATH:-$PATH} "$SH" "$VERDICT" "$@") \
    >"$SCRATCH/out" 2>"$SCRATCH/err"
  RC=$?
  OUT=$(cat "$SCRATCH/out")
  ERR=$(cat "$SCRATCH/err")
  LAST=$(tail -n 1 "$SCRATCH/out")
}

# --- assertions -------------------------------------------------------------

fail() {
  FAILED=$((FAILED + 1))
  printf 'FAIL %s: %s\n' "$CURRENT" "$1"
  printf '     exit %s\n' "$RC"
  printf '%s\n' "$OUT" | sed 's/^/     out| /'
  printf '%s\n' "$ERR" | sed 's/^/     err| /'
}

pass() { PASSED=$((PASSED + 1)); }

expect_rc() {
  if [ "$RC" = "$1" ]; then pass; else fail "expected exit $1"; fi
}

expect_last() {
  if [ "$LAST" = "$1" ]; then pass; else fail "expected last line '$1'"; fi
}

expect_out() {
  case $OUT in *"$1"*) pass ;; *) fail "expected stdout to contain '$1'" ;; esac
}

expect_err() {
  case $ERR in *"$1"*) pass ;; *) fail "expected stderr to contain '$1'" ;; esac
}

expect_no_verdict_line() {
  if [ -z "$OUT" ]; then pass; else fail "expected no stdout"; fi
}

skip() {
  SKIPPED=$((SKIPPED + 1))
  printf 'SKIP %s: %s\n' "$CURRENT" "$1"
}

# ============================================================================
# Invocation and Task files
# ============================================================================

test_no_arguments_is_usage_error() {
  new_repo
  run
  expect_rc 4
  expect_no_verdict_line
}

test_two_arguments_is_usage_error() {
  new_repo
  verified_task breadcrumbs/a.task.md </dev/null
  verified_task breadcrumbs/b.task.md </dev/null
  run breadcrumbs/a.task.md breadcrumbs/b.task.md
  expect_rc 4
}

test_file_outside_breadcrumbs_is_not_a_task_file() {
  new_repo
  verified_task src/fix.task.md </dev/null
  run src/fix.task.md
  expect_rc 4
  expect_err "outside breadcrumbs/"
}

test_name_without_task_suffix_is_not_a_task_file() {
  new_repo
  verified_task breadcrumbs/TASK-0001.md </dev/null
  run breadcrumbs/TASK-0001.md
  expect_rc 4
  expect_err "does not end in .task.md"
}

test_missing_file_is_usage_error() {
  new_repo
  run breadcrumbs/nope.task.md
  expect_rc 4
  expect_err "no such file"
}

test_absolute_path_is_usage_error() {
  new_repo
  verified_task breadcrumbs/a.task.md </dev/null
  run "$REPO/breadcrumbs/a.task.md"
  expect_rc 4
  expect_err "relative to the repository root"
}

test_dot_dot_in_task_path_is_usage_error() {
  new_repo
  verified_task breadcrumbs/a.task.md </dev/null
  run breadcrumbs/../breadcrumbs/a.task.md
  expect_rc 4
}

test_leading_dot_slash_is_accepted() {
  new_repo
  put a.txt "x"
  printf '  - type: "contains"\n    path: "a.txt"\n    text: "x"\n' |
    verified_task breadcrumbs/a.task.md
  run ./breadcrumbs/a.task.md
  expect_rc 0
}

test_run_outside_repository_root_is_usage_error() {
  new_repo
  verified_task breadcrumbs/a.task.md </dev/null
  mkdir -p "$REPO/src"
  RUN_DIR=$REPO/src run breadcrumbs/a.task.md
  expect_rc 4
  expect_err "run verdict.sh from the repository root"
}

test_spec_example_ids() {
  # From the specification's Task files examples.
  new_repo
  put a.txt "x"
  printf '  - type: "contains"\n    path: "a.txt"\n    text: "x"\n' |
    verified_task breadcrumbs/TASK-0001.task.md
  printf '  - type: "contains"\n    path: "a.txt"\n    text: "x"\n' |
    verified_task breadcrumbs/cli/add-save.task.md
  run breadcrumbs/TASK-0001.task.md
  expect_rc 0
  expect_last "TASK-0001: Confirmed"
  run breadcrumbs/cli/add-save.task.md
  expect_rc 0
  expect_last "add-save: Confirmed"
}

test_ids_are_the_repository_choice() {
  # "How ids are chosen is up to the repository."
  local id
  for id in -draft v1.2 2026_fix 'with space' ''; do
    new_repo
    verified_task "breadcrumbs/$id.task.md" </dev/null
    run "breadcrumbs/$id.task.md"
    if [ "$RC" = 0 ] && [ "$LAST" = "$id: Confirmed" ]; then
      pass
    else
      fail "id '$id': expected exit 0 and '$id: Confirmed'"
    fi
  done
}

test_ids_equal_ignoring_case_are_not_an_error() {
  new_repo
  verified_task breadcrumbs/Fix.task.md </dev/null
  verified_task breadcrumbs/old/fix.task.md </dev/null
  run breadcrumbs/Fix.task.md
  expect_rc 0
  run breadcrumbs/old/fix.task.md
  expect_rc 0
}

# ============================================================================
# Verified Tasks
# ============================================================================

test_implemented_with_spec_version_is_verified() {
  new_repo
  front breadcrumbs/a.task.md "$V\n"
  run breadcrumbs/a.task.md
  expect_rc 0
  expect_last "a: Confirmed"
}

test_proposed_task_is_not_verified() {
  new_repo
  front breadcrumbs/a.task.md 'status: PROPOSED\nspec_version: "0.1.0"\nclaims:\n  - type: "contains"\n    path: "missing.txt"\n    text: "x"\n'
  run breadcrumbs/a.task.md
  expect_rc 2
  expect_last "a: not verified"
}

test_approved_task_is_not_verified() {
  new_repo
  front breadcrumbs/a.task.md 'status: APPROVED\nspec_version: "0.1.0"\n'
  run breadcrumbs/a.task.md
  expect_rc 2
  expect_last "a: not verified"
}

test_any_other_status_is_not_verified() {
  local s
  for s in DONE Implemented IMPLEMENTED_ '' "'IMPLEMENTED '"; do
    new_repo
    front breadcrumbs/a.task.md "status: $s\nspec_version: \"0.1.0\"\n"
    run breadcrumbs/a.task.md
    if [ "$RC" = 2 ]; then pass; else fail "status '$s': expected exit 2"; fi
  done
}

test_missing_status_is_not_verified() {
  new_repo
  front breadcrumbs/a.task.md 'spec_version: "0.1.0"\nclaims:\n'
  run breadcrumbs/a.task.md
  expect_rc 2
}

test_other_spec_version_is_not_verified() {
  local v
  for v in '"0.2.0"' '"0.1"' '"0.1.0 "' '"v0.1.0"'; do
    new_repo
    front breadcrumbs/a.task.md "status: IMPLEMENTED\nspec_version: $v\n"
    run breadcrumbs/a.task.md
    if [ "$RC" = 2 ]; then pass; else fail "spec_version $v: expected exit 2"; fi
  done
}

test_missing_spec_version_is_not_verified() {
  new_repo
  front breadcrumbs/a.task.md 'status: IMPLEMENTED\nclaims:\n'
  run breadcrumbs/a.task.md
  expect_rc 2
  expect_last "a: not verified"
}

test_file_without_front_matter_is_not_verified() {
  new_repo
  printf '# Just a heading\n' | task breadcrumbs/a.task.md
  run breadcrumbs/a.task.md
  expect_rc 2
  new_repo
  : | task breadcrumbs/a.task.md
  run breadcrumbs/a.task.md
  expect_rc 2
  new_repo
  printf -- '---\n%s\n' "$V" | task breadcrumbs/a.task.md
  run breadcrumbs/a.task.md
  expect_rc 2
  new_repo
  printf -- '---\n---\n' | task breadcrumbs/a.task.md
  run breadcrumbs/a.task.md
  expect_rc 2
}

test_unverified_task_is_not_judged() {
  # Neither its claims nor its format: "It does not judge any other Task file."
  new_repo
  front breadcrumbs/a.task.md 'status: PROPOSED\nspec_version: "0.1.0"\nclaims: "none yet"\n'
  run breadcrumbs/a.task.md
  expect_rc 2
  new_repo
  front breadcrumbs/a.task.md 'status: IMPLEMENTED\nspec_version: "0.2.0"\nclaims:\n  - type: "exists"\n    path: "x"\n'
  run breadcrumbs/a.task.md
  expect_rc 2
}

test_status_and_spec_version_in_any_order() {
  new_repo
  front breadcrumbs/a.task.md 'claims:\nspec_version: "0.1.0"\nowner: "me"\nstatus: IMPLEMENTED\n'
  run breadcrumbs/a.task.md
  expect_rc 0
}

test_badly_written_status_or_version_is_an_error() {
  # The values make the file verified; how they are written breaks the format.
  local fm
  for fm in 'status: "IMPLEMENTED"\nspec_version: "0.1.0"\n' \
    "status: 'IMPLEMENTED'\nspec_version: \"0.1.0\"\n" \
    'status: &s IMPLEMENTED\nspec_version: "0.1.0"\n' \
    'status: !!str IMPLEMENTED\nspec_version: "0.1.0"\n' \
    'status: IMPLEMENTED\nspec_version: 0.1.0\n' \
    "status: IMPLEMENTED\nspec_version: '0.1.0'\n"; do
    new_repo
    front breadcrumbs/a.task.md "$fm"
    run breadcrumbs/a.task.md
    if [ "$RC" = 3 ] && [ -z "$OUT" ]; then pass; else fail "front matter '$fm': expected exit 3 and no stdout"; fi
  done
}

# ============================================================================
# Format violations in a verified Task
# ============================================================================

test_format_violation_is_an_error_without_verdict() {
  new_repo
  put a.txt "x"
  printf '  - type: "exists"\n    path: "a.txt"\n' | verified_task breadcrumbs/a.task.md
  run breadcrumbs/a.task.md
  expect_rc 3
  expect_no_verdict_line
  expect_err "breadcrumbs/a.task.md: line"
  expect_err "is not a claim type"
}

test_every_violation_is_reported() {
  new_repo
  printf '  - type: "contains"\n    path: "a.txt"\n  - type: "sha256"\n    path: "b.txt"\n' |
    verified_task breadcrumbs/a.task.md
  run breadcrumbs/a.task.md
  expect_rc 3
  expect_err "a contains claim requires text"
  expect_err "a sha256 claim requires hash"
}

test_format_violations_are_errors() {
  # One case per rule of Format, Front matter and Claims.
  local claims
  for claims in \
    '  - type: "Contains"\n    path: "a.txt"\n    text: "x"\n' \
    '  - path: "a.txt"\n    text: "x"\n' \
    '  - type: "contains"\n    text: "x"\n' \
    '  - type: "contains"\n    path: "a.txt"\n' \
    '  - type: "sha256"\n    path: "a.txt"\n' \
    '  - type: "sha256"\n    path: "a.txt"\n    hash: "9f86d081"\n' \
    '  - type: "sha256"\n    path: "a.txt"\n    hash: "9F86D081884C7D659A2FEAA0C55AD015A3BF4F1B2B0B822CD15D6C15B0F00A08"\n' \
    '  - type: "sha256"\n    path: "missing"\n    hash: "nothex"\n' \
    '  - type: contains\n    path: "a.txt"\n    text: "x"\n' \
    "  - type: 'contains'\n    path: \"a.txt\"\n    text: \"x\"\n" \
    '  - type: "contains"\n    path: "a.txt"\n    text: "x"\n    extra:\n      - "y"\n' \
    '  - type: "contains"\n    path: "a.txt"\n    path: "b.txt"\n    text: "x"\n' \
    '  - type: "contains"\n    path: "a.txt"\n    text: "a\\nb"\n' \
    '  - type: "contains"\n    path: "a.txt"\n    text: "abc\n' \
    '  - "contains a.txt"\n' \
    '  - {type: "contains", path: "a.txt", text: "x"}\n' \
    '\t- type: "contains"\n' \
    '    - type: "contains"\n      path: "a.txt"\n      text: "x"\n  - type: "contains"\n' \
    '  - type: "contains"\n    path: "/etc/passwd"\n    text: "x"\n' \
    '  - type: "contains"\n    path: "a/../a.txt"\n    text: "x"\n' \
    '  - type: "contains"\n    path: "./a.txt"\n    text: "x"\n'; do
    new_repo
    put a.txt "x"
    # shellcheck disable=SC2059
    printf -- "$claims" | verified_task breadcrumbs/a.task.md
    run breadcrumbs/a.task.md
    if [ "$RC" = 3 ]; then pass; else fail "claims '$claims': expected exit 3"; fi
  done
}

test_claims_must_be_a_block_sequence() {
  new_repo
  front breadcrumbs/a.task.md "$V\nclaims: \"a.txt\"\n"
  run breadcrumbs/a.task.md
  expect_rc 3
  new_repo
  front breadcrumbs/a.task.md "$V\nclaims: []\n"
  run breadcrumbs/a.task.md
  expect_rc 3
  new_repo
  front breadcrumbs/a.task.md "$V\nclaims:\n  type: \"contains\"\n  path: \"a\"\n"
  run breadcrumbs/a.task.md
  expect_rc 3
}

test_duplicate_top_level_key_is_an_error() {
  new_repo
  front breadcrumbs/a.task.md "$V\nstatus: IMPLEMENTED\n"
  run breadcrumbs/a.task.md
  expect_rc 3
  expect_err "duplicate key"
}

test_control_character_is_an_error() {
  new_repo
  front breadcrumbs/a.task.md "$V\nnote: \"a\001b\"\n"
  run breadcrumbs/a.task.md
  expect_rc 3
  expect_err "control character"
}

# ============================================================================
# Front matter (Notation)
# ============================================================================

test_crlf_line_endings_are_accepted() {
  new_repo
  put a.txt "x"
  printf -- '---\r\nstatus: IMPLEMENTED\r\nspec_version: "0.1.0"\r\nclaims:\r\n  - type: "contains"\r\n    path: "a.txt"\r\n    text: "x"\r\n---\r\n' |
    task breadcrumbs/a.task.md
  run breadcrumbs/a.task.md
  expect_rc 0
}

test_front_matter_without_final_newline() {
  new_repo
  printf -- '---\n%s\n---' "$V" | task breadcrumbs/a.task.md
  run breadcrumbs/a.task.md
  expect_rc 0
}

test_comments_are_accepted() {
  new_repo
  put a.txt "x"
  task breadcrumbs/a.task.md <<'EOF'
---
# Recorded after review.
status: IMPLEMENTED # done
spec_version: "0.1.0"   # the rules this Task follows
claims:
  # the flag
  - type: "contains"  # plain text
    path: "a.txt"
    text: "x"
---
EOF
  run breadcrumbs/a.task.md
  expect_rc 0
}

test_hash_inside_double_quotes_is_not_a_comment() {
  new_repo
  put a.txt "color: #fff"
  printf '  - type: "contains"\n    path: "a.txt"\n    text: "color: #fff"\n' |
    verified_task breadcrumbs/a.task.md
  run breadcrumbs/a.task.md
  expect_rc 0
}

test_unknown_keys_are_ignored() {
  # The subset binds the keys the specification defines, not the others.
  new_repo
  put a.txt "x"
  task breadcrumbs/a.task.md <<'EOF'
---
parents:
  - "INT-0001"
  - "REQ-0002"
owner:
  name: 'Ada'
  since: 2026
tags: [cli, save]
note: |
  Free text,
  on two lines.
status: IMPLEMENTED
spec_version: "0.1.0"
claims:
  - type: "contains"
    path: "a.txt"
    text: "x"
    note: "ignored by the tool"
---

# Unknown keys
EOF
  run breadcrumbs/a.task.md
  expect_rc 0
}

test_compact_sequence_under_key_is_accepted() {
  new_repo
  put a.txt "x"
  task breadcrumbs/a.task.md <<'EOF'
---
claims:
- type: "contains"
  path: "a.txt"
  text: "x"
status: IMPLEMENTED
spec_version: "0.1.0"
---
EOF
  run breadcrumbs/a.task.md
  expect_rc 0
}

test_blank_lines_in_front_matter_are_accepted() {
  new_repo
  put a.txt "x"
  task breadcrumbs/a.task.md <<'EOF'
---
status: IMPLEMENTED

spec_version: "0.1.0"
claims:

  - type: "contains"
    path: "a.txt"
    text: "x"
---
EOF
  run breadcrumbs/a.task.md
  expect_rc 0
}

test_body_may_contain_front_matter_lookalikes() {
  new_repo
  put a.txt "x"
  task breadcrumbs/a.task.md <<'EOF'
---
status: IMPLEMENTED
spec_version: "0.1.0"
claims:
  - type: "contains"
    path: "a.txt"
    text: "x"
---

# Body

---
status: PROPOSED
---
EOF
  run breadcrumbs/a.task.md
  expect_rc 0
}

test_task_body_may_be_empty() {
  new_repo
  put a.txt "x"
  front breadcrumbs/a.task.md "$V\nclaims:\n  - type: \"contains\"\n    path: \"a.txt\"\n    text: \"x\"\n"
  run breadcrumbs/a.task.md
  expect_rc 0
}

# ============================================================================
# Claims: general
# ============================================================================

test_paths_naming_no_file_are_refuted() {
  local p
  for p in 'a//a.txt' '' 'dir/' 'missing.txt'; do
    new_repo
    mkdir -p "$REPO/a" "$REPO/dir"
    put a.txt "x"
    put a/a.txt "x"
    printf '  - type: "contains"\n    path: "%s"\n    text: "x"\n' "$p" |
      verified_task breadcrumbs/a.task.md
    run breadcrumbs/a.task.md
    if [ "$RC" = 1 ]; then pass; else fail "path '$p': expected exit 1"; fi
  done
}

test_directory_at_path_is_refuted() {
  new_repo
  mkdir -p "$REPO/src"
  printf '  - type: "contains"\n    path: "src"\n    text: ""\n' | verified_task breadcrumbs/a.task.md
  run breadcrumbs/a.task.md
  expect_rc 1
}

test_symlink_content_is_its_target() {
  # Git stores a symbolic link as a file holding its target.
  new_repo
  put real.txt "hello"
  ln -s real.txt "$REPO/link.txt"
  printf '  - type: "contains"\n    path: "link.txt"\n    text: "real.txt"\n' | verified_task breadcrumbs/a.task.md
  run breadcrumbs/a.task.md
  expect_rc 0
  new_repo
  put real.txt "hello"
  ln -s real.txt "$REPO/link.txt"
  printf '  - type: "contains"\n    path: "link.txt"\n    text: "hello"\n' | verified_task breadcrumbs/a.task.md
  run breadcrumbs/a.task.md
  expect_rc 1
  new_repo
  ln -s test "$REPO/link"
  printf '  - type: "sha256"\n    path: "link"\n    hash: "%s"\n' "$HASH_TEST" | verified_task breadcrumbs/a.task.md
  run breadcrumbs/a.task.md
  expect_rc 0
}

test_symlinked_directory_in_path_is_refuted() {
  new_repo
  put real/a.txt "x"
  ln -s real "$REPO/alias"
  printf '  - type: "contains"\n    path: "alias/a.txt"\n    text: "x"\n' | verified_task breadcrumbs/a.task.md
  run breadcrumbs/a.task.md
  expect_rc 1
}

test_unreadable_file_is_refuted() {
  if [ "$(id -u)" = 0 ]; then
    skip "running as root: permissions are not enforced"
    return
  fi
  new_repo
  put secret.txt "x"
  chmod 000 "$REPO/secret.txt"
  printf '  - type: "contains"\n    path: "secret.txt"\n    text: ""\n' | verified_task breadcrumbs/a.task.md
  run breadcrumbs/a.task.md
  expect_rc 1
  expect_out "cannot be read"
  chmod 600 "$REPO/secret.txt"
}

test_paths_with_spaces_and_leading_dashes() {
  new_repo
  put "my dir/-x.txt" "x"
  printf '  - type: "contains"\n    path: "my dir/-x.txt"\n    text: "x"\n' | verified_task breadcrumbs/a.task.md
  run breadcrumbs/a.task.md
  expect_rc 0
}

# ============================================================================
# Contains claims
# ============================================================================

test_contains_spec_example() {
  new_repo
  put src/cli.py 'import os

def save(path):
    pass
'
  printf '  - type: "contains"\n    path: "src/cli.py"\n    text: "def save(path):"\n' |
    verified_task breadcrumbs/a.task.md
  run breadcrumbs/a.task.md
  expect_rc 0
}

test_contains_is_case_sensitive() {
  new_repo
  put src/cli.py 'def Save(path):'
  printf '  - type: "contains"\n    path: "src/cli.py"\n    text: "def save(path):"\n' |
    verified_task breadcrumbs/a.task.md
  run breadcrumbs/a.task.md
  expect_rc 1
  expect_out "text not found"
}

test_contains_missing_file_is_refuted() {
  new_repo
  printf '  - type: "contains"\n    path: "src/cli.py"\n    text: "def save(path):"\n' |
    verified_task breadcrumbs/a.task.md
  run breadcrumbs/a.task.md
  expect_rc 1
  expect_out "no file at this path"
}

test_contains_the_install_version_line() {
  # TASK-0001 claims the installed specification's full Version line.
  new_repo
  put .breadcrumb-kit/breadcrumb-sdd.md '<!-- DO NOT EDIT. -->

# Breadcrumb SDD Specification

Version 0.1.0 (2026-09-26)

Editor: Rodolfo Mendes
'
  printf '  - type: "contains"\n    path: ".breadcrumb-kit/breadcrumb-sdd.md"\n    text: "Version 0.1.0 (2026-09-26)"\n' |
    verified_task breadcrumbs/TASK-0001.task.md
  run breadcrumbs/TASK-0001.task.md
  expect_rc 0
}

test_contains_matches_part_of_a_line() {
  new_repo
  put a.txt 'prefix Version 0.1.0 suffix'
  printf '  - type: "contains"\n    path: "a.txt"\n    text: "Version 0.1.0"\n' |
    verified_task breadcrumbs/a.task.md
  run breadcrumbs/a.task.md
  expect_rc 0
}

test_contains_does_not_span_lines() {
  new_repo
  put a.txt 'abc
def'
  printf '  - type: "contains"\n    path: "a.txt"\n    text: "cd"\n' | verified_task breadcrumbs/a.task.md
  run breadcrumbs/a.task.md
  expect_rc 1
}

test_contains_empty_text_holds_for_any_existing_file() {
  new_repo
  put empty.txt ""
  printf '  - type: "contains"\n    path: "empty.txt"\n    text: ""\n' | verified_task breadcrumbs/a.task.md
  run breadcrumbs/a.task.md
  expect_rc 0
  new_repo
  printf '  - type: "contains"\n    path: "empty.txt"\n    text: ""\n' | verified_task breadcrumbs/a.task.md
  run breadcrumbs/a.task.md
  expect_rc 1
}

test_contains_decodes_escapes() {
  new_repo
  put a.txt 'say "hi" to C:\temp'
  printf '  - type: "contains"\n    path: "a.txt"\n    text: "say \\"hi\\" to C:\\\\temp"\n' |
    verified_task breadcrumbs/a.task.md
  run breadcrumbs/a.task.md
  expect_rc 0
}

test_contains_literal_tab() {
  # A Makefile recipe line begins with a tab.
  new_repo
  put Makefile "$(printf 'build:\n\tgo build ./...\n')"
  printf '  - type: "contains"\n    path: "Makefile"\n    text: "\tgo build"\n' |
    verified_task breadcrumbs/a.task.md
  run breadcrumbs/a.task.md
  expect_rc 0
}

test_contains_utf8_text() {
  new_repo
  put README.md 'Olá, Uberlândia — ok'
  printf '  - type: "contains"\n    path: "README.md"\n    text: "Uberlândia —"\n' |
    verified_task breadcrumbs/a.task.md
  run breadcrumbs/a.task.md
  expect_rc 0
}

test_contains_text_starting_with_dash() {
  new_repo
  put a.txt '--format json'
  printf '  - type: "contains"\n    path: "a.txt"\n    text: "--format"\n' | verified_task breadcrumbs/a.task.md
  run breadcrumbs/a.task.md
  expect_rc 0
}

test_contains_regex_characters_are_literal() {
  new_repo
  put a.txt 'a.c'
  printf '  - type: "contains"\n    path: "a.txt"\n    text: "a.c"\n' | verified_task breadcrumbs/a.task.md
  run breadcrumbs/a.task.md
  expect_rc 0
  new_repo
  put a.txt 'abc'
  printf '  - type: "contains"\n    path: "a.txt"\n    text: "a.c"\n' | verified_task breadcrumbs/a.task.md
  run breadcrumbs/a.task.md
  expect_rc 1
}

test_contains_in_binary_file() {
  new_repo
  printf 'head\000\001MAGIC\000tail' >"$REPO/blob.bin"
  printf '  - type: "contains"\n    path: "blob.bin"\n    text: "MAGIC"\n' | verified_task breadcrumbs/a.task.md
  run breadcrumbs/a.task.md
  expect_rc 0
}

test_contains_crlf_file() {
  new_repo
  printf 'line one\r\nline two\r\n' >"$REPO/a.txt"
  printf '  - type: "contains"\n    path: "a.txt"\n    text: "line two"\n' | verified_task breadcrumbs/a.task.md
  run breadcrumbs/a.task.md
  expect_rc 0
}

# ============================================================================
# Hash claims
# ============================================================================

test_sha256_spec_example() {
  new_repo
  put config/defaults.toml "test"
  printf '  - type: "sha256"\n    path: "config/defaults.toml"\n    hash: "%s"\n' "$HASH_TEST" |
    verified_task breadcrumbs/a.task.md
  run breadcrumbs/a.task.md
  expect_rc 0
}

test_sha256_mismatch_is_refuted() {
  new_repo
  put config/defaults.toml "test
"
  printf '  - type: "sha256"\n    path: "config/defaults.toml"\n    hash: "%s"\n' "$HASH_TEST" |
    verified_task breadcrumbs/a.task.md
  run breadcrumbs/a.task.md
  expect_rc 1
}

test_sha256_missing_file_is_refuted() {
  new_repo
  printf '  - type: "sha256"\n    path: "config/defaults.toml"\n    hash: "%s"\n' "$HASH_TEST" |
    verified_task breadcrumbs/a.task.md
  run breadcrumbs/a.task.md
  expect_rc 1
}

tool_path() { # a PATH holding only the listed tools
  local bin=$SCRATCH/bin-$1 t
  rm -rf "$bin"
  mkdir -p "$bin"
  shift
  for t in "$@"; do
    ln -s "$(command -v "$t")" "$bin/$t"
  done
  printf '%s' "$bin"
}

test_sha256_with_shasum_only() {
  if ! command -v shasum >/dev/null 2>&1; then
    skip "shasum not installed"
    return
  fi
  new_repo
  put a.txt "test"
  printf '  - type: "sha256"\n    path: "a.txt"\n    hash: "%s"\n' "$HASH_TEST" | verified_task breadcrumbs/a.task.md
  RUN_PATH=$(tool_path shasum grep shasum) run breadcrumbs/a.task.md
  expect_rc 0
}

test_sha256_with_sha256sum_only() {
  if ! command -v sha256sum >/dev/null 2>&1; then
    skip "sha256sum not installed"
    return
  fi
  new_repo
  put a.txt "test"
  printf '  - type: "sha256"\n    path: "a.txt"\n    hash: "%s"\n' "$HASH_TEST" | verified_task breadcrumbs/a.task.md
  RUN_PATH=$(tool_path sha256sum grep sha256sum) run breadcrumbs/a.task.md
  expect_rc 0
}

test_no_hashing_tool_is_environment_error() {
  new_repo
  put a.txt "test"
  printf '  - type: "sha256"\n    path: "a.txt"\n    hash: "%s"\n' "$HASH_TEST" | verified_task breadcrumbs/a.task.md
  RUN_PATH=$(tool_path none grep) run breadcrumbs/a.task.md
  expect_rc 4
  expect_err "no sha256sum or shasum"
}

test_contains_needs_no_hashing_tool() {
  new_repo
  put a.txt "x"
  printf '  - type: "contains"\n    path: "a.txt"\n    text: "x"\n' | verified_task breadcrumbs/a.task.md
  RUN_PATH=$(tool_path none2 grep) run breadcrumbs/a.task.md
  expect_rc 0
}

# ============================================================================
# Verdicts
# ============================================================================

test_all_confirmed_is_confirmed() {
  new_repo
  put a.txt "x"
  put b.txt "test"
  printf '  - type: "contains"\n    path: "a.txt"\n    text: "x"\n  - type: "sha256"\n    path: "b.txt"\n    hash: "%s"\n' "$HASH_TEST" |
    verified_task breadcrumbs/a.task.md
  run breadcrumbs/a.task.md
  expect_rc 0
  expect_last "a: Confirmed"
}

test_any_refuted_is_refuted() {
  new_repo
  put a.txt "x"
  printf '  - type: "contains"\n    path: "a.txt"\n    text: "x"\n  - type: "contains"\n    path: "missing"\n    text: "x"\n' |
    verified_task breadcrumbs/a.task.md
  run breadcrumbs/a.task.md
  expect_rc 1
  expect_last "a: Refuted"
}

test_no_claims_is_confirmed() {
  # "Confirmed if every one of its claims is Confirmed": true of no claims.
  new_repo
  front breadcrumbs/a.task.md "$V\n"
  run breadcrumbs/a.task.md
  expect_rc 0
  expect_last "a: Confirmed"
  new_repo
  front breadcrumbs/a.task.md "$V\nclaims:\n"
  run breadcrumbs/a.task.md
  expect_rc 0
}

test_every_claim_is_reported() {
  new_repo
  put a.txt "x"
  printf '  - type: "contains"\n    path: "missing"\n    text: "x"\n  - type: "contains"\n    path: "a.txt"\n    text: "x"\n' |
    verified_task breadcrumbs/a.task.md
  run breadcrumbs/a.task.md
  expect_rc 1
  expect_out "Refuted   contains missing: no file at this path"
  expect_out "Confirmed contains a.txt"
}

test_verdict_is_deterministic() {
  new_repo
  put a.txt "x"
  printf '  - type: "contains"\n    path: "a.txt"\n    text: "x"\n  - type: "contains"\n    path: "b.txt"\n    text: "x"\n' |
    verified_task breadcrumbs/a.task.md
  run breadcrumbs/a.task.md
  local first=$OUT
  run breadcrumbs/a.task.md
  if [ "$OUT" = "$first" ]; then pass; else fail "two runs differ"; fi
}

test_the_specification_example_task() {
  # The fictional Task in the specification's Format section, verbatim.
  new_repo
  put src/cli.py 'def save(path):
    ...
'
  put config/defaults.toml "test"
  task breadcrumbs/add-save.task.md <<'EOF'
---
status: IMPLEMENTED
spec_version: "0.1.0"
claims:
  - type: "contains"
    path: "src/cli.py"
    text: "def save(path):"
  - type: "sha256"
    path: "config/defaults.toml"
    hash: "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08"
---

# Add a save command to the CLI

The CLI can save the current session to a file.
EOF
  run breadcrumbs/add-save.task.md
  expect_rc 0
  expect_last "add-save: Confirmed"
}

test_a_task_can_break_and_hold_again() {
  # Claims are about states, not changes: the same Task, three states.
  new_repo
  put a.txt "v1"
  printf '  - type: "contains"\n    path: "a.txt"\n    text: "v1"\n' | verified_task breadcrumbs/a.task.md
  run breadcrumbs/a.task.md
  expect_rc 0
  put a.txt "v2"
  run breadcrumbs/a.task.md
  expect_rc 1
  put a.txt "v1 again"
  run breadcrumbs/a.task.md
  expect_rc 0
}

# ============================================================================

main() {
  local t
  # shellcheck disable=SC2016
  printf 'verdict.sh tests under %s (%s)\n' "$SH" "$("$SH" -c 'echo $BASH_VERSION')"
  for t in $(compgen -A function test_); do
    case $t in *"$FILTER"*) ;; *) continue ;; esac
    CURRENT=${t#test_}
    RUN_DIR='' RUN_PATH=''
    "$t"
  done
  printf '%d passed, %d failed, %d skipped\n' "$PASSED" "$FAILED" "$SKIPPED"
  [ "$FAILED" -eq 0 ]
}

main
