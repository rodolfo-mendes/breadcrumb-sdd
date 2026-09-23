#!/bin/bash
#
# task.sh - check the syntax of one Breadcrumb SDD Task file.
#
# Implements Breadcrumb SDD specification 0.1.0. Where this tool and the
# specification disagree, the specification wins and this tool is wrong.
#
# A syntax validator: it checks FILE against Format and Claims, whatever its
# status or spec_version. It does not evaluate claims, and it does not check
# the file's name or location: whoever passes FILE says it is a Task file.
#
# Usage: task.sh [--] FILE
#
# FILE is read relative to the current directory.
#
# Output:
#   (nothing)          FILE does not exist or is empty (zero bytes)
#   Ok                 FILE follows the rules
#   Error: <message>   one line per violation found
#
# After a violation that breaks the structure (indentation, a tab, a line that
# is not "key: value"), the rest of the front matter cannot be read reliably,
# so checking stops there.
#
# Exit status: 0 nothing or Ok, 1 at least one violation, 2 usage error.
#
# verdict.sh sources this file to check a Task's format; sourced, it only
# defines functions and runs nothing.
#
# Needs Bash 3.2 or later (sed for --help).

LC_ALL=C
export LC_ALL

TAB=$(printf '\t')
CR=$(printf '\r')
# Characters YAML does not allow in a document: C0 controls other than tab,
# line feed and carriage return; DEL; and C1 controls other than NEL, as UTF-8.
CTRL=$(printf '\001\002\003\004\005\006\007\010\013\014\016\017\020\021\022\023\024\025\026\027\030\031\032\033\034\035\036\037\177')
C1_LEAD=$(printf '\302')
C1_TAIL=$(printf '\200\201\202\203\204\206\207\210\211\212\213\214\215\216\217\220\221\222\223\224\225\226\227\230\231\232\233\234\235\236\237')
NL='
'
ERRORS=''

usage_error() {
  printf 'task.sh: %s\n' "$1" >&2
  exit 2
}

# error records a violation and goes on; fatal records one and stops.
error() {
  ERRORS=$ERRORS"Error: $1$NL"
}

fatal() {
  error "$1"
  finish
}

finish() {
  if [ -z "$ERRORS" ]; then
    printf 'Ok\n'
    exit 0
  fi
  printf '%s' "$ERRORS"
  exit 1
}

trim_trailing() { # $1 -> TRIMMED
  TRIMMED=$1
  while :; do
    case $TRIMMED in
      *' ' | *"$TAB") TRIMMED=${TRIMMED%?} ;;
      *) break ;;
    esac
  done
}

trim_leading() { # $1 -> TRIMMED
  TRIMMED=$1
  while :; do
    case $TRIMMED in
      ' '* | "$TAB"*) TRIMMED=${TRIMMED#?} ;;
      *) break ;;
    esac
  done
}

# Remove a YAML comment: a "#" at the start or after whitespace, outside
# quotes. A quote opens a quoted scalar only at the start or after
# whitespace, so the apostrophe in "it's" is not a quote. Sets SC.
strip_comment() {
  local s=$1 n=${#1} i=0 c q='' prev=' '
  while [ "$i" -lt "$n" ]; do
    c=${s:i:1}
    if [ "$q" = '"' ]; then
      if [ "$c" = "\\" ]; then
        i=$((i + 2))
        prev=x
        continue
      fi
      [ "$c" = '"' ] && q=''
    elif [ "$q" = "'" ]; then
      if [ "$c" = "'" ]; then
        if [ "${s:i+1:1}" = "'" ]; then
          i=$((i + 2))
          continue
        fi
        q=''
      fi
    else
      case $c in
        '#')
          case $prev in
            ' ' | "$TAB")
              trim_trailing "${s:0:i}"
              SC=$TRIMMED
              return
              ;;
          esac
          ;;
        '"' | "'")
          case $prev in ' ' | "$TAB") q=$c ;; esac
          ;;
      esac
    fi
    prev=$c
    i=$((i + 1))
  done
  trim_trailing "$s"
  SC=$TRIMMED
}

