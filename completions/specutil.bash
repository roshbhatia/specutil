__specutil_completion_values_0() {
  'specutil' '__values' 'changes' "${COMP_LINE:0:COMP_POINT}" 2>/dev/null || true
}
__specutil_completion_values_1() {
  'specutil' '__values' 'changes' "${COMP_LINE:0:COMP_POINT}" 2>/dev/null || true
}
__specutil_completion_values_2() {
  'specutil' '__values' 'providers' 2>/dev/null || true
}
__specutil_completion_values_3() {
  'specutil' '__values' 'changes' "${COMP_LINE:0:COMP_POINT}" 2>/dev/null || true
}
__specutil_completion_values_4() {
  'specutil' '__values' 'changes' "${COMP_LINE:0:COMP_POINT}" 2>/dev/null || true
}
__specutil_completion_values_5() {
  'specutil' '__values' 'providers' 2>/dev/null || true
}
__specutil_completion_values_6() {
  'specutil' '__values' 'changes' "${COMP_LINE:0:COMP_POINT}" 2>/dev/null || true
}
__specutil_completion_values_7() {
  'specutil' '__values' 'changes' "${COMP_LINE:0:COMP_POINT}" 2>/dev/null || true
}
__specutil_completion_values_8() {
  'specutil' '__values' 'changes' "${COMP_LINE:0:COMP_POINT}" 2>/dev/null || true
}
__specutil_completion_values_9() {
  'specutil' '__values' 'changes' "${COMP_LINE:0:COMP_POINT}" 2>/dev/null || true
}
__specutil_completion_values_10() {
  'specutil' '__values' 'changes' "${COMP_LINE:0:COMP_POINT}" 2>/dev/null || true
}
__specutil_completion_values_11() {
  'specutil' '__values' 'changes' "${COMP_LINE:0:COMP_POINT}" 2>/dev/null || true
}
__specutil_completion_values_12() {
  'specutil' '__values' 'changes' "${COMP_LINE:0:COMP_POINT}" 2>/dev/null || true
}
__specutil_completion_values_13() {
  'specutil' '__values' 'changes' "${COMP_LINE:0:COMP_POINT}" 2>/dev/null || true
}
__specutil_completion_values_14() {
  'specutil' '__values' 'changes' "${COMP_LINE:0:COMP_POINT}" 2>/dev/null || true
}
__specutil_completion_values_15() {
  'specutil' '__values' 'changes' "${COMP_LINE:0:COMP_POINT}" 2>/dev/null || true
}
__specutil_completion_filter() {
  local prefix="$1"
  local prepend="${2-}"
  local candidate
  local existing
  local duplicate
  COMPREPLY=()
  while IFS= read -r candidate || [[ -n "$candidate" ]]; do
    [[ "$candidate" == "$prefix"* ]] || continue
    candidate="$prepend$candidate"
    duplicate=0
    for existing in "${COMPREPLY[@]}"; do
      if [[ "$existing" == "$candidate" ]]; then
        duplicate=1
        break
      fi
    done
    (( duplicate )) || COMPREPLY+=("$candidate")
  done
}

