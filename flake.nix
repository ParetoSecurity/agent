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
              source = "${toString ./.}/pkg";
              target = "/mnt/package";
            };
            testScript = builtins.readFile "${toString ./.}/test/integration/${script}";
          })
          .driver;
        testRelease = {
          distro,
          version,
          script,
        }:
          (inputs.nix-vm-test.lib.x86_64-linux.${distro}.${version} {
            sharedDirs = {};
            testScript = builtins.readFile "${toString ./.}/test/integration/${script}";
          })
          .driver;
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
          xfce = pkgsAllowUnsupported.testers.runNixOSTest ./test/integration/desktop/xfce.nix;
        };
      };
    };
}