# Front matter lines, without comments and blank lines.
# FM_IND = indentation in spaces, FM_TXT = content, FM_LNO = line number.
read_front_matter() { # $1 file; fails with the reason in FM_ERR
  local line n=0 rest indent
  FM_N=0
  while IFS= read -r line || [ -n "$line" ]; do
    n=$((n + 1))
    line=${line%"$CR"}
    if [ "$n" -eq 1 ]; then
      if [ "$line" != '---' ]; then
        FM_ERR="no front matter: line 1 must be ---"
        return 1
      fi
      continue
    fi
    [ "$line" = '---' ] && return 0
    case $line in
      *["$CTRL"]* | *"$C1_LEAD"["$C1_TAIL"]*)
        error "line $n: control character (YAML does not allow it)"
        continue
        ;;
      *"$CR"*)
        error "line $n: carriage return inside a line (YAML reads it as a line break)"
        continue
        ;;
    esac
    rest=${line#"${line%%[! ]*}"}
    indent=$((${#line} - ${#rest}))
    strip_comment "$rest"
    [ -n "$SC" ] || continue
    FM_IND[FM_N]=$indent
    FM_TXT[FM_N]=$SC
    FM_LNO[FM_N]=$n
    FM_N=$((FM_N + 1))
  done <"$1"
  if [ "$n" -eq 0 ]; then
    FM_ERR="the file is empty; it must begin with front matter"
  else
    FM_ERR="front matter has no closing --- line"
  fi
  return 1
}

# ---------------------------------------------------------------------------
# Scalars.

decode_dq() { # sets DQ; fails unless $1 is one double-quoted string of the subset
  local s=$1 n i c out=''
  n=${#s}
  [ "$n" -ge 2 ] || return 1
  [ "${s:0:1}" = '"' ] && [ "${s:n-1:1}" = '"' ] || return 1
  i=1
  while [ "$i" -lt $((n - 1)) ]; do
    c=${s:i:1}
    case $c in
      "\\")
        i=$((i + 1))
        [ "$i" -lt $((n - 1)) ] || return 1
        c=${s:i:1}
        case $c in
          '"' | "\\") out=$out$c ;;
          *) return 1 ;;
        esac
        ;;
      '"') return 1 ;;
      *) out=$out$c ;;
    esac
    i=$((i + 1))
  done
  DQ=$out
}

decode_sq() { # sets SQ; fails unless $1 is one single-quoted string
  local s=$1 n i c out=''
  n=${#s}
  [ "$n" -ge 2 ] || return 1
  [ "${s:0:1}" = "'" ] && [ "${s:n-1:1}" = "'" ] || return 1
  i=1
  while [ "$i" -lt $((n - 1)) ]; do
    c=${s:i:1}
    if [ "$c" = "'" ]; then
      i=$((i + 1))
      [ "$i" -lt $((n - 1)) ] && [ "${s:i:1}" = "'" ] || return 1
    fi
    out=$out$c
    i=$((i + 1))
  done
  SQ=$out
}

scan_quoted() { # sets QLEN, the length of the quoted token $1 starts with
  local t=$1 n=${#1} i=1 q=${1:0:1} c
  while [ "$i" -lt "$n" ]; do
    c=${t:i:1}
    if [ "$q" = '"' ]; then
      case $c in
        "\\")
          i=$((i + 2))
          continue
          ;;
        '"')
          QLEN=$((i + 1))
          return 0
          ;;
      esac
    elif [ "$c" = "'" ]; then
      if [ "${t:i+1:1}" = "'" ]; then
        i=$((i + 2))
        continue
      fi
      QLEN=$((i + 1))
      return 0
    fi
    i=$((i + 1))
  done
  return 1
}

# ---------------------------------------------------------------------------
# Block structure.

is_seq_item() {
  case $1 in '-' | '- '* | "-$TAB"*) return 0 ;; esac
  return 1
}

split_entry() { # "key: value" -> KEY (resolved), RAWKEY, REST; fails if not an entry
  local t=$1 k rest a b
  case $t in
    '"'* | "'"*)
      scan_quoted "$t" || return 1
      k=${t:0:QLEN}
      trim_leading "${t:QLEN}"
      rest=$TRIMMED
      case $rest in
        ':') REST='' ;;
        ': '* | ":$TAB"*) REST=${rest:2} ;;
        *) return 1 ;;
      esac
      RAWKEY=$k
      if [ "${k:0:1}" = '"' ]; then
        if decode_dq "$k"; then KEY=$DQ; else KEY=$k; fi
      else
        if decode_sq "$k"; then KEY=$SQ; else KEY=$k; fi
      fi
      ;;
    *)
      a=${t%%": "*}
      b=${t%%":$TAB"*}
      [ "${#b}" -lt "${#a}" ] && a=$b
      if [ "$a" != "$t" ]; then
        k=$a
        REST=${t:${#a}+2}
      else
        case $t in
          *':')
            k=${t%:}
            REST=''
            ;;
          *) return 1 ;;
        esac
      fi
      trim_trailing "$k"
      KEY=$TRIMMED
      RAWKEY=$KEY
      [ -n "$KEY" ] || return 1
      ;;
  esac
  trim_leading "$REST"
  REST=$TRIMMED
}

