complete -c specutil -e
complete -c specutil -f
function __specutil_completion_values_0
  begin
    command 'specutil' '__values' 'changes' (commandline -cp) 2>/dev/null; or true
  end | string match -rv '\t'; or true
end
function __specutil_completion_values_1
  begin
    command 'specutil' '__values' 'changes' (commandline -cp) 2>/dev/null; or true
  end | string match -rv '\t'; or true
end
function __specutil_completion_values_2
  begin
    command 'specutil' '__values' 'providers' 2>/dev/null; or true
  end | string match -rv '\t'; or true
end
function __specutil_completion_values_3
  begin
    command 'specutil' '__values' 'changes' (commandline -cp) 2>/dev/null; or true
  end | string match -rv '\t'; or true
end
function __specutil_completion_values_4
  begin
    command 'specutil' '__values' 'changes' (commandline -cp) 2>/dev/null; or true
  end | string match -rv '\t'; or true
end
function __specutil_completion_values_5
  begin
    command 'specutil' '__values' 'providers' 2>/dev/null; or true
  end | string match -rv '\t'; or true
end
function __specutil_completion_values_6
  begin
    command 'specutil' '__values' 'changes' (commandline -cp) 2>/dev/null; or true
  end | string match -rv '\t'; or true
end
function __specutil_completion_values_7
  begin
    command 'specutil' '__values' 'changes' (commandline -cp) 2>/dev/null; or true
  end | string match -rv '\t'; or true
end
function __specutil_completion_values_8
  begin
    command 'specutil' '__values' 'changes' (commandline -cp) 2>/dev/null; or true
  end | string match -rv '\t'; or true
end
function __specutil_completion_values_9
  begin
    command 'specutil' '__values' 'changes' (commandline -cp) 2>/dev/null; or true
  end | string match -rv '\t'; or true
end
function __specutil_completion_values_10
  begin
    command 'specutil' '__values' 'changes' (commandline -cp) 2>/dev/null; or true
  end | string match -rv '\t'; or true
end
function __specutil_completion_values_11
  begin
    command 'specutil' '__values' 'changes' (commandline -cp) 2>/dev/null; or true
  end | string match -rv '\t'; or true
end
function __specutil_completion_values_12
  begin
    command 'specutil' '__values' 'changes' (commandline -cp) 2>/dev/null; or true
  end | string match -rv '\t'; or true
end
function __specutil_completion_values_13
  begin
    command 'specutil' '__values' 'changes' (commandline -cp) 2>/dev/null; or true
  end | string match -rv '\t'; or true
end
function __specutil_completion_values_14
  begin
    command 'specutil' '__values' 'changes' (commandline -cp) 2>/dev/null; or true
  end | string match -rv '\t'; or true
end
function __specutil_completion_values_15
  begin
    command 'specutil' '__values' 'changes' (commandline -cp) 2>/dev/null; or true
  end | string match -rv '\t'; or true
end

