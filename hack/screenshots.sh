#!/usr/bin/env bash
set -euo pipefail

repo_dir=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
cd "$repo_dir"
output_dir=${SPECUTIL_MEDIA_OUTPUT_DIR:-"$repo_dir/docs"}
mkdir -p "$output_dir"

media_fingerprint() {
  {
    printf '%s\n' flake.lock flake.nix go.mod go.sum hack/specutil.tape hack/screenshots.sh
    find cmd internal extras -type f \
      \( -name '*.go' -o -name '*.tmpl' -o -name '*.yaml' -o -name '*.nix' \) \
      ! -name '*_test.go' -print | LC_ALL=C sort
  } | while IFS= read -r path; do
    sha256sum "$path"
  done | sha256sum | cut -d ' ' -f 1
}

media_is_valid() {
  local gif_format
  local png_codec
  [[ -s $output_dir/specutil.png && -s $output_dir/specutil.gif ]] || return 1
  png_codec=$(ffprobe -v error -select_streams v:0 -show_entries stream=codec_name \
    -of default=noprint_wrappers=1:nokey=1 "$output_dir/specutil.png") || return 1
  gif_format=$(ffprobe -v error -show_entries format=format_name \
    -of default=noprint_wrappers=1:nokey=1 "$output_dir/specutil.gif") || return 1
  [[ $png_codec == png && $gif_format == gif ]]
}

if [[ ${1:-} == "--check" ]]; then
  expected=$(media_fingerprint)
  current=$(cat "$output_dir/specutil.media.sha256" 2> /dev/null || true)
  if [[ $current != "$expected" ]] || ! media_is_valid; then
    echo "specutil media is stale; run ./hack/screenshots.sh" >&2
    exit 1
  fi
  exit 0
fi

media_root=$(mktemp -d)
fixture="$media_root/home/rotation-service"
trap 'rm -rf "$media_root"' EXIT
mkdir -p \
  "$fixture/openspec/changes/add-key-versioning" \
  "$fixture/openspec/changes/rotate-signing-keys" \
  "$media_root/cache" \
  "$media_root/config" \
  "$media_root/data" \
  "$media_root/home"

printf '%s\n' \
  'changes:' \
  '  rotate-signing-keys:' \
  '    depends_on: [add-key-versioning]' \
  > "$fixture/openspec/specutil.yaml"
printf '%s\n' \
  '## Why' '' \
  'Signing keys need stable version identifiers before rotation can be automated.' '' \
  '## What Changes' '' \
  '- Add versioned key records and lookup support.' \
  > "$fixture/openspec/changes/add-key-versioning/proposal.md"
printf '%s\n' \
  '## 1. Storage' '' \
  '- [x] 1.1 Add a version column to signing key records' \
  '- [x] 1.2 Resolve tokens by key version' \
  > "$fixture/openspec/changes/add-key-versioning/tasks.md"
printf '%s\n' \
  '## Why' '' \
  'Production keys must rotate without invalidating active sessions.' '' \
  '## What Changes' '' \
  '- Add staged activation and retirement for signing keys.' \
  > "$fixture/openspec/changes/rotate-signing-keys/proposal.md"
printf '%s\n' \
  '## 1. Rotation' '' \
  '- [ ] 1.1 Add staged key activation' \
  '- [ ] 1.2 Retire keys after the token overlap window' '' \
  '## 2. Verification' '' \
  '- [ ] 2.1 Verify active sessions across a rotation' \
  > "$fixture/openspec/changes/rotate-signing-keys/tasks.md"

full_path=$(nix build .#full --no-link --print-out-paths)
export HOME="$media_root/home"
export XDG_CACHE_HOME="$media_root/cache"
export XDG_CONFIG_HOME="$media_root/config"
export XDG_DATA_HOME="$media_root/data"
unset SPECUTIL_PROVIDERS_DIRECTORY

(
  cd "$fixture"
  PATH="$full_path/bin:$PATH" freeze \
    --execute "specutil next rotate-signing-keys" \
    --output "$output_dir/specutil.png" \
    --width 1100 \
    --padding 24 \
    --margin 16 \
    --window

  PATH="$full_path/bin:$PATH" vhs "$repo_dir/hack/specutil.tape" \
    --output "$output_dir/specutil.gif"
)

if ! media_is_valid; then
  echo "specutil media generation produced an invalid image" >&2
  exit 1
fi
media_fingerprint > "$output_dir/specutil.media.sha256"