tab_check() { # the line at P must not be indented with a tab
  case ${FM_TXT[P]} in
    "$TAB"*) fatal "line ${FM_LNO[P]}: tab in indentation" ;;
  esac
}

no_deeper() { # $1 indent; the next line must not be indented deeper
  if [ "$P" -lt "$FM_N" ] && [ "${FM_IND[P]}" -gt "$1" ]; then
    fatal "line ${FM_LNO[P]}: unexpected indentation"
  fi
}


# A value that may continue on the lines below: nothing (nested content), a
# block scalar, a flow collection, or a quoted scalar not closed on its line.
multiline_value() {
  case $1 in
    '' | '|'* | '>'* | '['* | '{'*) return 0 ;;
    '"'* | "'"*) scan_quoted "$1" && return 1 ;;
    *) return 1 ;;
  esac
  return 0
}

# After an invalid value: skip what belongs to it, or reject a stray line.
after_invalid() { # $1 value text, $2 indent of its key
  if multiline_value "$1"; then
    skip_deeper "$2"
  else
    no_deeper "$2"
  fi
}

skip_deeper() { # $1 indent; skip the lines indented deeper
  while [ "$P" -lt "$FM_N" ] && [ "${FM_IND[P]}" -gt "$1" ]; do
    P=$((P + 1))
  done
}

skip_unknown() { # $1 top indent, $2 value text: skip an ignored key's content
  while [ "$P" -lt "$FM_N" ]; do
    if [ "${FM_IND[P]}" -gt "$1" ]; then
      P=$((P + 1))
    elif [ -z "$2" ] && [ "${FM_IND[P]}" -eq "$1" ] && is_seq_item "${FM_TXT[P]}"; then
      P=$((P + 1))
    else
      break
    fi
  done
}

# ---------------------------------------------------------------------------
# Format.

check_status() { # $1 value text, $2 line; the value itself is not judged
  case $1 in
    '') error "line $2: status has no value" ;;
    '&'* | '!'* | '*'*) error "line $2: status must be a plain scalar, without anchor, alias or tag" ;;
    '"'* | "'"*) error "line $2: status must be written as a plain scalar, not quoted" ;;
    '['* | '{'* | '|'* | '>'*) error "line $2: status must be a plain scalar" ;;
    *) return 0 ;;
  esac
  after_invalid "$1" "$TOP"
}

check_spec_version() { # $1 value text, $2 line; the value itself is not judged
  case $1 in
    '') error "line $2: spec_version has no value" ;;
    '"'*)
      if ! scan_quoted "$1" || [ "$QLEN" -ne "${#1}" ]; then
        error "line $2: spec_version is not a double-quoted string on one line"
      elif ! decode_dq "$1"; then
        error "line $2: spec_version uses an escape other than \\\" or \\\\"
      else
        return 0
      fi
      ;;
    *) error "line $2: spec_version must be written as a double-quoted string" ;;
  esac
  after_invalid "$1" "$TOP"
}

