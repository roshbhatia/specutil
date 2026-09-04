#compdef specutil
__specutil_completion_values_0() {
  local -a values
  values=()
  values+=("${(@f)$('specutil' '__values' 'changes' "${BUFFER[1,CURSOR]}" 2>/dev/null)}")
  compadd -a values
}
__specutil_completion_values_1() {
  local -a values
  values=()
  values+=("${(@f)$('specutil' '__values' 'changes' "${BUFFER[1,CURSOR]}" 2>/dev/null)}")
  compadd -a values
}
__specutil_completion_values_2() {
  local -a values
  values=()
  values+=("${(@f)$('specutil' '__values' 'providers' 2>/dev/null)}")
  compadd -a values
}
__specutil_completion_values_3() {
  local -a values
  values=()
  values+=("${(@f)$('specutil' '__values' 'changes' "${BUFFER[1,CURSOR]}" 2>/dev/null)}")
  compadd -a values
}
__specutil_completion_values_4() {
  local -a values
  values=()
  values+=("${(@f)$('specutil' '__values' 'changes' "${BUFFER[1,CURSOR]}" 2>/dev/null)}")
  compadd -a values
}
__specutil_completion_values_5() {
  local -a values
  values=()
  values+=("${(@f)$('specutil' '__values' 'providers' 2>/dev/null)}")
  compadd -a values
}
__specutil_completion_values_6() {
  local -a values
  values=()
  values+=("${(@f)$('specutil' '__values' 'changes' "${BUFFER[1,CURSOR]}" 2>/dev/null)}")
  compadd -a values
}
__specutil_completion_values_7() {
  local -a values
  values=()
  values+=("${(@f)$('specutil' '__values' 'changes' "${BUFFER[1,CURSOR]}" 2>/dev/null)}")
  compadd -a values
}
__specutil_completion_values_8() {
  local -a values
  values=()
  values+=("${(@f)$('specutil' '__values' 'changes' "${BUFFER[1,CURSOR]}" 2>/dev/null)}")
  compadd -a values
}
__specutil_completion_values_9() {
  local -a values
  values=()
  values+=("${(@f)$('specutil' '__values' 'changes' "${BUFFER[1,CURSOR]}" 2>/dev/null)}")
  compadd -a values
}
__specutil_completion_values_10() {
  local -a values
  values=()
  values+=("${(@f)$('specutil' '__values' 'changes' "${BUFFER[1,CURSOR]}" 2>/dev/null)}")
  compadd -a values
}
__specutil_completion_values_11() {
  local -a values
  values=()
  values+=("${(@f)$('specutil' '__values' 'changes' "${BUFFER[1,CURSOR]}" 2>/dev/null)}")
  compadd -a values
}
__specutil_completion_values_12() {
  local -a values
  values=()
  values+=("${(@f)$('specutil' '__values' 'changes' "${BUFFER[1,CURSOR]}" 2>/dev/null)}")
  compadd -a values
}
__specutil_completion_values_13() {
  local -a values
  values=()
  values+=("${(@f)$('specutil' '__values' 'changes' "${BUFFER[1,CURSOR]}" 2>/dev/null)}")
  compadd -a values
}
__specutil_completion_values_14() {
  local -a values
  values=()
  values+=("${(@f)$('specutil' '__values' 'changes' "${BUFFER[1,CURSOR]}" 2>/dev/null)}")
  compadd -a values
}
__specutil_completion_values_15() {
  local -a values
  values=()
  values+=("${(@f)$('specutil' '__values' 'changes' "${BUFFER[1,CURSOR]}" 2>/dev/null)}")
  compadd -a values
}

