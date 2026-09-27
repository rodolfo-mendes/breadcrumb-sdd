#!/bin/bash
#
# Tests for tools/tasks.sh. No dependencies beyond the tool's own.
#
# Usage: tests/tasks_test.sh [name-filter]
#   VERDICT_BASH=/bin/bash tests/tasks_test.sh   # pick the shell under test
#
# Tests the orchestration only: finding the repository root, listing Task
# files, running task.sh on each, and turning its output into the report.
# The syntax rules themselves are tested in tests/task_test.sh.

LC_ALL=C
export LC_ALL

HERE=$(cd "$(dirname "$0")" && pwd)
TOOL="$HERE/../tools/tasks.sh"
SH=$(command -v "${VERDICT_BASH:-bash}") || {
  echo "shell not found: ${VERDICT_BASH:-bash}" >&2
  exit 2
}
FILTER=${1:-}
SCRATCH=$(mktemp -d "${TMPDIR:-/tmp}/tasks-test.XXXXXX")
trap 'rm -rf "$SCRATCH"' EXIT

HASH=9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08
HEAD='status: IMPLEMENTED
spec_version: "0.1.0"'

PASSED=0
FAILED=0
CURRENT=
N=0

# --- fixtures ---------------------------------------------------------------

# new_repo: a repository with Breadcrumb Kit installed and a breadcrumbs/ directory.
new_repo() {
  N=$((N + 1))
  REPO=$SCRATCH/repo$N
  mkdir -p "$REPO/breadcrumbs" "$REPO/.breadcrumb-kit"
}

# file PATH: write stdin to PATH (relative to the repository root).
file() {
  mkdir -p "$REPO/$(dirname "$1")"
  cat >"$REPO/$1"
}

# front PATH: a Task file whose front matter lines come from stdin.
front() {
  { printf -- '---\n'; cat; printf -- '---\n\n# A change\n'; } | file "$1"
}

# task PATH: a verified Task file whose claims come from stdin.
task() {
  { printf '%s\nclaims:\n' "$HEAD"; cat; } | front "$1"
}

GOOD_CLAIM='  - type: "contains"
    path: "src/cli.py"
    text: "def save(path):"'

run() {
  (cd "${RUN_DIR:-$REPO}" && "$SH" "${RUN_TOOL:-$TOOL}" "$@") >"$SCRATCH/out" 2>"$SCRATCH/err"
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

# expect_report LINE...: stdout is exactly these lines, TAB written as "|".
expect_report() {
  local want='' line
  for line in "$@"; do
    want=$want${want:+
}$(printf '%s' "$line" | tr '|' '\t')
  done
  if [ "$OUT" = "$want" ]; then pass; else fail "expected report:
$want"; fi
}

expect_err() {
  case $ERR in *"$1"*) pass ;; *) fail "expected stderr to contain '$1'" ;; esac
}

expect_silent() {
  if [ -z "$OUT" ]; then pass; else fail "expected no output"; fi
}

# ============================================================================
# Report shape (the tools' general form)
# ============================================================================

test_no_violation_prints_nothing() {
  new_repo
  printf '%s\n' "$GOOD_CLAIM" | task breadcrumbs/a.task.md
  run
  expect_rc 0
  expect_silent
}

test_empty_breadcrumbs_prints_nothing() {
  new_repo
  run -v
  expect_rc 0
  expect_silent
}

test_verbose_lists_successes_without_violation_column() {
  new_repo
  printf '%s\n' "$GOOD_CLAIM" | task breadcrumbs/a.task.md
  run -v
  expect_rc 0
  expect_report "a.task.md|Success"
  run --verbose
  expect_report "a.task.md|Success"
}

test_error_line_has_three_columns_and_quoted_violation() {
  new_repo
  front breadcrumbs/a.task.md <<EOF
status: IMPLEMENTED
EOF
  run
  expect_rc 1
  expect_report 'a.task.md|Error|"spec_version is missing"'
}

test_each_violation_is_a_row() {
  new_repo
  front breadcrumbs/a.task.md <<'EOF'
claims:
  - type: "exists"
    path: "a"
EOF
  run
  expect_rc 1
  expect_report \
    "a.task.md|Error|\"line 3: claim 1: 'exists' is not a claim type\"" \
    'a.task.md|Error|"status is missing"' \
    'a.task.md|Error|"spec_version is missing"'
}

test_normal_mode_lists_only_violations() {
  new_repo
  printf '%s\n' "$GOOD_CLAIM" | task breadcrumbs/good.task.md
  front breadcrumbs/bad.task.md <<EOF
status: IMPLEMENTED
EOF
  run
  expect_rc 1
  expect_report 'bad.task.md|Error|"spec_version is missing"'
  run -v
  expect_rc 1
  expect_report 'bad.task.md|Error|"spec_version is missing"' "good.task.md|Success"
}

