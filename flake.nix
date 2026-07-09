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
          vendorHash = "sha256-3ToA/pUNam7Bk4UsfwF9CGRsbZdqgRfIuQ4vl+zpxWw=";
          subPackages = [ "cmd/specutil" ];
          meta = {
            description = "Project OpenSpec change artifacts into other artifacts and visualizations";
            mainProgram = "specutil";
          };
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
        "sync-to-linear"  = ./skills/sync-to-linear/SKILL.md;
        "sync-to-notion"  = ./skills/sync-to-notion/SKILL.md;
        "discover-deps"   = ./skills/discover-deps/SKILL.md;
      };
    };
}
