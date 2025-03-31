{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    nix-vm-test.url = "github:numtide/nix-vm-test";
    systems.url = "github:nix-systems/default";
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
        flakePackage = import ./package.nix {inherit pkgs lib;};
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
        checks.test-nixos = import ./test/integration/nixos.nix {
          inherit pkgs system flakePackage;
        };

        checks.pwd-manager = pkgs.testers.runNixOSTest ./test/integration/pwd-manager.nix;
        checks.firewall = pkgs.testers.runNixOSTest ./test/integration/firewall.nix;
        checks.screenlock = pkgs.testers.runNixOSTest ./test/integration/screenlock.nix;

        packages.test-debian = testPackage {
          distro = "debian";
          version = "13";
          script = "debian.py";
        };
        packages.test-fedora = testPackage {
          distro = "fedora";
          version = "40";
          script = "fedora.py";
        };
        packages.test-ubuntu = testPackage {
          distro = "ubuntu";
          version = "23_10";
          script = "ubuntu.py";
        };

        packages.test-release-debian = testRelease {
          distro = "debian";
          version = "13";
          script = "debian-release.py";
        };
        packages.test-release-fedora = testRelease {
          distro = "fedora";
          version = "40";
          script = "fedora-release.py";
        };
        packages.test-release-ubuntu = testRelease {
          distro = "ubuntu";
          version = "23_10";
          script = "ubuntu-release.py";
        };
      };
    };
}