_specutil() {
  local context=''
  local word
  local consume_value=0
  local options_done=0
  for word in ${words[2,$((CURRENT - 1))]}; do
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
  case "$context" in
    '')
      _arguments \
        '(-C)--repo[repository root containing the openspec/ directory]:value:' \
        '1:command:(completion check config generate graph next provider render review web)'

      ;;
    'completion')
      _arguments \
        '2:shell:(bash zsh fish nu)'
      ;;
    'check')
      _arguments \
        '--as[output format: text|json]:value:' \
        '--change[check a single change (or pass as positional arg)]:value:__specutil_completion_values_1' \
        '--list-rules[list every built-in rule and exit]' \
        '(-o)--out[write output to a file instead of stdout]:value:' \
        '(-C)--repo[repository root containing the openspec/ directory]:value:' \
        '*:argument:__specutil_completion_values_0'

      ;;
    'config')
      _arguments \
        '(-C)--repo[repository root containing the openspec/ directory]:value:' \
        '2:command:(schema)'

      ;;
    'config schema')
      _arguments \
        '(-C)--repo[repository root containing the openspec/ directory]:value:' \
        '*:argument:'

      ;;
    'generate')
      _arguments \
        '--check[fail when a generated artifact is stale]' \
        '(-h)--help[help for generate]' \
        '(-C)--repo[repository root containing the openspec/ directory]:value:' \
        '*:argument:'

      ;;
    'graph')
      _arguments \
        '--as[output format: json|mermaid|dot|detail]:value:' \
        '--command[executable passed to the optional command provider]:value:' \
        '(-o)--out[write output to a file instead of stdout]:value:' \
        '--provider[external suggestion provider (default: heuristic only)]:value:__specutil_completion_values_2' \
        '--suggest[infer candidate edges from shared capabilities (read-only)]' \
        '(-C)--repo[repository root containing the openspec/ directory]:value:' \
        '*:argument:'

      ;;
    'next')
      _arguments \
        '--as[output format: text|json]:value:' \
        '--change[report a single change (or pass as positional arg)]:value:__specutil_completion_values_4' \
        '(-o)--out[write output to a file instead of stdout]:value:' \
        '(-C)--repo[repository root containing the openspec/ directory]:value:' \
        '*:argument:__specutil_completion_values_3'

      ;;
    'provider')
      _arguments \
        '(-C)--repo[repository root containing the openspec/ directory]:value:' \
        '2:command:(list validate)'

      ;;
    'provider list')
      _arguments \
        '--json[emit provider metadata as JSON]' \
        '(-C)--repo[repository root containing the openspec/ directory]:value:' \
        '*:argument:'

      ;;
    'provider validate')
      _arguments \
        '--json[emit validation reports as JSON]' \
        '(-C)--repo[repository root containing the openspec/ directory]:value:' \
        '*:argument:__specutil_completion_values_5'

      ;;
    'render')
      _arguments \
        '--as[target format: rfc|design|tickets (required)]:value:' \
        '--change[change name to render (or pass as positional arg)]:value:__specutil_completion_values_7' \
        '(-o)--out[write output to a file instead of stdout]:value:' \
        '--templates[override built-in template directory]:value:' \
        '(-C)--repo[repository root containing the openspec/ directory]:value:' \
        '*:argument:__specutil_completion_values_6'

      ;;
    'review')
      _arguments \
        '(-C)--repo[repository root containing the openspec/ directory]:value:' \
        '2:command:(diff ingest set show)'

      ;;
    'review diff')
      _arguments \
        '--as[output format: text|json]:value:' \
        '--base[git ref to compare against (default: the reviewed commit, else HEAD)]:value:' \
        '--change[change whose review supplies the base (or pass as positional arg)]:value:__specutil_completion_values_9' \
        '(-o)--out[write output to a file instead of stdout]:value:' \
        '--path[restrict the diff to these paths]:value:' \
        '--spec-only[restrict the diff to the changes own artifact directory]' \
        '(-C)--repo[repository root containing the openspec/ directory]:value:' \
        '*:argument:__specutil_completion_values_8'

      ;;
    'review ingest')
      _arguments \
        '--change[override the change named in the feedback document]:value:__specutil_completion_values_10' \
        '--dry-run[print the brief without writing the record]' \
        '(-o)--out[write output to a file instead of stdout]:value:' \
        '(-C)--repo[repository root containing the openspec/ directory]:value:' \
        '*:argument:'

      ;;
    'review set')
      _arguments \
        '--change[change to record against (or pass as positional arg)]:value:__specutil_completion_values_12' \
        '--clear-comments[drop the task comments carried in the record]' \
        '--decision[approved|changes-requested|commented (required)]:value:' \
        '--note[note to record with the decision]:value:' \
        '(-o)--out[write output to a file instead of stdout]:value:' \
        '(-C)--repo[repository root containing the openspec/ directory]:value:' \
        '*:argument:__specutil_completion_values_11'

      ;;
    'review show')
      _arguments \
        '--as[output format: text|json]:value:' \
        '--change[change to report (or pass as positional arg)]:value:__specutil_completion_values_14' \
        '(-o)--out[write output to a file instead of stdout]:value:' \
        '(-C)--repo[repository root containing the openspec/ directory]:value:' \
        '*:argument:__specutil_completion_values_13'

      ;;
    'web')
      _arguments \
        '--base[git ref for --diff (default: the reviewed commit, else HEAD)]:value:' \
        '--change[change the --diff belongs to]:value:__specutil_completion_values_15' \
        '--diff[include the working-tree diff for annotation (requires a single change)]' \
        '--open[open the generated page in the default browser]' \
        '(-o)--out[output HTML file path (default: timestamped temp file; - for stdout)]:value:' \
        '(-C)--repo[repository root containing the openspec/ directory]:value:' \
        '*:argument:'

      ;;
  esac
}
compdef _specutil specutil