function __specutil_completion_context
  set -l context ''
  set -l words (commandline -opc)
  set -l consume_value 0
  set -l options_done 0
  for word in $words[2..-1]
    if test $consume_value -eq 1
      set consume_value 0
      continue
    end
    if test $options_done -eq 1
      continue
    end
    if test "$word" = '--'
      set options_done 1
      continue
    end
    switch "$context:$word"
      case ':--repo' ':-C'
        set consume_value 1
        continue
      case ':--repo=*'
        continue
      case ':-C=*'
        continue
      case 'check:--as'
        set consume_value 1
        continue
      case 'check:--as=*'
        continue
      case 'check:--change'
        set consume_value 1
        continue
      case 'check:--change=*'
        continue
      case 'check:--out' 'check:-o'
        set consume_value 1
        continue
      case 'check:--out=*'
        continue
      case 'check:-o=*'
        continue
      case 'check:--repo' 'check:-C'
        set consume_value 1
        continue
      case 'check:--repo=*'
        continue
      case 'check:-C=*'
        continue
      case 'config:--repo' 'config:-C'
        set consume_value 1
        continue
      case 'config:--repo=*'
        continue
      case 'config:-C=*'
        continue
      case 'config schema:--repo' 'config schema:-C'
        set consume_value 1
        continue
      case 'config schema:--repo=*'
        continue
      case 'config schema:-C=*'
        continue
      case 'generate:--repo' 'generate:-C'
        set consume_value 1
        continue
      case 'generate:--repo=*'
        continue
      case 'generate:-C=*'
        continue
      case 'graph:--as'
        set consume_value 1
        continue
      case 'graph:--as=*'
        continue
      case 'graph:--command'
        set consume_value 1
        continue
      case 'graph:--command=*'
        continue
      case 'graph:--out' 'graph:-o'
        set consume_value 1
        continue
      case 'graph:--out=*'
        continue
      case 'graph:-o=*'
        continue
      case 'graph:--provider'
        set consume_value 1
        continue
      case 'graph:--provider=*'
        continue
      case 'graph:--repo' 'graph:-C'
        set consume_value 1
        continue
      case 'graph:--repo=*'
        continue
      case 'graph:-C=*'
        continue
      case 'next:--as'
        set consume_value 1
        continue
      case 'next:--as=*'
        continue
      case 'next:--change'
        set consume_value 1
        continue
      case 'next:--change=*'
        continue
      case 'next:--out' 'next:-o'
        set consume_value 1
        continue
      case 'next:--out=*'
        continue
      case 'next:-o=*'
        continue
      case 'next:--repo' 'next:-C'
        set consume_value 1
        continue
      case 'next:--repo=*'
        continue
      case 'next:-C=*'
        continue
      case 'provider:--repo' 'provider:-C'
        set consume_value 1
        continue
      case 'provider:--repo=*'
        continue
      case 'provider:-C=*'
        continue
      case 'provider list:--repo' 'provider list:-C'
        set consume_value 1
        continue
      case 'provider list:--repo=*'
        continue
      case 'provider list:-C=*'
        continue
      case 'provider validate:--repo' 'provider validate:-C'
        set consume_value 1
        continue
      case 'provider validate:--repo=*'
        continue
      case 'provider validate:-C=*'
        continue
      case 'render:--as'
        set consume_value 1
        continue
      case 'render:--as=*'
        continue
      case 'render:--change'
        set consume_value 1
        continue
      case 'render:--change=*'
        continue
      case 'render:--out' 'render:-o'
        set consume_value 1
        continue
      case 'render:--out=*'
        continue
      case 'render:-o=*'
        continue
      case 'render:--templates'
        set consume_value 1
        continue
      case 'render:--templates=*'
        continue
      case 'render:--repo' 'render:-C'
        set consume_value 1
        continue
      case 'render:--repo=*'
        continue
      case 'render:-C=*'
        continue
      case 'review:--repo' 'review:-C'
        set consume_value 1
        continue
      case 'review:--repo=*'
        continue
      case 'review:-C=*'
        continue
      case 'review diff:--as'
        set consume_value 1
        continue
      case 'review diff:--as=*'
        continue
      case 'review diff:--base'
        set consume_value 1
        continue
      case 'review diff:--base=*'
        continue
      case 'review diff:--change'
        set consume_value 1
        continue
      case 'review diff:--change=*'
        continue
      case 'review diff:--out' 'review diff:-o'
        set consume_value 1
        continue
      case 'review diff:--out=*'
        continue
      case 'review diff:-o=*'
        continue
      case 'review diff:--path'
        set consume_value 1
        continue
      case 'review diff:--path=*'
        continue
      case 'review diff:--repo' 'review diff:-C'
        set consume_value 1
        continue
      case 'review diff:--repo=*'
        continue
      case 'review diff:-C=*'
        continue
      case 'review ingest:--change'
        set consume_value 1
        continue
      case 'review ingest:--change=*'
        continue
      case 'review ingest:--out' 'review ingest:-o'
        set consume_value 1
        continue
      case 'review ingest:--out=*'
        continue
      case 'review ingest:-o=*'
        continue
      case 'review ingest:--repo' 'review ingest:-C'
        set consume_value 1
        continue
      case 'review ingest:--repo=*'
        continue
      case 'review ingest:-C=*'
        continue
      case 'review set:--change'
        set consume_value 1
        continue
      case 'review set:--change=*'
        continue
      case 'review set:--decision'
        set consume_value 1
        continue
      case 'review set:--decision=*'
        continue
      case 'review set:--note'
        set consume_value 1
        continue
      case 'review set:--note=*'
        continue
      case 'review set:--out' 'review set:-o'
        set consume_value 1
        continue
      case 'review set:--out=*'
        continue
      case 'review set:-o=*'
        continue
      case 'review set:--repo' 'review set:-C'
        set consume_value 1
        continue
      case 'review set:--repo=*'
        continue
      case 'review set:-C=*'
        continue
      case 'review show:--as'
        set consume_value 1
        continue
      case 'review show:--as=*'
        continue
      case 'review show:--change'
        set consume_value 1
        continue
      case 'review show:--change=*'
        continue
      case 'review show:--out' 'review show:-o'
        set consume_value 1
        continue
      case 'review show:--out=*'
        continue
      case 'review show:-o=*'
        continue
      case 'review show:--repo' 'review show:-C'
        set consume_value 1
        continue
      case 'review show:--repo=*'
        continue
      case 'review show:-C=*'
        continue
      case 'web:--base'
        set consume_value 1
        continue
      case 'web:--base=*'
        continue
      case 'web:--change'
        set consume_value 1
        continue
      case 'web:--change=*'
        continue
      case 'web:--out' 'web:-o'
        set consume_value 1
        continue
      case 'web:--out=*'
        continue
      case 'web:-o=*'
        continue
      case 'web:--repo' 'web:-C'
        set consume_value 1
        continue
      case 'web:--repo=*'
        continue
      case 'web:-C=*'
        continue
    end
    switch "$context:$word"
      case ':completion'
        set context 'completion'
      case ':check'
        set context 'check'
      case ':config'
        set context 'config'
      case 'config:schema'
        set context 'config schema'
      case ':generate'
        set context 'generate'
      case ':graph'
        set context 'graph'
      case ':next'
        set context 'next'
      case ':provider'
        set context 'provider'
      case 'provider:list'
        set context 'provider list'
      case 'provider:validate'
        set context 'provider validate'
      case ':render'
        set context 'render'
      case ':review'
        set context 'review'
      case 'review:diff'
        set context 'review diff'
      case 'review:ingest'
        set context 'review ingest'
      case 'review:set'
        set context 'review set'
      case 'review:show'
        set context 'review show'
      case ':web'
        set context 'web'
    end
  end
  echo $context