test_paths_are_relative_to_breadcrumbs_and_sorted() {
  new_repo
  printf '%s\n' "$GOOD_CLAIM" | task breadcrumbs/z.task.md
  printf '%s\n' "$GOOD_CLAIM" | task breadcrumbs/cli/add-save.task.md
  printf '%s\n' "$GOOD_CLAIM" | task breadcrumbs/TASK-0001.task.md
  run -v
  expect_report "TASK-0001.task.md|Success" "cli/add-save.task.md|Success" "z.task.md|Success"
}

test_quotes_in_violation_are_escaped() {
  new_repo
  task breadcrumbs/a.task.md <<'EOF'
  - type: "say \"hi\""
    path: "a"
EOF
  run
  expect_report "a.task.md|Error|\"line 5: claim 1: 'say \\\"hi\\\"' is not a claim type\""
}

test_unknown_option_is_usage_error() {
  new_repo
  run --quiet
  expect_rc 2
  expect_silent
}

test_help() {
  new_repo
  run -h
  expect_rc 0
}

# ============================================================================
# Which files are Task files
# ============================================================================

test_spec_task_file_examples() {
  new_repo
  printf '%s\n' "$GOOD_CLAIM" | task breadcrumbs/TASK-0001.task.md
  printf '%s\n' "$GOOD_CLAIM" | task breadcrumbs/cli/add-save.task.md
  printf '%s\n' "status: broken" | front breadcrumbs/notes.md
  printf '%s\n' "status: broken" | front src/fix.task.md
  run -v
  expect_report "TASK-0001.task.md|Success" "cli/add-save.task.md|Success"
}

test_ids_are_not_judged() {
  new_repo
  printf '%s\n' "$GOOD_CLAIM" | task breadcrumbs/-draft.task.md
  printf '%s\n' "$GOOD_CLAIM" | task breadcrumbs/v1.2.task.md
  printf '%s\n' "$GOOD_CLAIM" | task breadcrumbs/Fix.task.md
  printf '%s\n' "$GOOD_CLAIM" | task breadcrumbs/old/fix.task.md
  run
  expect_rc 0
  expect_silent
}

test_directory_named_like_a_task_is_ignored() {
  new_repo
  mkdir -p "$REPO/breadcrumbs/odd.task.md"
  run -v
  expect_rc 0
  expect_silent
}

test_empty_task_file_is_not_listed() {
  # task.sh skips empty files, so they appear in no report.
  new_repo
  : | file breadcrumbs/empty.task.md
  printf '%s\n' "$GOOD_CLAIM" | task breadcrumbs/a.task.md
  run -v
  expect_rc 0
  expect_report "a.task.md|Success"
}

test_task_files_with_spaces_in_their_path() {
  new_repo
  printf '%s\n' "$GOOD_CLAIM" | task "breadcrumbs/my dir/a b.task.md"
  run -v
  expect_report "my dir/a b.task.md|Success"
}

# ============================================================================
# Repository root
# ============================================================================

test_runs_from_a_subdirectory() {
  new_repo
  mkdir -p "$REPO/src/deep/er"
  front breadcrumbs/a.task.md <<EOF
status: IMPLEMENTED
EOF
  RUN_DIR=$REPO/src/deep/er run
  expect_rc 1
  expect_report 'a.task.md|Error|"spec_version is missing"'
}

test_runs_from_inside_breadcrumbs() {
  new_repo
  printf '%s\n' "$GOOD_CLAIM" | task breadcrumbs/cli/a.task.md
  RUN_DIR=$REPO/breadcrumbs/cli run -v
  expect_report "cli/a.task.md|Success"
}

test_nearest_kit_wins() {
  new_repo
  local outer=$REPO
  front breadcrumbs/outer.task.md <<EOF
status: IMPLEMENTED
EOF
  REPO=$outer/vendor/inner
  mkdir -p "$REPO/breadcrumbs" "$REPO/.breadcrumb-kit"
  printf '%s\n' "$GOOD_CLAIM" | task breadcrumbs/inner.task.md
  run -v
  expect_rc 0
  expect_report "inner.task.md|Success"
}

test_without_kit_is_an_error() {
  new_repo
  rmdir "$REPO/.breadcrumb-kit"
  front breadcrumbs/a.task.md <<EOF
status: IMPLEMENTED
EOF
  run
  expect_rc 2
  expect_silent
  expect_err "no .breadcrumb-kit/"
}

test_kit_as_a_file_does_not_count() {
  new_repo
  rmdir "$REPO/.breadcrumb-kit"
  : >"$REPO/.breadcrumb-kit"
  run
  expect_rc 2
}

