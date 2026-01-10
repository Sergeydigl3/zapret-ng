{
  description = "zapret-ng - network censorship circumvention tool";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      {
        packages.default = pkgs.buildGoModule {
          pname = "zapret-ng";
          version = "0.1.0";
          src = ./.;

          vendorHash = null;

          subPackages = [ "cmd/zapret-daemon" "cmd/zapret" ];

          buildInputs = with pkgs; [
            protobuf_28
          ];

          nativeBuildInputs = with pkgs; [
            pkg-config
          ];

          preBuild = ''
            ${pkgs.lib.getExe pkgs.protobuf_28}/bin/protoc --proto_path=. \
              --go_out=. \
              --go_opt=paths=source_relative \
              --twirp_out=. \
              --twirp_opt=paths=source_relative \
              ./rpc/daemon/service.proto

            go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.11
            go install github.com/twitchtv/twirp@v8.1.3
          '';

          meta = {
            description = "Network censorship circumvention tool";
            homepage = "https://github.com/Sergeydigl3/zapret-ng";
            license = pkgs.lib.licenses.mit;
            platforms = pkgs.lib.platforms.linux;
            maintainers = [ ];
          };
        };

        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go_1_25
            protobuf_28
            pkg-config
            git
          ];

          shellHook = ''
            echo "zapret-ng development environment loaded"
            echo "Go version: $(go version)"
            echo "Protoc version: $(protoc --version)"
            echo "Run 'make build' to build the project"
          '';
        };
      }
    );
}
