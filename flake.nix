{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
  };

  outputs =
    inputs@{
      flake-parts,
      nixpkgs,
      ...
    }:
    flake-parts.lib.mkFlake { inherit inputs; } {
      systems = nixpkgs.lib.systems.flakeExposed;

      perSystem =
        { pkgs, ... }:
        let
          # Extend pkgs with our paretosecurity overlay
          pkgsOverlayed = pkgs.extend (
            final: prev: {
              paretosecurity = prev.paretosecurity.overrideAttrs (_: {
                src = {
                  outPath = ./.;
                  rev = final.lib.substring 0 8 (builtins.hashFile "sha256" ./go.sum);
                };
                version = "${builtins.hashFile "sha256" "${toString ./go.sum}"}";
                vendorHash = "sha256-3fL67Ir1LfT3DwYKwPPruROafM/CpB/qguT8tMxy6rg=";
              });
            }
          );
        in
        {
          packages.default = pkgsOverlayed.paretosecurity;

          checks = {
            cli = pkgsOverlayed.testers.runNixOSTest ./test/integration/cli.nix;
            firewall = pkgsOverlayed.testers.runNixOSTest ./test/integration/firewall.nix;
            help = pkgsOverlayed.testers.runNixOSTest ./test/integration/help.nix;
            luks = pkgsOverlayed.testers.runNixOSTest ./test/integration/luks.nix;
            zfs = pkgsOverlayed.testers.runNixOSTest ./test/integration/zfs.nix;
            pwd-manager = pkgsOverlayed.testers.runNixOSTest ./test/integration/pwd-manager.nix;
            screenlock = pkgsOverlayed.testers.runNixOSTest ./test/integration/screenlock.nix;
            secureboot = pkgsOverlayed.testers.runNixOSTest ./test/integration/secureboot.nix;
            trayicon = pkgsOverlayed.testers.runNixOSTest ./test/integration/trayicon.nix;
            xfce = pkgsOverlayed.testers.runNixOSTest ./test/integration/desktop/xfce.nix;
            autologin = pkgsOverlayed.testers.runNixOSTest ./test/integration/autologin.nix;
          };
        };
    };
}
