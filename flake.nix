{
  description = "specutil — projection of spec-framework change artifacts into docs and sync plans";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs =
    {
      nixpkgs,
      flake-utils,
      ...
    }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = import nixpkgs { inherit system; };
        go = pkgs.go;
      in
      {
        packages.default = pkgs.buildGoModule {
          pname = "specutil";
          version = "0.1.0";
          src = ./.;
          # vendorHash is recomputed when the module graph changes; run
          # `nix build` and copy the hash it reports here on dependency bumps.
          vendorHash = "sha256-p3W9SBXEWTL7rWpW95cOoNJ9CArJlAsi3Vfy/m1d2z0=";
          subPackages = [ "cmd/specutil" ];
          meta = {
            description = "Project OpenSpec change artifacts into other artifacts and visualizations";
            mainProgram = "specutil";
          };
        };

        # `nix fmt` is documented in the README and wired to `task fmt:nix`;
        # without this attribute both fail on a missing formatter output. A bare
        # nixfmt would be handed every path in the tree, including the Go and
        # Markdown files it cannot parse, so scope it to .nix files when invoked
        # with no arguments.
        formatter = pkgs.writeShellApplication {
          name = "specutil-nixfmt";
          runtimeInputs = [
            pkgs.fd
            pkgs.nixfmt-rfc-style
          ];
          text = ''
            if [ "$#" -gt 0 ]; then
              exec nixfmt "$@"
            fi

            exec fd --extension nix --type file --exec-batch nixfmt
          '';
        };

        devShells.default = pkgs.mkShell {
          packages = [
            go
            pkgs.gopls
            pkgs.gotools
            pkgs.go-tools # staticcheck
            pkgs.gofumpt
            pkgs.go-task
            pkgs.goreleaser
            pkgs.shfmt
          ];
          shellHook = ''
            export GOTOOLCHAIN=local
          '';
        };
      }
    )
    // {
      # Skill files for AI agent harnesses. Paths resolve into the Nix store
      # when this flake is consumed as an input, so consumers can use them
      # directly as home.file sources without any build step.
      lib.skills = {
        "sync-to-linear" = ./skills/sync-to-linear/SKILL.md;
        "sync-to-notion" = ./skills/sync-to-notion/SKILL.md;
        "sync-to-github-issues" = ./skills/sync-to-github-issues/SKILL.md;
        "discover-deps" = ./skills/discover-deps/SKILL.md;
        "review-change" = ./skills/review-change/SKILL.md;
      };
    };
}
