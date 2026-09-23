#!/bin/bash
#
# tasks.sh - check the syntax of every Task file in a repository.
#
# Implements Breadcrumb SDD specification 0.1.0. Where this tool and the
# specification disagree, the specification wins and this tool is wrong.
#
# Runs task.sh, found next to this script, on every Task file: every file
# under breadcrumbs/ whose name ends in .task.md, in byte order of its path.
#
# Usage: tasks.sh [-v | --verbose]
#
# Runs from the repository root: the nearest directory, from the current one
# upward, that holds .breadcrumb-kit/.
#
# Output: one line per violation, as three tab-separated columns: the path
# relative to breadcrumbs/, "Error", and the violation in double quotes. A
# file with several violations has several lines. No output means no
# violation. With -v, Task files that follow the rules are listed too, as
# the path and "Success". Files task.sh skips (empty) are not listed.
#
# Exit status: 0 no violation, 1 at least one violation, 2 usage or
# environment error.
#
# Needs Bash 3.2 or later, find and sort (sed for --help).

LC_ALL=C
export LC_ALL

usage_error() {
  printf 'tasks.sh: %s\n' "$1" >&2
  exit 2
}

quote() { # a violation as a double-quoted string
  local m=$1
  m=${m//\\/\\\\}
  m=${m//\"/\\\"}
  printf '"%s"' "$m"
}

find_root() { # sets ROOT: the nearest directory upward holding .breadcrumb-kit/
  local d=$PWD
  while :; do
    if [ -d "$d/.breadcrumb-kit" ]; then
      ROOT=${d:-/}
      return 0
    fi
    [ -z "$d" ] || [ "$d" = / ] && return 1
    d=${d%/*}
  done
}

main() {
  local verbose=0 here task_tool f rel out rc line errors=0
  while [ "$#" -gt 0 ]; do
    case $1 in
      -v | --verbose) verbose=1 ;;
      -h | --help)
        sed -n '3,25s/^# \{0,1\}//p' "$0"
        exit 0
        ;;
      *) usage_error "unknown argument: $1 (usage: tasks.sh [-v | --verbose])" ;;
    esac
    shift
  done

  case $0 in
    */*) here=$(cd "${0%/*}" && pwd) ;;
    *) here=$PWD ;;
  esac
  task_tool=$here/task.sh
  [ -f "$task_tool" ] || usage_error "task.sh not found next to tasks.sh ($here)"

  find_root || usage_error "no .breadcrumb-kit/ in $PWD or above: run tasks.sh inside a repository with Breadcrumb Kit installed"
  cd "$ROOT" || usage_error "cannot enter $ROOT"

  [ -d breadcrumbs ] || exit 0
  while IFS= read -r -d '' f; do
    rel=${f#breadcrumbs/}
    out=$("$BASH" "$task_tool" -- "$f")
    rc=$?
    case $rc in
      0 | 1) ;;
      *) usage_error "task.sh failed on $f (exit $rc)" ;;
    esac
    [ -n "$out" ] || continue
    while IFS= read -r line; do
      case $line in
        Ok)
          [ "$verbose" -eq 0 ] || printf '%s\t%s\n' "$rel" Success
          ;;
        'Error: '*)
          errors=$((errors + 1))
          printf '%s\t%s\t%s\n' "$rel" Error "$(quote "${line#Error: }")"
          ;;
        *) usage_error "unexpected output from task.sh on $f: $line" ;;
      esac
    done <<EOF
$out
EOF
  done < <(find breadcrumbs -type f -name '*.task.md' -print0 | sort -z)

  [ "$errors" -eq 0 ]
}

main "$@"