end
complete -c specutil -n 'test (__specutil_completion_context) = ""' -l repo -s C -r -d 'repository root containing the openspec/ directory'
complete -c specutil -f -n 'test (__specutil_completion_context) = ""' -a completion -d 'Generate shell completions'
complete -c specutil -f -n 'test (__specutil_completion_context) = ""' -a check -d 'Validate changes against the rubric declared in specutil.yaml'
complete -c specutil -f -n 'test (__specutil_completion_context) = ""' -a config -d 'Inspect project configuration support'
complete -c specutil -f -n 'test (__specutil_completion_context) = ""' -a generate -d 'Regenerate artifacts in the specutil source repository'
complete -c specutil -f -n 'test (__specutil_completion_context) = ""' -a graph -d 'Output the cross-change dependency graph in various formats'
complete -c specutil -f -n 'test (__specutil_completion_context) = ""' -a next -d 'Report which subtasks are runnable now'
complete -c specutil -f -n 'test (__specutil_completion_context) = ""' -a provider -d 'Inspect optional suggestion providers'
complete -c specutil -f -n 'test (__specutil_completion_context) = ""' -a render -d 'Render a change into a shareable document (rfc|design|tickets)'
complete -c specutil -f -n 'test (__specutil_completion_context) = ""' -a review -d 'Record a human verdict on a change and report what moved since'
complete -c specutil -f -n 'test (__specutil_completion_context) = ""' -a web -d 'Open a browser view of the change board, dependency graph, and task details'
complete -c specutil -f -n 'test (__specutil_completion_context) = "completion"' -a 'bash zsh fish nu'
complete -c specutil -n 'test (__specutil_completion_context) = "check"' -l as -r -d 'output format: text|json'
complete -c specutil -n 'test (__specutil_completion_context) = "check"' -f -l change -r -a '(__specutil_completion_values_1)' -d 'check a single change (or pass as positional arg)'
complete -c specutil -n 'test (__specutil_completion_context) = "check"' -l list-rules -d 'list every built-in rule and exit'
complete -c specutil -n 'test (__specutil_completion_context) = "check"' -l out -s o -r -d 'write output to a file instead of stdout'
complete -c specutil -n 'test (__specutil_completion_context) = "check"' -l repo -s C -r -d 'repository root containing the openspec/ directory'
complete -c specutil -f -n 'test (__specutil_completion_context) = "check"' -a '(__specutil_completion_values_0)'
complete -c specutil -n 'test (__specutil_completion_context) = "config"' -l repo -s C -r -d 'repository root containing the openspec/ directory'
complete -c specutil -f -n 'test (__specutil_completion_context) = "config"' -a schema -d 'Print the project configuration JSON Schema'
complete -c specutil -n 'test (__specutil_completion_context) = "config schema"' -l repo -s C -r -d 'repository root containing the openspec/ directory'
complete -c specutil -n 'test (__specutil_completion_context) = "generate"' -l check -d 'fail when a generated artifact is stale'
complete -c specutil -n 'test (__specutil_completion_context) = "generate"' -l help -s h -d 'help for generate'
complete -c specutil -n 'test (__specutil_completion_context) = "generate"' -l repo -s C -r -d 'repository root containing the openspec/ directory'
complete -c specutil -n 'test (__specutil_completion_context) = "graph"' -l as -r -d 'output format: json|mermaid|dot|detail'
complete -c specutil -n 'test (__specutil_completion_context) = "graph"' -l command -r -d 'executable passed to the optional command provider'
complete -c specutil -n 'test (__specutil_completion_context) = "graph"' -l out -s o -r -d 'write output to a file instead of stdout'
complete -c specutil -n 'test (__specutil_completion_context) = "graph"' -f -l provider -r -a '(__specutil_completion_values_2)' -d 'external suggestion provider (default: heuristic only)'
complete -c specutil -n 'test (__specutil_completion_context) = "graph"' -l suggest -d 'infer candidate edges from shared capabilities (read-only)'
complete -c specutil -n 'test (__specutil_completion_context) = "graph"' -l repo -s C -r -d 'repository root containing the openspec/ directory'
complete -c specutil -n 'test (__specutil_completion_context) = "next"' -l as -r -d 'output format: text|json'
complete -c specutil -n 'test (__specutil_completion_context) = "next"' -f -l change -r -a '(__specutil_completion_values_4)' -d 'report a single change (or pass as positional arg)'
complete -c specutil -n 'test (__specutil_completion_context) = "next"' -l out -s o -r -d 'write output to a file instead of stdout'
complete -c specutil -n 'test (__specutil_completion_context) = "next"' -l repo -s C -r -d 'repository root containing the openspec/ directory'
complete -c specutil -f -n 'test (__specutil_completion_context) = "next"' -a '(__specutil_completion_values_3)'
complete -c specutil -n 'test (__specutil_completion_context) = "provider"' -l repo -s C -r -d 'repository root containing the openspec/ directory'
complete -c specutil -f -n 'test (__specutil_completion_context) = "provider"' -a list -d 'List discovered suggestion providers'
complete -c specutil -f -n 'test (__specutil_completion_context) = "provider"' -a validate -d 'Validate provider manifests and runtime commands'
complete -c specutil -n 'test (__specutil_completion_context) = "provider list"' -l json -d 'emit provider metadata as JSON'
complete -c specutil -n 'test (__specutil_completion_context) = "provider list"' -l repo -s C -r -d 'repository root containing the openspec/ directory'
complete -c specutil -n 'test (__specutil_completion_context) = "provider validate"' -l json -d 'emit validation reports as JSON'
complete -c specutil -n 'test (__specutil_completion_context) = "provider validate"' -l repo -s C -r -d 'repository root containing the openspec/ directory'
complete -c specutil -f -n 'test (__specutil_completion_context) = "provider validate"' -a '(__specutil_completion_values_5)'
complete -c specutil -n 'test (__specutil_completion_context) = "render"' -l as -r -d 'target format: rfc|design|tickets (required)'
complete -c specutil -n 'test (__specutil_completion_context) = "render"' -f -l change -r -a '(__specutil_completion_values_7)' -d 'change name to render (or pass as positional arg)'
complete -c specutil -n 'test (__specutil_completion_context) = "render"' -l out -s o -r -d 'write output to a file instead of stdout'
complete -c specutil -n 'test (__specutil_completion_context) = "render"' -l templates -r -d 'override built-in template directory'
complete -c specutil -n 'test (__specutil_completion_context) = "render"' -l repo -s C -r -d 'repository root containing the openspec/ directory'
complete -c specutil -f -n 'test (__specutil_completion_context) = "render"' -a '(__specutil_completion_values_6)'
complete -c specutil -n 'test (__specutil_completion_context) = "review"' -l repo -s C -r -d 'repository root containing the openspec/ directory'
complete -c specutil -f -n 'test (__specutil_completion_context) = "review"' -a diff -d 'Show the working-tree diff since a change was reviewed'
complete -c specutil -f -n 'test (__specutil_completion_context) = "review"' -a ingest -d 'Fold an annotation export from the web page into the review record'
complete -c specutil -f -n 'test (__specutil_completion_context) = "review"' -a set -d 'Record a decision on a change without going through the browser'
complete -c specutil -f -n 'test (__specutil_completion_context) = "review"' -a show -d 'Report the recorded decision, open comments, and drift since review'
complete -c specutil -n 'test (__specutil_completion_context) = "review diff"' -l as -r -d 'output format: text|json'
complete -c specutil -n 'test (__specutil_completion_context) = "review diff"' -l base -r -d 'git ref to compare against (default: the reviewed commit, else HEAD)'
complete -c specutil -n 'test (__specutil_completion_context) = "review diff"' -f -l change -r -a '(__specutil_completion_values_9)' -d 'change whose review supplies the base (or pass as positional arg)'
complete -c specutil -n 'test (__specutil_completion_context) = "review diff"' -l out -s o -r -d 'write output to a file instead of stdout'
complete -c specutil -n 'test (__specutil_completion_context) = "review diff"' -l path -r -d 'restrict the diff to these paths'
complete -c specutil -n 'test (__specutil_completion_context) = "review diff"' -l spec-only -d 'restrict the diff to the change\'s own artifact directory'
complete -c specutil -n 'test (__specutil_completion_context) = "review diff"' -l repo -s C -r -d 'repository root containing the openspec/ directory'
complete -c specutil -f -n 'test (__specutil_completion_context) = "review diff"' -a '(__specutil_completion_values_8)'
complete -c specutil -n 'test (__specutil_completion_context) = "review ingest"' -f -l change -r -a '(__specutil_completion_values_10)' -d 'override the change named in the feedback document'
complete -c specutil -n 'test (__specutil_completion_context) = "review ingest"' -l dry-run -d 'print the brief without writing the record'
complete -c specutil -n 'test (__specutil_completion_context) = "review ingest"' -l out -s o -r -d 'write output to a file instead of stdout'
complete -c specutil -n 'test (__specutil_completion_context) = "review ingest"' -l repo -s C -r -d 'repository root containing the openspec/ directory'
complete -c specutil -n 'test (__specutil_completion_context) = "review set"' -f -l change -r -a '(__specutil_completion_values_12)' -d 'change to record against (or pass as positional arg)'
complete -c specutil -n 'test (__specutil_completion_context) = "review set"' -l clear-comments -d 'drop the task comments carried in the record'
complete -c specutil -n 'test (__specutil_completion_context) = "review set"' -l decision -r -d 'approved|changes-requested|commented (required)'
complete -c specutil -n 'test (__specutil_completion_context) = "review set"' -l note -r -d 'note to record with the decision'
complete -c specutil -n 'test (__specutil_completion_context) = "review set"' -l out -s o -r -d 'write output to a file instead of stdout'
complete -c specutil -n 'test (__specutil_completion_context) = "review set"' -l repo -s C -r -d 'repository root containing the openspec/ directory'
complete -c specutil -f -n 'test (__specutil_completion_context) = "review set"' -a '(__specutil_completion_values_11)'
complete -c specutil -n 'test (__specutil_completion_context) = "review show"' -l as -r -d 'output format: text|json'
complete -c specutil -n 'test (__specutil_completion_context) = "review show"' -f -l change -r -a '(__specutil_completion_values_14)' -d 'change to report (or pass as positional arg)'
complete -c specutil -n 'test (__specutil_completion_context) = "review show"' -l out -s o -r -d 'write output to a file instead of stdout'
complete -c specutil -n 'test (__specutil_completion_context) = "review show"' -l repo -s C -r -d 'repository root containing the openspec/ directory'
complete -c specutil -f -n 'test (__specutil_completion_context) = "review show"' -a '(__specutil_completion_values_13)'
complete -c specutil -n 'test (__specutil_completion_context) = "web"' -l base -r -d 'git ref for --diff (default: the reviewed commit, else HEAD)'
complete -c specutil -n 'test (__specutil_completion_context) = "web"' -f -l change -r -a '(__specutil_completion_values_15)' -d 'change the --diff belongs to'
complete -c specutil -n 'test (__specutil_completion_context) = "web"' -l diff -d 'include the working-tree diff for annotation (requires a single change)'
complete -c specutil -n 'test (__specutil_completion_context) = "web"' -l open -d 'open the generated page in the default browser'
complete -c specutil -n 'test (__specutil_completion_context) = "web"' -l out -s o -r -d 'output HTML file path (default: timestamped temp file; \'-\' for stdout)'
complete -c specutil -n 'test (__specutil_completion_context) = "web"' -l repo -s C -r -d 'repository root containing the openspec/ directory'
