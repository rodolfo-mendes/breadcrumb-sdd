#!/bin/bash
#
# Tests for tools/task.sh, the syntax validator for one Task file.
#
# Usage: tests/task_test.sh [name-filter]
#   VERDICT_BASH=/bin/bash tests/task_test.sh   # pick the shell under test
#
# Each test writes a Task file into a throwaway repository and runs task.sh
# on it. task.sh prints nothing (missing or empty file), "Ok", or one
# "Error: <message>" line per violation. Tests follow the sections of the
# specification.

LC_ALL=C
export LC_ALL

HERE=$(cd "$(dirname "$0")" && pwd)
TOOL="$HERE/../tools/task.sh"
SH=$(command -v "${VERDICT_BASH:-bash}") || {
  echo "shell not found: ${VERDICT_BASH:-bash}" >&2
  exit 2
}
FILTER=${1:-}
SCRATCH=$(mktemp -d "${TMPDIR:-/tmp}/task-test.XXXXXX")
trap 'rm -rf "$SCRATCH"' EXIT

TAB=$(printf '\t')
HASH=9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08
HEAD='status: IMPLEMENTED
spec_version: "0.1.0"'
GOOD_CLAIM='  - type: "contains"
    path: "src/cli.py"
    text: "def save(path):"'

PASSED=0
FAILED=0
CURRENT=
N=0

# --- fixtures ---------------------------------------------------------------

new_repo() {
  N=$((N + 1))
  REPO=$SCRATCH/repo$N
  : >"$SCRATCH/last"
  mkdir -p "$REPO/breadcrumbs"
}

# file PATH: write stdin to PATH (relative to the repository root).
file() {
  mkdir -p "$REPO/$(dirname "$1")"
  cat >"$REPO/$1"
  printf '%s' "$1" >"$SCRATCH/last" # survives the subshell of a pipeline
}

# front PATH: a Task file whose front matter lines come from stdin.
front() {
  { printf -- '---\n'; cat; printf -- '---\n\n# A change\n'; } | file "$1"
}

# task PATH: a Task file with status and spec_version, claims from stdin.
task() {
  { printf '%s\nclaims:\n' "$HEAD"; cat; } | front "$1"
}

# run [ARGS...]: run task.sh from the repository root; with no arguments,
# on the last file written.
run() {
  local last
  last=$(cat "$SCRATCH/last")
  if [ "$#" -eq 0 ] && [ -n "$last" ]; then set -- "$last"; fi
  (cd "${RUN_DIR:-$REPO}" && "$SH" "$TOOL" "$@") >"$SCRATCH/out" 2>"$SCRATCH/err"
  RC=$?
  OUT=$(cat "$SCRATCH/out")
  ERR=$(cat "$SCRATCH/err")
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

# expect_out LINE...: stdout is exactly these lines.
expect_out() {
  local want='' line
  for line in "$@"; do
    want=$want${want:+
}$line
  done
  if [ "$OUT" = "$want" ]; then pass; else fail "expected output:
$want"; fi
}

expect_ok() {
  expect_rc 0
  expect_out Ok
}

expect_silent() {
  if [ -z "$OUT" ]; then pass; else fail "expected no output"; fi
}

# expect_error SUBSTRING: some "Error: " line contains SUBSTRING.
expect_error() {
  local line
  while IFS= read -r line; do
    case $line in "Error: "*"$1"*)
      pass
      return
      ;;
    esac
  done <<EOF
$OUT
EOF
  fail "expected an Error line mentioning '$1'"
}

# check_error SUBSTRING: a Task file from stdin has a violation.
check_error() {
  new_repo
  file breadcrumbs/t.task.md
  run
  expect_rc 1
  expect_error "$1"
}

# check_success: a Task file from stdin follows the rules.
check_success() {
  new_repo
  file breadcrumbs/t.task.md
  run
  expect_ok
}

