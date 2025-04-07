{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/master";
    systems.url = "github:nix-systems/default";
  };
  outputs = inputs @ {
    flake-parts,
    nixpkgs,
    self,
    ...
  }:
    flake-parts.lib.mkFlake {inherit inputs;} {
      systems = nixpkgs.lib.systems.flakeExposed;

      perSystem = {
        config,
        pkgs,
        lib,
        self,
        system,
        ...
      }: {
        packages.default = import ./package.nix {inherit pkgs lib;};
        checks =
          lib.mapAttrs (
            name: path:
              pkgs.testers.runNixOSTest path
          ) {
            pwd-manager = ./test/integration/pwd-manager.nix;
            firewall = ./test/integration/firewall.nix;
            secureboot = ./test/integration/secureboot.nix;
            screenlock = ./test/integration/screenlock.nix;
            luks = ./test/integration/luks.nix;
          };
      };
    };
}