test_installed_specification_is_not_needed() {
  new_repo
  printf '%s\n' "$GOOD_CLAIM" | task breadcrumbs/a.task.md
  run -v
  expect_rc 0
  expect_report "a.task.md|Success"
}

test_installed_specification_version_is_ignored() {
  new_repo
  printf '# Breadcrumb SDD Specification\n\nVersion 0.2.0 (2027-01-01)\n' >"$REPO/.breadcrumb-kit/breadcrumb-sdd.md"
  printf '%s\n' "$GOOD_CLAIM" | task breadcrumbs/a.task.md
  run -v
  expect_report "a.task.md|Success"
}

test_no_breadcrumbs_directory_prints_nothing() {
  new_repo
  rmdir "$REPO/breadcrumbs"
  run -v
  expect_rc 0
  expect_silent
}

# ============================================================================
# Finding task.sh
# ============================================================================

test_task_sh_is_found_next_to_tasks_sh() {
  # Installed layout: both tools in .breadcrumb-kit/tools/.
  new_repo
  mkdir "$REPO/.breadcrumb-kit/tools"
  cp "$TOOL" "$HERE/../tools/task.sh" "$REPO/.breadcrumb-kit/tools/"
  front breadcrumbs/a.task.md <<EOF
status: IMPLEMENTED
EOF
  RUN_TOOL=.breadcrumb-kit/tools/tasks.sh run
  expect_rc 1
  expect_report 'a.task.md|Error|"spec_version is missing"'
  RUN_DIR=$REPO/breadcrumbs RUN_TOOL=../.breadcrumb-kit/tools/tasks.sh run
  expect_rc 1
  expect_report 'a.task.md|Error|"spec_version is missing"'
}

test_task_sh_failure_is_an_error() {
  # A task.sh that crashes must not read as "no violation".
  new_repo
  mkdir "$REPO/.breadcrumb-kit/tools"
  cp "$TOOL" "$REPO/.breadcrumb-kit/tools/"
  printf '#!/bin/bash\nexit 2\n' >"$REPO/.breadcrumb-kit/tools/task.sh"
  printf '%s\n' "$GOOD_CLAIM" | task breadcrumbs/a.task.md
  RUN_TOOL=.breadcrumb-kit/tools/tasks.sh run -v
  expect_rc 2
  expect_err "task.sh failed"
}

test_unexpected_task_sh_output_is_an_error() {
  new_repo
  mkdir "$REPO/.breadcrumb-kit/tools"
  cp "$TOOL" "$REPO/.breadcrumb-kit/tools/"
  printf '#!/bin/bash\necho Fine\n' >"$REPO/.breadcrumb-kit/tools/task.sh"
  printf '%s\n' "$GOOD_CLAIM" | task breadcrumbs/a.task.md
  RUN_TOOL=.breadcrumb-kit/tools/tasks.sh run
  expect_rc 2
  expect_err "unexpected output from task.sh"
}

test_missing_task_sh_is_an_error() {
  new_repo
  mkdir "$REPO/.breadcrumb-kit/tools"
  cp "$TOOL" "$REPO/.breadcrumb-kit/tools/"
  RUN_TOOL=.breadcrumb-kit/tools/tasks.sh run
  expect_rc 2
  expect_err "task.sh not found"
}

# ============================================================================
# The install Tasks from Release 1, in the current file naming
# ============================================================================

test_release_1_install_tasks() {
  new_repo
  task breadcrumbs/TASK-0001.task.md <<EOF
  - type: "contains"
    path: ".breadcrumb-kit/breadcrumb-sdd.md"
    text: "Version 0.1.0"
  - type: "sha256"
    path: ".breadcrumb-kit/breadcrumb-sdd.md"
    hash: "$HASH"
EOF
  task breadcrumbs/TASK-0002.task.md <<EOF
  - type: "contains"
    path: ".breadcrumb-kit/VERSION"
    text: "0.1.0"
  - type: "sha256"
    path: ".breadcrumb-kit/MANIFEST"
    hash: "$HASH"
EOF
  run -v
  expect_rc 0
  expect_report "TASK-0001.task.md|Success" "TASK-0002.task.md|Success"
}

# ============================================================================

main() {
  local t
  # shellcheck disable=SC2016
  printf 'tasks.sh tests under %s (%s)\n' "$SH" "$("$SH" -c 'echo $BASH_VERSION')"
  for t in $(compgen -A function test_); do
    case $t in *"$FILTER"*) ;; *) continue ;; esac
    CURRENT=${t#test_}
    RUN_DIR='' RUN_TOOL=''
    "$t"
  done
  printf '%d passed, %d failed\n' "$PASSED" "$FAILED"
  [ "$FAILED" -eq 0 ]
}

main
