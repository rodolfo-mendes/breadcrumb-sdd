#!/bin/bash
#
# verdict.sh - the verdict of one Breadcrumb SDD Task.
#
# Implements Breadcrumb SDD specification 0.1.0. Where this tool and the
# specification disagree, the specification wins and this tool is wrong.
#
# Usage: verdict.sh breadcrumbs/<...>/<id>.task.md
#
# Run it from the repository root. It reads files as they are in the working
# tree; to judge a commit, check that commit out first.
#
# Specification 0.1.0 verifies a Task file only if its status is IMPLEMENTED
# and its spec_version is "0.1.0", and does not judge any other Task file. A
# verified Task file that breaks the format is an error and has no verdict.
# The format is checked by task.sh, which must sit next to this script.
#
# Exit status:
#   0  Confirmed
#   1  Refuted
#   2  no verdict: the Task file is not verified
#   3  no verdict: the Task file is verified but breaks the format
#   4  usage or environment error
#
# The last line on stdout is "<id>: Confirmed", "<id>: Refuted" or
# "<id>: not verified". Format violations go to stderr, with no such line.
#
# Needs Bash 3.2 or later, grep, sha256sum or shasum, and readlink for
# claims on symbolic links.

LC_ALL=C
export LC_ALL

case $0 in
  */*) HERE=$(cd "${0%/*}" && pwd) ;;
  *) HERE=$PWD ;;
esac
if [ ! -f "$HERE/task.sh" ]; then
  printf 'verdict.sh: task.sh not found next to verdict.sh (%s)\n' "$HERE" >&2
  exit 4
fi
# shellcheck source=task.sh disable=SC1091
. "$HERE/task.sh"

TASK=
ID=

# Defined after task.sh, so it replaces task.sh's version.
usage_error() {
  printf 'verdict.sh: %s\n' "$1" >&2
  exit 4
}

names_path() { # relative, "/"-separated, no empty, "." or ".." segment
  case "/$1/" in
    *//* | */./* | */../*) return 1 ;;
  esac
  return 0
}

# ---------------------------------------------------------------------------
# Task file and verification.

check_task_file() {
  local name
  case $TASK in
    /*) usage_error "pass the Task path relative to the repository root, not $TASK" ;;
  esac
  names_path "$TASK" || usage_error "not a valid path: $TASK"
  case $TASK in
    breadcrumbs/*) ;;
    *) usage_error "not a Task file: $TASK is outside breadcrumbs/" ;;
  esac
  name=${TASK##*/}
  case $name in
    *.task.md) ;;
    *) usage_error "not a Task file: $name does not end in .task.md" ;;
  esac
  [ -f "$TASK" ] || usage_error "no such file: $TASK (run verdict.sh from the repository root)"
  [ -r "$TASK" ] || usage_error "cannot read $TASK"
  ID=${name%.task.md}
}

# The value of top-level key $1 as YAML reads it, when it is a scalar on one
# line. Sets VAL; fails when the key is absent or its value is anything else.
top_value() {
  local i=0 top=${FM_IND[0]} v
  while [ "$i" -lt "$FM_N" ]; do
    if [ "${FM_IND[i]}" -eq "$top" ] && ! is_seq_item "${FM_TXT[i]}" &&
      split_entry "${FM_TXT[i]}" && [ "$KEY" = "$1" ]; then
      # A value continued on the lines below is not a one-line scalar.
      [ $((i + 1)) -lt "$FM_N" ] && [ "${FM_IND[i + 1]}" -gt "$top" ] && return 1
      # An anchor or a tag does not change the value.
      v=$REST
      while :; do
        case $v in
          '&'*' '* | '!'*' '*)
            trim_leading "${v#* }"
            v=$TRIMMED
            ;;
          *) break ;;
        esac
      done
      case $v in
        '"'*) decode_dq "$v" || return 1; VAL=$DQ ;;
        "'"*) decode_sq "$v" || return 1; VAL=$SQ ;;
        '' | '&'* | '!'* | '*'* | '['* | '{'* | '|'* | '>'*) return 1 ;;
        *) VAL=$v ;;
      esac
      return 0
    fi
    i=$((i + 1))
  done
  return 1
}

# Whether specification 0.1.0 verifies the Task file: status IMPLEMENTED and
# spec_version "0.1.0", however they are written. Writing them wrongly is a
# format violation, which only a verified Task file can have.
verified() {
  read_front_matter "$TASK" || return 1
  [ "$FM_N" -gt 0 ] || return 1
  top_value status && [ "$VAL" = IMPLEMENTED ] || return 1
  top_value spec_version && [ "$VAL" = 0.1.0 ] || return 1
}

# ---------------------------------------------------------------------------
# Claims.

