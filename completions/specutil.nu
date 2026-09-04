export extern "specutil" [
  --repo(-C): string # repository root containing the openspec/ directory
  ...args: string@"__specutil_completion_none"
]

export extern "specutil completion" [
  shell: string@"nu-complete specutil shell"
]

def "nu-complete specutil shell" [] { [bash zsh fish nu] }

export extern "specutil check" [
  --as: string # output format: text|json
  --change: string@"__specutil_completion_values_1" # check a single change (or pass as positional arg)
  --list-rules # list every built-in rule and exit
  --out(-o): string # write output to a file instead of stdout
  --repo(-C): string # repository root containing the openspec/ directory
  ...args: string@"__specutil_completion_values_0"
]

export extern "specutil config" [
  --repo(-C): string # repository root containing the openspec/ directory
  ...args: string@"__specutil_completion_none"
]

export extern "specutil config schema" [
  --repo(-C): string # repository root containing the openspec/ directory
  ...args: string@"__specutil_completion_none"
]

export extern "specutil generate" [
  --check # fail when a generated artifact is stale
  --help(-h) # help for generate
  --repo(-C): string # repository root containing the openspec/ directory
  ...args: string@"__specutil_completion_none"
]

export extern "specutil graph" [
  --as: string # output format: json|mermaid|dot|detail
  --command: string # executable passed to the optional command provider
  --out(-o): string # write output to a file instead of stdout
  --provider: string@"__specutil_completion_values_2" # external suggestion provider (default: heuristic only)
  --suggest # infer candidate edges from shared capabilities (read-only)
  --repo(-C): string # repository root containing the openspec/ directory
  ...args: string@"__specutil_completion_none"
]

export extern "specutil next" [
  --as: string # output format: text|json
  --change: string@"__specutil_completion_values_4" # report a single change (or pass as positional arg)
  --out(-o): string # write output to a file instead of stdout
  --repo(-C): string # repository root containing the openspec/ directory
  ...args: string@"__specutil_completion_values_3"
]

export extern "specutil provider" [
  --repo(-C): string # repository root containing the openspec/ directory
  ...args: string@"__specutil_completion_none"
]

export extern "specutil provider list" [
  --json # emit provider metadata as JSON
  --repo(-C): string # repository root containing the openspec/ directory
  ...args: string@"__specutil_completion_none"
]

export extern "specutil provider validate" [
  --json # emit validation reports as JSON
  --repo(-C): string # repository root containing the openspec/ directory
  ...args: string@"__specutil_completion_values_5"
]

export extern "specutil render" [
  --as: string # target format: rfc|design|tickets (required)
  --change: string@"__specutil_completion_values_7" # change name to render (or pass as positional arg)
  --out(-o): string # write output to a file instead of stdout
  --templates: string # override built-in template directory
  --repo(-C): string # repository root containing the openspec/ directory
  ...args: string@"__specutil_completion_values_6"
]

export extern "specutil review" [
  --repo(-C): string # repository root containing the openspec/ directory
  ...args: string@"__specutil_completion_none"
]

export extern "specutil review diff" [
  --as: string # output format: text|json
  --base: string # git ref to compare against (default: the reviewed commit, else HEAD)
  --change: string@"__specutil_completion_values_9" # change whose review supplies the base (or pass as positional arg)
  --out(-o): string # write output to a file instead of stdout
  --path: string # restrict the diff to these paths
  --spec-only # restrict the diff to the change's own artifact directory
  --repo(-C): string # repository root containing the openspec/ directory
  ...args: string@"__specutil_completion_values_8"
]

export extern "specutil review ingest" [
  --change: string@"__specutil_completion_values_10" # override the change named in the feedback document
  --dry-run # print the brief without writing the record
  --out(-o): string # write output to a file instead of stdout
  --repo(-C): string # repository root containing the openspec/ directory
  ...args: string@"__specutil_completion_none"
]

export extern "specutil review set" [
  --change: string@"__specutil_completion_values_12" # change to record against (or pass as positional arg)
  --clear-comments # drop the task comments carried in the record
  --decision: string # approved|changes-requested|commented (required)
  --note: string # note to record with the decision
  --out(-o): string # write output to a file instead of stdout
  --repo(-C): string # repository root containing the openspec/ directory
  ...args: string@"__specutil_completion_values_11"
]

