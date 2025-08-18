{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
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
      }: let
        # Extend pkgs with our paretosecurity overlay
        pkgsOverlayed = pkgs.extend (final: prev: {
          paretosecurity = prev.paretosecurity.overrideAttrs (oldAttrs: {
            src = {
              outPath = ./.;
              rev = final.lib.substring 0 8 (builtins.hashFile "sha256" ./go.sum);
            };
            version = "${builtins.hashFile "sha256" "${toString ./go.sum}"}";
            vendorHash = "sha256-DlCGCheJHa4HPM7kfX/UbOfLukAiaoP7QZnabkZVASM=";
          });
        });
      in {
        packages.default = pkgsOverlayed.paretosecurity;

        checks = {
          cli = pkgsOverlayed.testers.runNixOSTest ./test/integration/cli.nix;
          firewall = pkgsOverlayed.testers.runNixOSTest ./test/integration/firewall.nix;
          help = pkgsOverlayed.testers.runNixOSTest ./test/integration/help.nix;
          luks = pkgsOverlayed.testers.runNixOSTest ./test/integration/luks.nix;
          pwd-manager = pkgsOverlayed.testers.runNixOSTest ./test/integration/pwd-manager.nix;
          screenlock = pkgsOverlayed.testers.runNixOSTest ./test/integration/screenlock.nix;
          secureboot = pkgsOverlayed.testers.runNixOSTest ./test/integration/secureboot.nix;
          trayicon = pkgsOverlayed.testers.runNixOSTest ./test/integration/trayicon.nix;
          xfce = pkgsOverlayed.testers.runNixOSTest ./test/integration/desktop/xfce.nix;
        };
      };
    };
}