# A file at path $1 in the working tree, reached without a symbolic link to a
# directory: Git stores no file under such a path. Git stores a symbolic link
# as a file whose content is the link's target, so for a link LINK is set to 1
# and TARGET to that content.
file_at() {
  local rest=$1 acc='' seg
  LINK=0
  TARGET=
  while :; do
    case $rest in
      */*) seg=${rest%%/*} rest=${rest#*/} ;;
      *) seg=$rest rest= ;;
    esac
    acc=${acc:+$acc/}$seg
    if [ -n "$rest" ]; then
      [ -d "$acc" ] && [ ! -L "$acc" ] || return 1
    elif [ -L "$acc" ]; then
      TARGET=$(readlink "$acc") || return 1
      LINK=1
      return 0
    else
      [ -f "$acc" ]
      return
    fi
  done
}

sha256_of() { # hashes stdin; sets SHA; fails when the input cannot be read
  local out
  if command -v sha256sum >/dev/null 2>&1; then
    out=$(sha256sum 2>/dev/null) || return 1
  elif command -v shasum >/dev/null 2>&1; then
    out=$(shasum -a 256 2>/dev/null) || return 1
  else
    usage_error "no sha256sum or shasum found; cannot evaluate sha256 claims"
  fi
  SHA=${out%% *}
  hash_ok "$SHA" || usage_error "unexpected output from the hashing tool: $out"
}

# A claim is Confirmed or Refuted. A missing or unreadable file refutes it.
evaluate_claim() { # $1 claim index; sets CV (value), CD (description), WHY
  local type=${CL_TYPE[$1]} path=${CL_PATH[$1]} text=${CL_TEXT[$1]} hash=${CL_HASH[$1]}
  CD="$type $path"
  CV=Refuted
  WHY=

  if ! names_path "$path" || ! file_at "$path"; then
    WHY="no file at this path"
    return
  fi
  if [ "$LINK" -eq 0 ] && [ ! -r "$path" ]; then
    WHY="the file cannot be read"
    return
  fi

  case $type in
    contains)
      if [ "$LINK" -eq 1 ]; then
        case $TARGET in *"$text"*) CV=Confirmed ;; *) WHY="text not found" ;; esac
      elif [ -z "$text" ]; then
        CV=Confirmed
      else
        grep -q -F -e "$text" -- "$path" 2>/dev/null
        case $? in
          0) CV=Confirmed ;;
          1) WHY="text not found" ;;
          *) WHY="the file cannot be read" ;;
        esac
      fi
      ;;
    sha256)
      if [ "$LINK" -eq 1 ]; then
        sha256_of < <(printf '%s' "$TARGET")
      else
        sha256_of <"$path"
      fi || {
        WHY="the file cannot be read"
        return
      }
      if [ "$SHA" = "$hash" ]; then
        CV=Confirmed
      else
        WHY="SHA-256 is $SHA"
      fi
      ;;
  esac
}

# ---------------------------------------------------------------------------

main() {
  local report line i=0 refuted=0

  case $# in
    1) ;;
    *) usage_error "usage: verdict.sh breadcrumbs/<...>/<id>.task.md" ;;
  esac
  case $1 in
    -h | --help)
      sed -n '3,29s/^# \{0,1\}//p' "$0"
      exit 0
      ;;
  esac
  TASK=${1#./}
  check_task_file

  if ! verified; then
    printf '%s: specification 0.1.0 verifies only status IMPLEMENTED with spec_version "0.1.0"\n' "$TASK"
    printf '%s: not verified\n' "$ID"
    exit 2
  fi

  # task.sh stops at the first structural violation, so run it in a subshell.
  # shellcheck disable=SC2034 # task.sh collects violations in ERRORS
  ERRORS=''
  if ! report=$(check_task "$TASK"; finish); then
    while IFS= read -r line; do
      printf 'verdict.sh: %s: %s\n' "$TASK" "${line#Error: }" >&2
    done <<EOF
$report
EOF
    exit 3
  fi
  # The file is well-formed: read it again here, to keep the claims.
  # shellcheck disable=SC2034
  ERRORS=''
  check_task "$TASK"

  printf '%s: IMPLEMENTED, %d claim(s)\n' "$TASK" "$CL_N"
  while [ "$i" -lt "$CL_N" ]; do
    evaluate_claim "$i"
    printf '  %-9s %s%s\n' "$CV" "$CD" "${WHY:+: $WHY}"
    [ "$CV" = Refuted ] && refuted=$((refuted + 1))
    i=$((i + 1))
  done

  if [ "$refuted" -gt 0 ]; then
    printf '%s: Refuted\n' "$ID"
    exit 1
  fi
  printf '%s: Confirmed\n' "$ID"
  exit 0
}

main "$@"
