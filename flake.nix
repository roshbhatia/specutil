{
  description = "Project OpenSpec change artifacts into review-ready outputs";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";

  outputs =
    { self, nixpkgs, ... }:
    let
      supportedSystems = [
        "aarch64-darwin"
        "aarch64-linux"
        "x86_64-linux"
      ];
      eachSystem = nixpkgs.lib.genAttrs supportedSystems;
    in
    {
      formatter = eachSystem (
        system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
        in
        pkgs.writeShellApplication {
          name = "specutil-format";
          runtimeInputs = [
            pkgs.fd
            pkgs.nixfmt
          ];
          text = ''
            if [ "$#" -gt 0 ] && [ "''${1#-}" = "$1" ]; then
              exec nixfmt "$@"
            fi
            exec fd --extension nix --type file --exec-batch nixfmt "$@"
          '';
        }
      );

      packages = eachSystem (
        system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
          version = "0.3.0";
          source = nixpkgs.lib.cleanSourceWith {
            src = ./.;
            filter =
              path: type:
              let
                name = builtins.baseNameOf path;
              in
              nixpkgs.lib.cleanSourceFilter path type
              && !(
                type == "directory"
                && builtins.elem name [
                  "dist"
                  "docs"
                ]
              );
          };
          mkPackage =
            {
              name,
              subPackage,
              builtName ? name,
              runtimeInputs ? [ ],
              completions ? false,
              providerManifest ? null,
            }:
            pkgs.buildGoModule {
              pname = name;
              inherit version;
              src = source;
              vendorHash = "sha256-z9lPkYq5k69ntR7sKhUqW0KCwMbC7lnEiJZhyVtToXo=";
              subPackages = [ subPackage ];
              nativeBuildInputs = [ pkgs.makeWrapper ] ++ pkgs.lib.optional completions pkgs.installShellFiles;
              nativeCheckInputs = [ pkgs.git ];
              doCheck = completions;
              checkPhase = pkgs.lib.optionalString completions ''
                runHook preCheck
                go test -race ./...
                go run ./cmd/specutil generate --check
                runHook postCheck
              '';
              ldflags = pkgs.lib.optionals completions [
                "-s"
                "-w"
                "-X main.version=${version}"
              ];
              postInstall = ''
                ${pkgs.lib.optionalString (builtName != name) ''
                  mv "$out/bin/${builtName}" "$out/bin/${name}"
                ''}
                wrapProgram "$out/bin/${name}" \
                  --prefix PATH : ${pkgs.lib.makeBinPath runtimeInputs}
              ''
              + pkgs.lib.optionalString completions ''
                installShellCompletion \
                  --cmd specutil \
                  --bash <("$out/bin/specutil" completion bash) \
                  --fish <("$out/bin/specutil" completion fish) \
                  --zsh <("$out/bin/specutil" completion zsh)
                mkdir -p "$out/share/nushell/vendor/autoload"
                "$out/bin/specutil" completion nu > "$out/share/nushell/vendor/autoload/specutil.nu"
                mkdir -p "$out/share/specutil/schema"
                install -m 0444 ${./schema/specutil.schema.json} "$out/share/specutil/schema/specutil.schema.json"
                install -m 0444 ${./schema/provider.schema.json} "$out/share/specutil/schema/provider.schema.json"
              ''
              + pkgs.lib.optionalString (providerManifest != null) ''
                mkdir -p "$out/share/specutil/providers"
                install -m 0444 ${providerManifest} "$out/share/specutil/providers/command.yaml"
              '';
              meta = {
                description = "Project OpenSpec changes into docs, graphs, checks, and review artifacts";
                homepage = "https://github.com/roshbhatia/specutil";
                license = pkgs.lib.licenses.mit;
                mainProgram = name;
                platforms = pkgs.lib.platforms.darwin ++ pkgs.lib.platforms.linux;
              };
              passthru = { inherit runtimeInputs; };
            };
          specutil = mkPackage {
            name = "specutil";
            subPackage = "./cmd/specutil";
            runtimeInputs = [ pkgs.git ];
            completions = true;
          };
          providerCommand = mkPackage {
            name = "specutil-provider-command";
            builtName = "command";
            subPackage = "./extras/command";
            providerManifest = ./extras/command/provider.yaml;
          };
          extras = pkgs.symlinkJoin {
            name = "specutil-providers-${version}";
            paths = [ providerCommand ];
          };
          full = pkgs.symlinkJoin {
            name = "specutil-full-${version}";
            paths = [
              specutil
              extras
            ];
            nativeBuildInputs = [ pkgs.makeWrapper ];
            postBuild = ''
              wrapProgram "$out/bin/specutil" \
                --prefix PATH : "${extras}/bin" \
                --prefix XDG_DATA_DIRS : "${extras}/share"
            '';
            meta = specutil.meta // {
              mainProgram = "specutil";
            };
          };
        in
        {
          inherit specutil extras full;
          provider-command = providerCommand;
          default = specutil;
        }
      );

      apps = eachSystem (system: {
        default = {
          type = "app";
          program = "${nixpkgs.lib.getExe self.packages.${system}.default}";
        };
      });

      checks = eachSystem (
        system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
          packages = self.packages.${system};
          testAgent = pkgs.writeShellScriptBin "test-agent" ''
            printf '%s\n' '{"suggestions":[{"from":"a","to":"b","reason":"uses A"}]}'
          '';
        in
        {
          default = packages.default;
          installed-schemas = pkgs.runCommand "specutil-installed-schemas" { } ''
            test -f ${packages.default}/share/specutil/schema/specutil.schema.json
            test -f ${packages.default}/share/specutil/schema/provider.schema.json
            cmp ${./schema/specutil.schema.json} ${packages.default}/share/specutil/schema/specutil.schema.json
            cmp ${./schema/provider.schema.json} ${packages.default}/share/specutil/schema/provider.schema.json
            touch "$out"
          '';
          provider-manifest =
            pkgs.runCommand "specutil-provider-manifest" { nativeBuildInputs = [ pkgs.cue ]; }
              ''
                cue vet ${./schema/provider.cue} ${./extras/command/provider.yaml} -d '#Provider'
                touch "$out"
              '';
          full-provider = pkgs.runCommand "specutil-full-provider" { nativeBuildInputs = [ pkgs.jq ]; } ''
            mkdir -p "$TMPDIR/repo/openspec/changes/a" "$TMPDIR/repo/openspec/changes/b"
            printf '## Why\n\nA.\n' > "$TMPDIR/repo/openspec/changes/a/proposal.md"
            printf '## 1. Work\n\n- [ ] 1.1 A\n' > "$TMPDIR/repo/openspec/changes/a/tasks.md"
            printf '## Why\n\nB.\n' > "$TMPDIR/repo/openspec/changes/b/proposal.md"
            printf '## 1. Work\n\n- [ ] 1.1 B\n' > "$TMPDIR/repo/openspec/changes/b/tasks.md"
            export HOME="$TMPDIR/home"
            export XDG_DATA_HOME="$TMPDIR/data"
            export XDG_DATA_DIRS="${packages.extras}/share"
            export PATH="${testAgent}/bin:${packages.full}/bin:${pkgs.coreutils}/bin:${pkgs.jq}/bin"
            specutil provider validate command
            output=$(specutil -C "$TMPDIR/repo" graph --suggest --provider command --command test-agent)
            test "$(printf '%s' "$output" | jq -r '.candidates[0].from')" = a
            touch "$out"
          '';
          media-freshness =
            pkgs.runCommand "specutil-media-freshness"
              {
                nativeBuildInputs = [
                  pkgs.coreutils
                  pkgs.ffmpeg
                  pkgs.findutils
                ];
              }
              ''
                ${pkgs.bash}/bin/bash ${./.}/hack/screenshots.sh --check
                touch "$out"
              '';
        }
      );

      devShells = eachSystem (
        system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
        in
        {
          default = pkgs.mkShellNoCC {
            packages = [
              pkgs.actionlint
              pkgs.charm-freeze
              pkgs.cue
              pkgs.fd
              pkgs.ffmpeg
              pkgs.fish
              pkgs.git
              pkgs.go
              pkgs.gofumpt
              pkgs.gopls
              pkgs.goreleaser
              pkgs.gotools
              pkgs.go-tools
              pkgs.nushell
              pkgs.ripgrep
              pkgs.shfmt
              pkgs.vhs
              pkgs.zsh
            ];
            shellHook = ''
              export GOTOOLCHAIN=local
            '';
          };
        }
      );

      lib.skills = {
        "discover-deps" = ./skills/discover-deps/SKILL.md;
        "review-change" = ./skills/review-change/SKILL.md;
      };
      lib.schemas = {
        project = ./schema/specutil.schema.json;
        provider = ./schema/provider.schema.json;
      };
    };
}