# ============================================================================
# The task.sh contract
# ============================================================================

test_valid_task_prints_ok() {
  new_repo
  printf '%s\n' "$GOOD_CLAIM" | task breadcrumbs/a.task.md
  run
  expect_ok
}

test_missing_file_prints_nothing() {
  new_repo
  run breadcrumbs/nope.task.md
  expect_rc 0
  expect_silent
}

test_empty_file_prints_nothing() {
  new_repo
  : | file breadcrumbs/a.task.md
  run
  expect_rc 0
  expect_silent
}

test_whitespace_only_file_is_not_empty() {
  new_repo
  printf '\n' | file breadcrumbs/a.task.md
  run
  expect_rc 1
  expect_out "Error: no front matter: line 1 must be ---"
}

test_every_violation_is_listed() {
  new_repo
  front breadcrumbs/a.task.md <<'EOF'
status: "IMPLEMENTED"
claims:
  - type: "exists"
    path: "/abs"
  - type: contains
    path: "a"
  - type: "sha256"
    path: "b"
    hash: "ABC"
EOF
  run
  expect_rc 1
  expect_out \
    "Error: line 2: status must be written as a plain scalar, not quoted" \
    "Error: line 4: claim 1: '/abs' is not a path" \
    "Error: line 4: claim 1: 'exists' is not a claim type" \
    "Error: line 6: claim 2: 'type' must be a double-quoted string" \
    "Error: line 8: claim 3: hash must be 64 lowercase hexadecimal digits" \
    "Error: spec_version is missing"
}

test_an_invalid_key_does_not_cascade() {
  # 'type' is present but invalid: no "has no type", no type-dependent checks.
  new_repo
  task breadcrumbs/a.task.md <<'EOF'
  - type: contains
    path: "a"
EOF
  run
  expect_out "Error: line 5: claim 1: 'type' must be a double-quoted string"
}

test_block_scalar_value_content_is_skipped() {
  new_repo
  front breadcrumbs/a.task.md <<'EOF'
status: |
  IMPLEMENTED
spec_version: "0.1.0"
EOF
  run
  expect_out "Error: line 2: status must be a plain scalar"
}

test_structural_violation_stops_checking() {
  new_repo
  front breadcrumbs/a.task.md <<'EOF'
status: "IMPLEMENTED"
   stray: "x"
claims:
  - type: "exists"
EOF
  run
  expect_rc 1
  expect_out \
    "Error: line 2: status must be written as a plain scalar, not quoted" \
    "Error: line 3: unexpected indentation"
}

test_path_is_relative_to_the_current_directory() {
  new_repo
  printf '%s\n' "$GOOD_CLAIM" | task breadcrumbs/cli/a.task.md
  RUN_DIR=$REPO/breadcrumbs run cli/a.task.md
  expect_ok
  RUN_DIR=$REPO/breadcrumbs run breadcrumbs/cli/a.task.md
  expect_rc 0
  expect_silent
}

test_name_and_location_are_not_checked() {
  # Whoever passes the file says it is a Task file.
  new_repo
  printf '%s\n' "$GOOD_CLAIM" | task src/notes.md
  run
  expect_ok
}

test_file_name_starting_with_dash() {
  new_repo
  printf '%s\n' "$GOOD_CLAIM" | task breadcrumbs/-draft.task.md
  run -- breadcrumbs/-draft.task.md
  expect_ok
}

test_absolute_path_is_accepted() {
  new_repo
  printf '%s\n' "$GOOD_CLAIM" | task breadcrumbs/a.task.md
  run "$REPO/breadcrumbs/a.task.md"
  expect_ok
}

test_no_argument_is_usage_error() {
  new_repo
  run
  expect_rc 2
  expect_silent
}

test_two_arguments_is_usage_error() {
  new_repo
  run a b
  expect_rc 2
  expect_silent
}

test_unknown_option_is_usage_error() {
  new_repo
  run -v breadcrumbs/a.task.md
  expect_rc 2
}