_specutil_complete() {
  local current="${COMP_WORDS[COMP_CWORD]}"
  local previous=""
  local context=""
  local word
  local index
  local consume_value=0
  local options_done=0
  if (( COMP_CWORD > 0 )); then
    previous="${COMP_WORDS[COMP_CWORD-1]}"
  fi
  for ((index=1; index<COMP_CWORD; index++)); do
    word="${COMP_WORDS[index]}"
    if (( consume_value )); then
      consume_value=0
      continue
    fi
    if (( options_done )); then
      continue
    fi
    if [[ "$word" == '--' ]]; then
      options_done=1
      continue
    fi
    case "$context:$word" in
      ':--repo') consume_value=1; continue ;;
      ':--repo='*) continue ;;
      ':-C') consume_value=1; continue ;;
      ':-C='*) continue ;;
      'check:--as') consume_value=1; continue ;;
      'check:--as='*) continue ;;
      'check:--change') consume_value=1; continue ;;
      'check:--change='*) continue ;;
      'check:--out') consume_value=1; continue ;;
      'check:--out='*) continue ;;
      'check:-o') consume_value=1; continue ;;
      'check:-o='*) continue ;;
      'check:--repo') consume_value=1; continue ;;
      'check:--repo='*) continue ;;
      'check:-C') consume_value=1; continue ;;
      'check:-C='*) continue ;;
      'config:--repo') consume_value=1; continue ;;
      'config:--repo='*) continue ;;
      'config:-C') consume_value=1; continue ;;
      'config:-C='*) continue ;;
      'config schema:--repo') consume_value=1; continue ;;
      'config schema:--repo='*) continue ;;
      'config schema:-C') consume_value=1; continue ;;
      'config schema:-C='*) continue ;;
      'generate:--repo') consume_value=1; continue ;;
      'generate:--repo='*) continue ;;
      'generate:-C') consume_value=1; continue ;;
      'generate:-C='*) continue ;;
      'graph:--as') consume_value=1; continue ;;
      'graph:--as='*) continue ;;
      'graph:--command') consume_value=1; continue ;;
      'graph:--command='*) continue ;;
      'graph:--out') consume_value=1; continue ;;
      'graph:--out='*) continue ;;
      'graph:-o') consume_value=1; continue ;;
      'graph:-o='*) continue ;;
      'graph:--provider') consume_value=1; continue ;;
      'graph:--provider='*) continue ;;
      'graph:--repo') consume_value=1; continue ;;
      'graph:--repo='*) continue ;;
      'graph:-C') consume_value=1; continue ;;
      'graph:-C='*) continue ;;
      'next:--as') consume_value=1; continue ;;
      'next:--as='*) continue ;;
      'next:--change') consume_value=1; continue ;;
      'next:--change='*) continue ;;
      'next:--out') consume_value=1; continue ;;
      'next:--out='*) continue ;;
      'next:-o') consume_value=1; continue ;;
      'next:-o='*) continue ;;
      'next:--repo') consume_value=1; continue ;;
      'next:--repo='*) continue ;;
      'next:-C') consume_value=1; continue ;;
      'next:-C='*) continue ;;
      'provider:--repo') consume_value=1; continue ;;
      'provider:--repo='*) continue ;;
      'provider:-C') consume_value=1; continue ;;
      'provider:-C='*) continue ;;
      'provider list:--repo') consume_value=1; continue ;;
      'provider list:--repo='*) continue ;;
      'provider list:-C') consume_value=1; continue ;;
      'provider list:-C='*) continue ;;
      'provider validate:--repo') consume_value=1; continue ;;
      'provider validate:--repo='*) continue ;;
      'provider validate:-C') consume_value=1; continue ;;
      'provider validate:-C='*) continue ;;
      'render:--as') consume_value=1; continue ;;
      'render:--as='*) continue ;;
      'render:--change') consume_value=1; continue ;;
      'render:--change='*) continue ;;
      'render:--out') consume_value=1; continue ;;
      'render:--out='*) continue ;;
      'render:-o') consume_value=1; continue ;;
      'render:-o='*) continue ;;
      'render:--templates') consume_value=1; continue ;;
      'render:--templates='*) continue ;;
      'render:--repo') consume_value=1; continue ;;
      'render:--repo='*) continue ;;
      'render:-C') consume_value=1; continue ;;
      'render:-C='*) continue ;;
      'review:--repo') consume_value=1; continue ;;
      'review:--repo='*) continue ;;
      'review:-C') consume_value=1; continue ;;
      'review:-C='*) continue ;;
      'review diff:--as') consume_value=1; continue ;;
      'review diff:--as='*) continue ;;
      'review diff:--base') consume_value=1; continue ;;
      'review diff:--base='*) continue ;;
      'review diff:--change') consume_value=1; continue ;;
      'review diff:--change='*) continue ;;
      'review diff:--out') consume_value=1; continue ;;
      'review diff:--out='*) continue ;;
      'review diff:-o') consume_value=1; continue ;;
      'review diff:-o='*) continue ;;
      'review diff:--path') consume_value=1; continue ;;
      'review diff:--path='*) continue ;;
      'review diff:--repo') consume_value=1; continue ;;
      'review diff:--repo='*) continue ;;
      'review diff:-C') consume_value=1; continue ;;
      'review diff:-C='*) continue ;;
      'review ingest:--change') consume_value=1; continue ;;
      'review ingest:--change='*) continue ;;
      'review ingest:--out') consume_value=1; continue ;;
      'review ingest:--out='*) continue ;;
      'review ingest:-o') consume_value=1; continue ;;
      'review ingest:-o='*) continue ;;
      'review ingest:--repo') consume_value=1; continue ;;
      'review ingest:--repo='*) continue ;;
      'review ingest:-C') consume_value=1; continue ;;
      'review ingest:-C='*) continue ;;
      'review set:--change') consume_value=1; continue ;;
      'review set:--change='*) continue ;;
      'review set:--decision') consume_value=1; continue ;;
      'review set:--decision='*) continue ;;
      'review set:--note') consume_value=1; continue ;;
      'review set:--note='*) continue ;;
      'review set:--out') consume_value=1; continue ;;
      'review set:--out='*) continue ;;
      'review set:-o') consume_value=1; continue ;;
      'review set:-o='*) continue ;;
      'review set:--repo') consume_value=1; continue ;;
      'review set:--repo='*) continue ;;
      'review set:-C') consume_value=1; continue ;;
      'review set:-C='*) continue ;;
      'review show:--as') consume_value=1; continue ;;
      'review show:--as='*) continue ;;
      'review show:--change') consume_value=1; continue ;;
      'review show:--change='*) continue ;;
      'review show:--out') consume_value=1; continue ;;
      'review show:--out='*) continue ;;
      'review show:-o') consume_value=1; continue ;;
      'review show:-o='*) continue ;;
      'review show:--repo') consume_value=1; continue ;;
      'review show:--repo='*) continue ;;
      'review show:-C') consume_value=1; continue ;;
      'review show:-C='*) continue ;;
      'web:--base') consume_value=1; continue ;;
      'web:--base='*) continue ;;
      'web:--change') consume_value=1; continue ;;
      'web:--change='*) continue ;;
      'web:--out') consume_value=1; continue ;;
      'web:--out='*) continue ;;
      'web:-o') consume_value=1; continue ;;
      'web:-o='*) continue ;;
      'web:--repo') consume_value=1; continue ;;
      'web:--repo='*) continue ;;
      'web:-C') consume_value=1; continue ;;
      'web:-C='*) continue ;;
    esac
    case "$context:$word" in
      ':completion') context='completion' ;;
      ':check') context='check' ;;
      ':config') context='config' ;;
      'config:schema') context='config schema' ;;
      ':generate') context='generate' ;;
      ':graph') context='graph' ;;
      ':next') context='next' ;;
      ':provider') context='provider' ;;
      'provider:list') context='provider list' ;;
      'provider:validate') context='provider validate' ;;
      ':render') context='render' ;;
      ':review') context='review' ;;
      'review:diff') context='review diff' ;;
      'review:ingest') context='review ingest' ;;
      'review:set') context='review set' ;;
      'review:show') context='review show' ;;
      ':web') context='web' ;;
    esac
  done
  case "$context:$previous" in
    'check:--change') __specutil_completion_filter "$current" < <(__specutil_completion_values_1); return ;;
    'graph:--provider') __specutil_completion_filter "$current" < <(__specutil_completion_values_2); return ;;
    'next:--change') __specutil_completion_filter "$current" < <(__specutil_completion_values_4); return ;;
    'render:--change') __specutil_completion_filter "$current" < <(__specutil_completion_values_7); return ;;
    'review diff:--change') __specutil_completion_filter "$current" < <(__specutil_completion_values_9); return ;;
    'review ingest:--change') __specutil_completion_filter "$current" < <(__specutil_completion_values_10); return ;;
    'review set:--change') __specutil_completion_filter "$current" < <(__specutil_completion_values_12); return ;;
    'review show:--change') __specutil_completion_filter "$current" < <(__specutil_completion_values_14); return ;;
    'web:--change') __specutil_completion_filter "$current" < <(__specutil_completion_values_15); return ;;
  esac
  case "$context:$current" in
    'check:--change='*) __specutil_completion_filter "${current#*=}" "--change=" < <(__specutil_completion_values_1); return ;;
    'graph:--provider='*) __specutil_completion_filter "${current#*=}" "--provider=" < <(__specutil_completion_values_2); return ;;
    'next:--change='*) __specutil_completion_filter "${current#*=}" "--change=" < <(__specutil_completion_values_4); return ;;
    'render:--change='*) __specutil_completion_filter "${current#*=}" "--change=" < <(__specutil_completion_values_7); return ;;
    'review diff:--change='*) __specutil_completion_filter "${current#*=}" "--change=" < <(__specutil_completion_values_9); return ;;
    'review ingest:--change='*) __specutil_completion_filter "${current#*=}" "--change=" < <(__specutil_completion_values_10); return ;;
    'review set:--change='*) __specutil_completion_filter "${current#*=}" "--change=" < <(__specutil_completion_values_12); return ;;
    'review show:--change='*) __specutil_completion_filter "${current#*=}" "--change=" < <(__specutil_completion_values_14); return ;;
    'web:--change='*) __specutil_completion_filter "${current#*=}" "--change=" < <(__specutil_completion_values_15); return ;;
  esac
  case "$context" in
    '')
      __specutil_completion_filter "$current" < <(
        printf '%s\n' 'completion' 'check' 'config' 'generate' 'graph' 'next' 'provider' 'render' 'review' 'web' '--repo' '-C'
      )
      ;;
    'completion')
      __specutil_completion_filter "$current" < <(
        printf '%s\n' 'bash' 'zsh' 'fish' 'nu'
      )
      ;;
    'check')
      __specutil_completion_filter "$current" < <(
        printf '%s\n' '--as' '--change' '--list-rules' '--out' '-o' '--repo' '-C'
        __specutil_completion_values_0
      )
      ;;
    'config')
      __specutil_completion_filter "$current" < <(
        printf '%s\n' 'schema' '--repo' '-C'
      )
      ;;
    'config schema')
      __specutil_completion_filter "$current" < <(
        printf '%s\n' '--repo' '-C'
      )
      ;;
    'generate')
      __specutil_completion_filter "$current" < <(
        printf '%s\n' '--check' '--help' '-h' '--repo' '-C'
      )
      ;;
    'graph')
      __specutil_completion_filter "$current" < <(
        printf '%s\n' '--as' '--command' '--out' '-o' '--provider' '--suggest' '--repo' '-C'
      )
      ;;
    'next')
      __specutil_completion_filter "$current" < <(
        printf '%s\n' '--as' '--change' '--out' '-o' '--repo' '-C'
        __specutil_completion_values_3
      )
      ;;
    'provider')
      __specutil_completion_filter "$current" < <(
        printf '%s\n' 'list' 'validate' '--repo' '-C'
      )
      ;;
    'provider list')
      __specutil_completion_filter "$current" < <(
        printf '%s\n' '--json' '--repo' '-C'
      )
      ;;
    'provider validate')
      __specutil_completion_filter "$current" < <(
        printf '%s\n' '--json' '--repo' '-C'
        __specutil_completion_values_5
      )
      ;;
    'render')
      __specutil_completion_filter "$current" < <(
        printf '%s\n' '--as' '--change' '--out' '-o' '--templates' '--repo' '-C'
        __specutil_completion_values_6
      )
      ;;
    'review')
      __specutil_completion_filter "$current" < <(
        printf '%s\n' 'diff' 'ingest' 'set' 'show' '--repo' '-C'
      )
      ;;
    'review diff')
      __specutil_completion_filter "$current" < <(
        printf '%s\n' '--as' '--base' '--change' '--out' '-o' '--path' '--spec-only' '--repo' '-C'
        __specutil_completion_values_8
      )
      ;;
    'review ingest')
      __specutil_completion_filter "$current" < <(
        printf '%s\n' '--change' '--dry-run' '--out' '-o' '--repo' '-C'
      )
      ;;
    'review set')
      __specutil_completion_filter "$current" < <(
        printf '%s\n' '--change' '--clear-comments' '--decision' '--note' '--out' '-o' '--repo' '-C'
        __specutil_completion_values_11
      )
      ;;
    'review show')
      __specutil_completion_filter "$current" < <(
        printf '%s\n' '--as' '--change' '--out' '-o' '--repo' '-C'
        __specutil_completion_values_13
      )
      ;;
    'web')
      __specutil_completion_filter "$current" < <(
        printf '%s\n' '--base' '--change' '--diff' '--open' '--out' '-o' '--repo' '-C'
      )
      ;;
  esac
}
complete -F _specutil_complete specutil
