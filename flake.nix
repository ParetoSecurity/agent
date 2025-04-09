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
            firewall = ./test/integration/firewall.nix;
            help = ./test/integration/help.nix;
            luks = ./test/integration/luks.nix;
            pwd-manager = ./test/integration/pwd-manager.nix;
            screenlock = ./test/integration/screenlock.nix;
            secureboot = ./test/integration/secureboot.nix;
          };
      };
    };
}
