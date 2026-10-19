{
  description = "Relay — a micro-social network for you and your real friends";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-26.05";

  outputs = { self, nixpkgs }:
    let
      systems = [
        "x86_64-linux"
        "aarch64-linux"
        "x86_64-darwin"
        "aarch64-darwin"
      ];
      forAllSystems = f:
        nixpkgs.lib.genAttrs systems (system:
          f (import nixpkgs { inherit system; }));
    in {
      devShells = forAllSystems (pkgs: {
        default = pkgs.mkShell {
          name = "relay-dev";

          packages = with pkgs; [
            # Go node and test gates
            go
            go-tools
            gopls
            delve
            stdenv.cc
            gnumake
            git
            gh
            curl
            jq
            gnutar
            gzip

            # Owner UI and tests
            nodejs

            # Screenshots, contact sheets, and media template checks
            chromium
            imagemagick
            ffmpeg

            # A watcher for local iteration
            watchexec
          ];

          shellHook = ''
            export CHROME_BIN="${pkgs.chromium}/bin/chromium"
            export PUPPETEER_EXECUTABLE_PATH="$CHROME_BIN"
            echo ""
            echo "Relay dev shell"
            echo "  make dev             build web + binary, run a local node"
            echo "  ./scripts/dev.sh     same, without make"
            echo "  make check           web + Go gates"
            echo "  make run             run the existing binary"
            echo "  http://localhost:8080/_relay/admin.html  (password: relay)"
            echo ""
          '';
        };
      });
    };
}
