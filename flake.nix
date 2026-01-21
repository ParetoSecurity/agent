{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    nixpkgs-2411.url = "github:NixOS/nixpkgs/nixos-24.11";
  };

  outputs =
    inputs@{
      flake-parts,
      nixpkgs,
      nixpkgs-2411,
      ...
    }:
    flake-parts.lib.mkFlake { inherit inputs; } {
      systems = nixpkgs.lib.systems.flakeExposed;

      perSystem =
        { pkgs, system, ... }:
        let
          # Create pkgs from nixpkgs-24.11 for Plasma 5 testing
          pkgs2411 = import nixpkgs-2411 { inherit system; };

          # Extend pkgs with our paretosecurity overlay
          pkgsOverlayed = pkgs.extend (
            final: prev: {
              paretosecurity = prev.paretosecurity.overrideAttrs (_: {
                src = {
                  outPath = ./.;
                  rev = final.lib.substring 0 8 (builtins.hashFile "sha256" ./go.sum);
                };
                version = "${builtins.hashFile "sha256" "${toString ./go.sum}"}";
                vendorHash = "sha256-jAcUf4VjtKB/Q2wHOVnCgcmPp5XNMqV8Ei9kpWoGO4Q=";
              });
            }
          );
        in
        {
          packages.default = pkgsOverlayed.paretosecurity;

          checks = {
            autologin = pkgsOverlayed.testers.runNixOSTest ./test/integration/autologin.nix;
            cli = pkgsOverlayed.testers.runNixOSTest ./test/integration/cli.nix;
            firewall = pkgsOverlayed.testers.runNixOSTest ./test/integration/firewall.nix;
            help = pkgsOverlayed.testers.runNixOSTest ./test/integration/help.nix;
            luks = pkgsOverlayed.testers.runNixOSTest ./test/integration/luks.nix;
            pwd-manager = pkgsOverlayed.testers.runNixOSTest ./test/integration/pwd-manager.nix;
            secureboot = pkgsOverlayed.testers.runNixOSTest ./test/integration/secureboot.nix;
            trayicon = pkgsOverlayed.testers.runNixOSTest ./test/integration/trayicon.nix;
            xfce = pkgsOverlayed.testers.runNixOSTest ./test/integration/desktop/xfce.nix;
            zfs = pkgsOverlayed.testers.runNixOSTest ./test/integration/zfs.nix;

            screenlock = pkgsOverlayed.testers.runNixOSTest ./test/integration/screenlock.nix;
            screenlock-plasma5 =
              let
                # Extend pkgs2411 with our paretosecurity overlay
                pkgs2411Overlayed = pkgs2411.extend (
                  _: _: {
                    inherit (pkgsOverlayed) paretosecurity;
                  }
                );
              in
              pkgs2411Overlayed.testers.runNixOSTest (
                import ./test/integration/screenlock-plasma5.nix {
                  inherit pkgsOverlayed;
                }
              );
          };
        };
    };
}