check_claims() { # $1 value text of "claims:", $2 line
  local si i=0 t content pad lno
  case $1 in
    '') ;;
    '&'* | '!'* | '*'*) error "line $2: claims: anchors, aliases and tags are not allowed" ;;
    '['*) error "line $2: claims must be a block sequence, not a flow sequence" ;;
    *) error "line $2: claims must be a block sequence" ;;
  esac
  if [ -n "$1" ]; then
    skip_unknown "$TOP" ""
    return
  fi
  if [ "$P" -lt "$FM_N" ] && is_seq_item "${FM_TXT[P]}" &&
    [ "${FM_IND[P]}" -ge "$TOP" ]; then
    si=${FM_IND[P]}
  elif [ "$P" -lt "$FM_N" ] && [ "${FM_IND[P]}" -gt "$TOP" ]; then
    error "line $2: claims must be a block sequence"
    skip_unknown "$TOP" ""
    return
  else
    return 0 # no value: the Task has no claims
  fi

  while [ "$P" -lt "$FM_N" ] && [ "${FM_IND[P]}" -eq "$si" ] && is_seq_item "${FM_TXT[P]}"; do
    i=$((i + 1))
    t=${FM_TXT[P]}
    lno=${FM_LNO[P]}
    if [ "$t" = '-' ]; then
      P=$((P + 1))
      if [ "$P" -lt "$FM_N" ] && [ "${FM_IND[P]}" -gt "$si" ] &&
        ! is_seq_item "${FM_TXT[P]}" && split_entry "${FM_TXT[P]}"; then
        check_claim "${FM_IND[P]}" "$i" "$si"
      else
        error "line $lno: claim $i is not a block mapping"
        skip_deeper "$si"
      fi
    else
      content=${t#-}
      pad=${content%%[! ]*}
      content=${content#"$pad"}
      case $content in
        '{'* | '['*) content='' ;;
      esac
      if [ -n "$pad" ] && [ -n "$content" ] && ! is_seq_item "$content" &&
        split_entry "$content"; then
        FM_IND[P]=$((si + 1 + ${#pad}))
        FM_TXT[P]=$content
        check_claim "${FM_IND[P]}" "$i" "$si"
      else
        error "line $lno: claim $i is not a block mapping"
        P=$((P + 1))
        skip_deeper "$si"
      fi
    fi
  done
}

path_ok() { # does not begin with "/", no "." or ".." segment
  case $1 in /*) return 1 ;; esac
  case "/$1/" in */./* | */../*) return 1 ;; esac
  return 0
}

hash_ok() { # 64 lowercase hexadecimal digits
  [ "${#1}" -eq 64 ] || return 1
  case $1 in *[!0123456789abcdef]*) return 1 ;; esac
  return 0
}

show() { # a claim value for a message: tabs and newlines made visible
  local v=${1//$TAB/\\t}
  printf '%s' "${v//$NL/\\n}"
}

# Key states: 0 absent, 1 a valid double-quoted value, 2 present but invalid.
# The decoded values of each claim are kept, by claim number minus one, in
# CL_TYPE, CL_PATH, CL_TEXT and CL_HASH; CL_N counts the claims. They are
# meaningful only when the file has no violation.
CL_N=0

check_claim() { # $1 indent of the claim's keys, $2 claim number, $3 indent of "-"
  local ci=$1 n=$2 si=$3 seen=$NL key lno first=${FM_LNO[P]} ok
  local s_type=0 s_path=0 s_text=0 s_hash=0 type='' path='' text='' hash=''
  while [ "$P" -lt "$FM_N" ] && [ "${FM_IND[P]}" -eq "$ci" ]; do
    tab_check
    lno=${FM_LNO[P]}
    if is_seq_item "${FM_TXT[P]}" || ! split_entry "${FM_TXT[P]}"; then
      fatal "line $lno: claim $n: expected key: value"
    fi
    key=$KEY
    P=$((P + 1))
    case $RAWKEY in
      '&'* | '!'* | '*'* | '?'*)
        error "line $lno: claim $n: anchors, aliases, tags and complex keys are not allowed"
        skip_deeper "$ci"
        continue
        ;;
    esac
    case $seen in
      *"$NL$key$NL"*) error "line $lno: claim $n: duplicate key '$(show "$key")'" ;;
    esac
    seen=$seen$key$NL
    ok=0
    case $REST in
      '') error "line $lno: claim $n: '$(show "$key")' has no value; every value in a claim is a double-quoted string" ;;
      '&'* | '!'* | '*'*) error "line $lno: claim $n: anchors, aliases and tags are not allowed" ;;
      '"'*)
        if ! scan_quoted "$REST" || [ "$QLEN" -ne "${#REST}" ]; then
          error "line $lno: claim $n: '$(show "$key")' is not a double-quoted string on one line"
        elif ! decode_dq "$REST"; then
          error "line $lno: claim $n: '$(show "$key")' uses an escape other than \\\" or \\\\"
        else
          ok=1
        fi
        ;;
      *) error "line $lno: claim $n: '$(show "$key")' must be a double-quoted string" ;;
    esac
    if [ "$ok" -eq 1 ]; then
      case $key in
        type) s_type=1 type=$DQ ;;
        path) s_path=1 path=$DQ ;;
        text) s_text=1 text=$DQ ;;
        hash) s_hash=1 hash=$DQ ;;
      esac
      no_deeper "$ci"
    else
      case $key in
        type) s_type=2 ;;
        path) s_path=2 ;;
        text) s_text=2 ;;
        hash) s_hash=2 ;;
      esac
      after_invalid "$REST" "$ci"
    fi
  done
  # A line that ends the claim early is an indentation error, not a missing key.
  if [ "$P" -lt "$FM_N" ]; then
    tab_check
    [ "${FM_IND[P]}" -gt "$si" ] && fatal "line ${FM_LNO[P]}: unexpected indentation"
  fi

  # shellcheck disable=SC2034 # read by verdict.sh
  CL_TYPE[n - 1]=$type
  # shellcheck disable=SC2034
  CL_PATH[n - 1]=$path
  # shellcheck disable=SC2034
  CL_TEXT[n - 1]=$text
  # shellcheck disable=SC2034
  CL_HASH[n - 1]=$hash
  CL_N=$n

  [ "$s_type" -eq 0 ] && error "line $first: claim $n has no type"
  [ "$s_path" -eq 0 ] && error "line $first: claim $n has no path"
  if [ "$s_path" -eq 1 ] && ! path_ok "$path"; then
    error "line $first: claim $n: '$(show "$path")' is not a path"
  fi
  [ "$s_type" -eq 1 ] || return 0
  case $type in
    contains)
      [ "$s_text" -eq 0 ] && error "line $first: claim $n: a contains claim requires text"
      ;;
    sha256)
      if [ "$s_hash" -eq 0 ]; then
        error "line $first: claim $n: a sha256 claim requires hash"
      elif [ "$s_hash" -eq 1 ] && ! hash_ok "$hash"; then
        error "line $first: claim $n: hash must be 64 lowercase hexadecimal digits"
      fi
      ;;
    *) error "line $first: claim $n: '$(show "$type")' is not a claim type" ;;
  esac
  return 0
}

check_task() { # $1 file
  local seen=$NL lno
  # shellcheck disable=SC2034 # read by verdict.sh
  CL_N=0
  read_front_matter "$1" || fatal "$FM_ERR"
  [ "$FM_N" -gt 0 ] || fatal "front matter is empty"

  TOP=${FM_IND[0]}
  P=0
  while [ "$P" -lt "$FM_N" ]; do
    [ "${FM_IND[P]}" -eq "$TOP" ] || fatal "line ${FM_LNO[P]}: unexpected indentation"
    tab_check
    lno=${FM_LNO[P]}
    is_seq_item "${FM_TXT[P]}" && fatal "line $lno: front matter must be a block mapping"
    split_entry "${FM_TXT[P]}" || fatal "line $lno: expected key: value"
    case $seen in *"$NL$KEY$NL"*) error "line $lno: duplicate key '$(show "$KEY")'" ;; esac
    seen=$seen$KEY$NL
    P=$((P + 1))
    case $KEY in
      status)
        check_status "$REST" "$lno"
        no_deeper "$TOP"
        ;;
      spec_version)
        check_spec_version "$REST" "$lno"
        no_deeper "$TOP"
        ;;
      claims) check_claims "$REST" "$lno" ;;
      *) skip_unknown "$TOP" "$REST" ;;
    esac
  done
  case $seen in *"${NL}status$NL"*) ;; *) error "status is missing" ;; esac
  case $seen in *"${NL}spec_version$NL"*) ;; *) error "spec_version is missing" ;; esac
}

main() {
  case $1 in
    -h | --help)
      sed -n '3,30s/^# \{0,1\}//p' "$0"
      exit 0
      ;;
    --) shift ;;
    -?*) usage_error "unknown option: $1 (usage: task.sh [--] FILE)" ;;
  esac
  [ "$#" -eq 1 ] || usage_error "usage: task.sh [--] FILE"
  [ -n "$1" ] || usage_error "usage: task.sh [--] FILE"

  # A missing or empty file is skipped: nothing to say about it.
  [ -e "$1" ] || exit 0
  [ -d "$1" ] && usage_error "$1 is a directory"
  [ -s "$1" ] || exit 0
  [ -r "$1" ] || usage_error "cannot read $1"

  check_task "$1"
  finish
}

# Sourced by verdict.sh: define the functions, run nothing.
if [ "${BASH_SOURCE[0]}" = "$0" ]; then
  main "$@"
fi
