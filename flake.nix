{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    nix-vm-test.url = "github:numtide/nix-vm-test";
  };

  outputs = inputs @ {
    flake-parts,
    nixpkgs,
    nix-vm-test,
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
        testPackage = {
          distro,
          version,
          script,
        }:
          (inputs.nix-vm-test.lib.x86_64-linux.${distro}.${version} {
            sharedDirs.packageDir = {
              source =  ./. + "/pkg";
              target = "/mnt/package";
            };
            testScript = builtins.readFile (./. + "/test/integration/${script}");
          })
          .driver;
        testRelease = {
          distro,
          version,
          script,
        }:
          (inputs.nix-vm-test.lib.x86_64-linux.${distro}.${version} {
            sharedDirs = {};
            testScript = builtins.readFile (./. + "/test/integration/${script}");
          })
          .driver;
      in {
        _module.args.pkgs = import inputs.nixpkgs {
          inherit system;
          config = {allowUnsupportedSystem = true; };
  };
        packages.default = flakePackage;

        checks = {
          cli = pkgs.testers.runNixOSTest ./test/integration/cli.nix;
          firewall = pkgs.testers.runNixOSTest ./test/integration/firewall.nix;
          help = pkgs.testers.runNixOSTest ./test/integration/help.nix;
          luks = pkgs.testers.runNixOSTest ./test/integration/luks.nix;
          pwd-manager = pkgs.testers.runNixOSTest ./test/integration/pwd-manager.nix;
          screenlock = pkgs.testers.runNixOSTest ./test/integration/screenlock.nix;
          secureboot = pkgs.testers.runNixOSTest ./test/integration/secureboot.nix;
          xfce = pkgs.testers.runNixOSTest ./test/integration/desktop/xfce.nix;
        };
      };
    };
}
