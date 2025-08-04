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
        flakePackage = pkgs.callPackage ./package.nix {};
      in {
        packages.default = flakePackage;

        checks = let
          # Create a custom version of pkgs with allowUnsupportedSystem = true, so
          # that we can run tests on Macs too:
          # $ nix build .#checks.aarch64-darwin.firewall
          pkgsAllowUnsupported = import nixpkgs {
            inherit system;
            config = {allowUnsupportedSystem = true;};
          };
        in {
          cli = pkgsAllowUnsupported.testers.runNixOSTest ./test/integration/cli.nix;
          firewall = pkgsAllowUnsupported.testers.runNixOSTest ./test/integration/firewall.nix;
          help = pkgsAllowUnsupported.testers.runNixOSTest ./test/integration/help.nix;
          luks = pkgsAllowUnsupported.testers.runNixOSTest ./test/integration/luks.nix;
          pwd-manager = pkgsAllowUnsupported.testers.runNixOSTest ./test/integration/pwd-manager.nix;
          screenlock = pkgsAllowUnsupported.testers.runNixOSTest ./test/integration/screenlock.nix;
          secureboot = pkgsAllowUnsupported.testers.runNixOSTest ./test/integration/secureboot.nix;
          trayicon = pkgsAllowUnsupported.testers.runNixOSTest ./test/integration/trayicon.nix;
          xfce = pkgsAllowUnsupported.testers.runNixOSTest ./test/integration/desktop/xfce.nix;
          gnome = pkgsAllowUnsupported.testers.runNixOSTest ./test/integration/desktop/gnome.nix;
          kde = pkgsAllowUnsupported.testers.runNixOSTest ./test/integration/desktop/kde.nix;
        };
      };
    };
}