test_directory_is_usage_error() {
  new_repo
  run breadcrumbs
  expect_rc 2
  expect_silent
}

test_help() {
  new_repo
  run --help
  expect_rc 0
}

# ============================================================================
# Every Task file is validated, whatever its status or spec_version
# ============================================================================

test_any_status_value_is_validated() {
  local st
  for st in IMPLEMENTED PROPOSED DONE; do
    new_repo
    front breadcrumbs/a.task.md <<EOF
status: $st
spec_version: "0.1.0"
claims:
$GOOD_CLAIM
EOF
    run
    if [ "$RC" = 0 ]; then pass; else fail "status $st: expected exit 0"; fi
    expect_out "Ok"
done
}

test_non_implemented_task_is_still_checked() {
  new_repo
  front breadcrumbs/a.task.md <<'EOF'
status: PROPOSED
spec_version: "0.1.0"
claims: "not a sequence"
EOF
  run
  expect_rc 1
  expect_error "claims must be a block sequence"
}

test_any_spec_version_value_is_validated() {
  new_repo
  front breadcrumbs/a.task.md <<EOF
status: IMPLEMENTED
spec_version: "0.2.0"
claims:
$GOOD_CLAIM
EOF
  run
  expect_rc 0
  expect_out "Ok"
}

test_missing_status_is_an_error() {
  new_repo
  front breadcrumbs/a.task.md <<'EOF'
spec_version: "0.1.0"
EOF
  run
  expect_rc 1
  expect_out "Error: status is missing"
}

test_missing_spec_version_is_an_error() {
  new_repo
  front breadcrumbs/a.task.md <<'EOF'
status: IMPLEMENTED
claims: "not a sequence"
EOF
  run
  expect_rc 1
  expect_error "claims must be a block sequence"
  new_repo
  front breadcrumbs/a.task.md <<'EOF'
status: IMPLEMENTED
EOF
  run
  expect_out "Error: spec_version is missing"
}

test_status_without_value_is_an_error() {
  new_repo
  front breadcrumbs/a.task.md <<'EOF'
status:
spec_version: "0.1.0"
EOF
  run
  expect_out "Error: line 2: status has no value"
}

test_spec_version_without_value_is_an_error() {
  new_repo
  front breadcrumbs/a.task.md <<'EOF'
status: IMPLEMENTED
spec_version:
EOF
  run
  expect_out "Error: line 3: spec_version has no value"
}

test_file_without_front_matter_is_an_error() {
  new_repo
  printf '# status: IMPLEMENTED
' | file breadcrumbs/a.task.md
  run
  expect_rc 1
  expect_out "Error: no front matter: line 1 must be ---"
}

test_unclosed_front_matter_is_an_error() {
  new_repo
  printf -- '---
status: IMPLEMENTED
spec_version: "0.1.0"
' | file breadcrumbs/a.task.md
  run
  expect_rc 1
  expect_out "Error: front matter has no closing --- line"
}

test_control_character_is_an_error() {
  # Front matter MUST be valid YAML, and YAML excludes control characters,
  # wherever they are: in a value, a comment, or under an ignored key.
  local c
  for c in '\001' '\033' '\177' '\302\200' '\302\237'; do
    check_error "line 4: control character" < <(printf -- "---\n%s\nnote: \"a${c}b\"\n---\n" "$HEAD")
    check_error "line 4: control character" < <(printf -- "---\n%s\n# a${c}b\n---\n" "$HEAD")
  done
}

test_nel_and_tab_are_not_control_errors() {
  check_success < <(printf -- '---\n%s\nnote: "a\302\205b"\nother: "a\tb"\n---\n' "$HEAD")
}

test_carriage_return_inside_a_line_is_an_error() {
  check_error "line 4: carriage return inside a line" < <(printf -- '---\n%s\nnote: "a\rb"\n---\n' "$HEAD")
}