export extern "specutil review show" [
  --as: string # output format: text|json
  --change: string@"__specutil_completion_values_14" # change to report (or pass as positional arg)
  --out(-o): string # write output to a file instead of stdout
  --repo(-C): string # repository root containing the openspec/ directory
  ...args: string@"__specutil_completion_values_13"
]

export extern "specutil web" [
  --base: string # git ref for --diff (default: the reviewed commit, else HEAD)
  --change: string@"__specutil_completion_values_15" # change the --diff belongs to
  --diff # include the working-tree diff for annotation (requires a single change)
  --open # open the generated page in the default browser
  --out(-o): string # output HTML file path (default: timestamped temp file; '-' for stdout)
  --repo(-C): string # repository root containing the openspec/ directory
  ...args: string@"__specutil_completion_none"
]

def "__specutil_completion_none" [] { [] }

def "__specutil_completion_values_0" [context?: string] {
  [
    (try { run-external "specutil" "__values" "changes" ($context | default "") | lines } catch { [] })
  ] | flatten | uniq
}

def "__specutil_completion_values_1" [context?: string] {
  [
    (try { run-external "specutil" "__values" "changes" ($context | default "") | lines } catch { [] })
  ] | flatten | uniq
}

def "__specutil_completion_values_2" [context?: string] {
  [
    (try { run-external "specutil" "__values" "providers" | lines } catch { [] })
  ] | flatten | uniq
}

def "__specutil_completion_values_3" [context?: string] {
  [
    (try { run-external "specutil" "__values" "changes" ($context | default "") | lines } catch { [] })
  ] | flatten | uniq
}

def "__specutil_completion_values_4" [context?: string] {
  [
    (try { run-external "specutil" "__values" "changes" ($context | default "") | lines } catch { [] })
  ] | flatten | uniq
}

def "__specutil_completion_values_5" [context?: string] {
  [
    (try { run-external "specutil" "__values" "providers" | lines } catch { [] })
  ] | flatten | uniq
}

def "__specutil_completion_values_6" [context?: string] {
  [
    (try { run-external "specutil" "__values" "changes" ($context | default "") | lines } catch { [] })
  ] | flatten | uniq
}

def "__specutil_completion_values_7" [context?: string] {
  [
    (try { run-external "specutil" "__values" "changes" ($context | default "") | lines } catch { [] })
  ] | flatten | uniq
}

def "__specutil_completion_values_8" [context?: string] {
  [
    (try { run-external "specutil" "__values" "changes" ($context | default "") | lines } catch { [] })
  ] | flatten | uniq
}

def "__specutil_completion_values_9" [context?: string] {
  [
    (try { run-external "specutil" "__values" "changes" ($context | default "") | lines } catch { [] })
  ] | flatten | uniq
}

def "__specutil_completion_values_10" [context?: string] {
  [
    (try { run-external "specutil" "__values" "changes" ($context | default "") | lines } catch { [] })
  ] | flatten | uniq
}

def "__specutil_completion_values_11" [context?: string] {
  [
    (try { run-external "specutil" "__values" "changes" ($context | default "") | lines } catch { [] })
  ] | flatten | uniq
}

def "__specutil_completion_values_12" [context?: string] {
  [
    (try { run-external "specutil" "__values" "changes" ($context | default "") | lines } catch { [] })
  ] | flatten | uniq
}

def "__specutil_completion_values_13" [context?: string] {
  [
    (try { run-external "specutil" "__values" "changes" ($context | default "") | lines } catch { [] })
  ] | flatten | uniq
}

def "__specutil_completion_values_14" [context?: string] {
  [
    (try { run-external "specutil" "__values" "changes" ($context | default "") | lines } catch { [] })
  ] | flatten | uniq
}

def "__specutil_completion_values_15" [context?: string] {
  [
    (try { run-external "specutil" "__values" "changes" ($context | default "") | lines } catch { [] })
  ] | flatten | uniq
}
