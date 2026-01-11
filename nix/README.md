# Nix support for zapret-discord-youtube-ng

## Installation

### From flake (recommended)
```bash
nix flake install github:Sergeydigl3/zapret-discord-youtube-ng
```

### From local clone
```bash
git clone https://github.com/Sergeydigl3/zapret-discord-youtube-ng
cd zapret-nix
nix flake install
```

### Into NixOS configuration
Add to your `configuration.nix`:
```nix
{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    zapret-discord-youtube-ng.url = "github:Sergeydigl3/zapret-discord-youtube-ng";
  };

  outputs = { self, nixpkgs, zapret-discord-youtube-ng }: {
    nixosConfigurations.myhost = nixpkgs.lib.nixosSystem {
      system = "x86_64-linux";
      modules = [
        {
          nixpkgs.overlays = [ zapret-discord-youtube-ng.overlays.default ];
          environment.systemPackages = with pkgs; [ zapret-discord-youtube-ng ];
        }
      ];
    };
  };
}
```

## Development

Enter development shell with all dependencies:
```bash
nix develop
```

Then use standard commands:
```bash
make build
make run-daemon
```

## How it works

- `flake.nix` defines the package and dev environment
- When installed, Nix automatically:
  1. Clones the repository
  2. Builds Go modules
  3. Generates protobuf code
  4. Compiles binaries
  5. Installs them to your Nix profile

No pre-built binaries needed!