test_crlf_line_endings_are_not_errors() {
  check_success < <(printf -- '---\r\n%s\r\n---\r\n' "$HEAD")
}

test_sourcing_runs_nothing() {
  # verdict.sh sources task.sh for its functions.
  new_repo
  # shellcheck disable=SC2016 # expanded by the inner shell
  (cd "$REPO" && "$SH" -c '. "$1"; echo sourced' sh "$TOOL") >"$SCRATCH/out" 2>&1
  RC=$?
  OUT=$(cat "$SCRATCH/out")
  expect_rc 0
  expect_out sourced
}

test_empty_front_matter_is_an_error() {
  new_repo
  printf -- '---
---
' | file breadcrumbs/a.task.md
  run
  expect_out "Error: front matter is empty"
}

test_quoted_status_is_an_error() {
  new_repo
  front breadcrumbs/a.task.md <<EOF
status: "IMPLEMENTED"
spec_version: "0.1.0"
claims:
$GOOD_CLAIM
EOF
  run
  expect_rc 1
  expect_error "plain scalar, not quoted"
}

test_tagged_status_is_an_error() {
  new_repo
  front breadcrumbs/a.task.md <<EOF
status: !!str IMPLEMENTED
spec_version: "0.1.0"
claims:
$GOOD_CLAIM
EOF
  run
  expect_rc 1
  expect_error "without anchor, alias or tag"
}

test_flow_status_is_an_error() {
  new_repo
  front breadcrumbs/a.task.md <<'EOF'
status: [IMPLEMENTED]
spec_version: "0.1.0"
EOF
  run
  expect_error "status must be a plain scalar"
}

test_plain_spec_version_is_an_error() {
  new_repo
  front breadcrumbs/a.task.md <<EOF
status: IMPLEMENTED
spec_version: 0.1.0
claims:
$GOOD_CLAIM
EOF
  run
  expect_rc 1
  expect_error "spec_version must be written as a double-quoted string"
}

test_single_quoted_spec_version_is_an_error() {
  new_repo
  front breadcrumbs/a.task.md <<EOF
status: IMPLEMENTED
spec_version: '0.1.0'
claims:
$GOOD_CLAIM
EOF
  run
  expect_rc 1
  expect_error "double-quoted"
}

test_spec_version_with_other_escape_is_an_error() {
  new_repo
  front breadcrumbs/a.task.md <<'EOF'
status: IMPLEMENTED
spec_version: "0.1.0\t"
EOF
  run
  expect_error "spec_version uses an escape other than"
}

test_claims_are_not_evaluated() {
  # src/cli.py does not exist: the verdict would be Refuted, the format is fine.
  new_repo
  printf '%s\n' "$GOOD_CLAIM" | task breadcrumbs/a.task.md
  [ ! -e "$REPO/src/cli.py" ] && pass
  run
  expect_rc 0
  expect_out "Ok"
}

# ============================================================================
# Format
# ============================================================================

test_the_specification_example_task() {
  check_success <<'EOF'
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
}

test_keys_in_any_order() {
  check_success <<EOF
---
claims:
$GOOD_CLAIM
spec_version: "0.1.0"
status: IMPLEMENTED
---
EOF
}

test_empty_body_is_allowed() {
  check_success <<EOF
---
$HEAD
claims:
$GOOD_CLAIM
---
EOF
}

