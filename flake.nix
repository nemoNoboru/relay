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
            # Backend (backend/)
            uv

            # Mobile app (mobile/)
            nodejs
            bun

            git
            gh
            curl
            jq
          ];

          shellHook = ''
            echo ""
            echo "Relay dev shell"
            echo "  backend/   cd backend && uv sync && uv run pytest"
            echo "  mobile/    cd mobile && bun install && bun run start"
            echo ""
          '';
        };
      });
    };
}