test_body_is_not_judged() {
  check_success <<EOF
---
$HEAD
claims:
$GOOD_CLAIM
---

---
status: broken: [
---
EOF
}

test_task_without_claims_key_is_valid() {
  # A Task with no claims is Confirmed; syntax allows it.
  check_success <<EOF
---
$HEAD
---
EOF
}

test_claims_without_value_is_valid() {
  check_success <<EOF
---
$HEAD
claims:
---
EOF
}

test_claims_flow_sequence_is_an_error() {
  check_error "not a flow sequence" <<EOF
---
$HEAD
claims: []
---
EOF
}

test_claims_scalar_is_an_error() {
  check_error "claims must be a block sequence" <<EOF
---
$HEAD
claims: "src/cli.py"
---
EOF
}

test_claims_mapping_is_an_error() {
  check_error "claims must be a block sequence" <<EOF
---
$HEAD
claims:
  type: "contains"
  path: "a"
  text: "x"
---
EOF
}

test_claims_anchor_is_an_error() {
  check_error "anchors, aliases and tags" <<EOF
---
$HEAD
claims: &c
$GOOD_CLAIM
---
EOF
}

test_compact_claims_sequence_is_allowed() {
  check_success <<EOF
---
$HEAD
claims:
- type: "contains"
  path: "a"
  text: "x"
---
EOF
}

test_dash_on_its_own_line_is_allowed() {
  check_success <<EOF
---
$HEAD
claims:
  -
    type: "contains"
    path: "a"
    text: "x"
---
EOF
}

test_duplicate_top_level_key_is_an_error() {
  check_error "duplicate key 'status'" <<EOF
---
$HEAD
status: IMPLEMENTED
claims:
$GOOD_CLAIM
---
EOF
}

test_duplicate_unknown_key_is_an_error() {
  # Valid YAML has unique keys, ignored ones included.
  check_error "duplicate key 'owner'" <<EOF
---
$HEAD
owner: "a"
owner: "b"
claims:
$GOOD_CLAIM
---
EOF
}

test_top_level_sequence_line_is_an_error() {
  check_error "must be a block mapping" <<EOF
---
$HEAD
- "stray"
claims:
$GOOD_CLAIM
---
EOF
}

test_bad_indentation_is_an_error() {
  check_error "unexpected indentation" <<EOF
---
$HEAD
claims:
    - type: "contains"
      path: "a"
      text: "x"
  - type: "contains"
---
EOF
}

test_status_continuation_line_is_an_error() {
  check_error "unexpected indentation" <<EOF
---
status: IMPLEMENTED
  more
spec_version: "0.1.0"
claims:
$GOOD_CLAIM
---
EOF
}

test_tab_indentation_is_an_error() {
  check_error "tab in indentation" <<EOF
---
$HEAD
claims:
  - type: "contains"
  ${TAB}path: "a"
    text: "x"
---
EOF
}

test_crlf_line_endings_are_allowed() {
  new_repo
  printf -- '---\r\nstatus: IMPLEMENTED\r\nspec_version: "0.1.0"\r\nclaims:\r\n  - type: "contains"\r\n    path: "a"\r\n    text: "x"\r\n---\r\n' |
    file breadcrumbs/t.task.md
  run
  expect_out "Ok"
}

test_blank_lines_are_allowed() {
  check_success <<EOF
---
status: IMPLEMENTED

spec_version: "0.1.0"
claims:

  - type: "contains"

    path: "a"
    text: "x"
---
EOF
}

# --- YAML the specification leaves open -------------------------------------

test_comments_are_allowed() {
  check_success <<'EOF'
---
# A Task
status: IMPLEMENTED   # done
spec_version: "0.1.0" # the rules it follows
claims:
  # the first claim
  - type: "contains"  # exact bytes
    path: "src/cli.py"
    text: "x = 1  # not a comment in the claim"
---
EOF
}

test_hash_inside_quotes_is_not_a_comment() {
  new_repo
  task breadcrumbs/t.task.md <<'EOF'
  - type: "contains"
    path: "a.css"
    text: "color: #fff"
EOF
  run
  expect_out "Ok"
}

test_unknown_keys_may_use_any_yaml() {
  check_success <<'EOF'
---
status: IMPLEMENTED
parents: [INT-0001, REQ-0002]
notes: |
  Free text with "quotes", a # hash
  and: colons.
meta: &m
  owner: 'Ada'
  since: 2026
copy: *m
tags:
- lab
- tooling
spec_version: "0.1.0"
claims:
  - type: "contains"
    path: "a"
    text: "x"
---
EOF
}

test_quoted_keys_are_keys() {
  check_success <<'EOF'
---
"status": IMPLEMENTED
'spec_version': "0.1.0"
claims:
  - "type": "contains"
    path: "a"
    text: "x"
---
EOF
}

# ============================================================================
# Claims
# ============================================================================

test_claim_that_is_a_scalar_is_an_error() {
  check_error "line 5: claim 1 is not a block mapping" <<EOF
---
$HEAD
claims:
  - "contains src/cli.py"
---
EOF
}

test_claim_that_is_a_flow_mapping_is_an_error() {
  check_error "line 5: claim 1 is not a block mapping" <<EOF
---
$HEAD
claims:
  - {type: "contains", path: "a", text: "x"}
---
EOF
}

test_nested_sequence_is_an_error() {
  check_error "not a block mapping" <<EOF
---
$HEAD
claims:
  - - type: "contains"
---
EOF
}

test_claims_are_numbered() {
  check_error "claim 2" <<EOF
---
$HEAD
claims:
$GOOD_CLAIM
  - type: "contains"
    path: "a"
---
EOF
}

test_plain_value_is_an_error() {
  check_error "'type' must be a double-quoted string" <<EOF
---
$HEAD
claims:
  - type: contains
    path: "a"
    text: "x"
---
EOF
}

test_single_quoted_value_is_an_error() {
  check_error "'path' must be a double-quoted string" <<EOF
---
$HEAD
claims:
  - type: "contains"
    path: 'a'
    text: "x"
---
EOF
}

test_value_without_line_break_rule() {
  check_error "not a double-quoted string on one line" <<EOF
---
$HEAD
claims:
  - type: "contains"
    path: "a"
    text: "first line
      second line"
---
EOF
}

test_other_escapes_are_an_error() {
  check_error "uses an escape other than" <<'EOF'
---
status: IMPLEMENTED
spec_version: "0.1.0"
claims:
  - type: "contains"
    path: "a"
    text: "a\nb"
---
EOF
}

test_allowed_escapes() {
  check_success <<'EOF'
---
status: IMPLEMENTED
spec_version: "0.1.0"
claims:
  - type: "contains"
    path: "a"
    text: "say \"hi\" to C:\\temp"
---
EOF
}

test_literal_tab_in_value_is_allowed() {
  new_repo
  printf '  - type: "contains"\n    path: "Makefile"\n    text: "\tgo build"\n' | task breadcrumbs/t.task.md
  run
  expect_out "Ok"
}

test_key_without_value_is_an_error() {
  check_error "'text' has no value" <<EOF
---
$HEAD
claims:
  - type: "contains"
    path: "a"
    text:
---
EOF
}

test_nested_value_is_an_error() {
  check_error "'text' has no value" <<EOF
---
$HEAD
claims:
  - type: "contains"
    path: "a"
    text:
      - "x"
---
EOF
}

test_anchor_in_claim_is_an_error() {
  check_error "anchors, aliases and tags" <<EOF
---
$HEAD
claims:
  - type: "contains"
    path: &p "a"
    text: "x"
---
EOF
}

test_duplicate_key_in_claim_is_an_error() {
  check_error "line 7: claim 1: duplicate key 'path'" <<EOF
---
$HEAD
claims:
  - type: "contains"
    path: "a"
    path: "b"
    text: "x"
---
EOF
}

test_missing_type_is_an_error() {
  check_error "claim 1 has no type" <<EOF
---
$HEAD
claims:
  - path: "a"
    text: "x"
---
EOF
}

test_missing_path_is_an_error() {
  check_error "claim 1 has no path" <<EOF
---
$HEAD
claims:
  - type: "contains"
    text: "x"
---
EOF
}

test_unknown_type_is_an_error() {
  check_error "'exists' is not a claim type" <<EOF
---
$HEAD
claims:
  - type: "exists"
    path: "a"
---
EOF
}

test_type_is_case_sensitive() {
  check_error "'Contains' is not a claim type" <<EOF
---
$HEAD
claims:
  - type: "Contains"
    path: "a"
    text: "x"
---
EOF
}

test_extra_quoted_key_in_claim_is_allowed() {
  # Unknown keys are ignored; the claim still needs double-quoted values.
  check_success <<EOF
---
$HEAD
claims:
  - type: "contains"
    path: "a"
    text: "x"
    note: "why this claim"
---
EOF
}

test_extra_plain_key_in_claim_is_an_error() {
  check_error "'note' must be a double-quoted string" <<EOF
---
$HEAD
claims:
  - type: "contains"
    path: "a"
    text: "x"
    note: why this claim
---
EOF
}

# --- paths ------------------------------------------------------------------

test_invalid_paths_are_errors() {
  local p
  for p in '/etc/passwd' 'a/../b' '../b' './a' 'a/.' '.'; do
    new_repo
    printf '  - type: "contains"\n    path: "%s"\n    text: "x"\n' "$p" | task breadcrumbs/t.task.md
    run
    if [ "$RC" = 1 ]; then pass; else fail "path '$p': expected exit 1"; fi
    expect_error "is not a path"
  done
}

test_paths_the_specification_does_not_exclude() {
  # Only a leading "/" and "." or ".." segments are named; the rest is permitted.
  local p
  for p in 'a//b' 'dir/' '' 'my dir/-x.txt' '.github/workflows/release.yml' 'a..b'; do
    new_repo
    printf '  - type: "contains"\n    path: "%s"\n    text: "x"\n' "$p" | task breadcrumbs/t.task.md
    run
    if [ "$RC" = 0 ]; then pass; else fail "path '$p': expected exit 0"; fi
  done
}

# --- contains ---------------------------------------------------------------

test_contains_requires_text() {
  check_error "a contains claim requires text" <<EOF
---
$HEAD
claims:
  - type: "contains"
    path: "a"
---
EOF
}

test_contains_empty_text_is_allowed() {
  check_success <<EOF
---
$HEAD
claims:
  - type: "contains"
    path: "a"
    text: ""
---
EOF
}

# --- sha256 -----------------------------------------------------------------

test_sha256_requires_hash() {
  check_error "a sha256 claim requires hash" <<EOF
---
$HEAD
claims:
  - type: "sha256"
    path: "a"
---
EOF
}

test_sha256_uppercase_hash_is_an_error() {
  check_error "64 lowercase hexadecimal digits" <<EOF
---
$HEAD
claims:
  - type: "sha256"
    path: "a"
    hash: "$(printf '%s' "$HASH" | tr 'abcdef' 'ABCDEF')"
---
EOF
}

test_sha256_short_hash_is_an_error() {
  check_error "64 lowercase hexadecimal digits" <<EOF
---
$HEAD
claims:
  - type: "sha256"
    path: "a"
    hash: "9f86d081"
---
EOF
}

test_sha256_with_text_key_is_allowed() {
  check_success <<EOF
---
$HEAD
claims:
  - type: "sha256"
    path: "a"
    hash: "$HASH"
    text: "left over"
---
EOF
}


# ============================================================================

main() {
  local t
  # shellcheck disable=SC2016
  printf 'task.sh tests under %s (%s)\n' "$SH" "$("$SH" -c 'echo $BASH_VERSION')"
  for t in $(compgen -A function test_); do
    case $t in *"$FILTER"*) ;; *) continue ;; esac
    CURRENT=${t#test_}
    RUN_DIR=''
    "$t"
  done
  printf '%d passed, %d failed\n' "$PASSED" "$FAILED"
  [ "$FAILED" -eq 0 ]
}

main
