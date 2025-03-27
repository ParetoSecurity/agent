let
  # Override paretosecurity to use the local codebase
  pareto = {
    pkgs,
    lib,
    ...
  }: {
    services.paretosecurity = {
      enable = true;
      package = pkgs.callPackage ../../package.nix {inherit lib;};
    };
  };

  # Easier tests debugging by SSH-ing into nodes
  ssh = {port}: {...}: {
    services.openssh = {
      enable = true;
      settings = {
        PermitRootLogin = "yes";
        PermitEmptyPasswords = "yes";
      };
    };
    security.pam.services.sshd.allowNullPassword = true;
    virtualisation.forwardPorts = [
      {
        from = "host";
        host.port = port;
        guest.port = 22;
      }
    ];
  };
in {
  name = "SecureBoot is enabled";

  nodes = {
    regularboot = {
      pkgs,
      lib,
      ...
    }: {
      imports = [
        (pareto {inherit pkgs lib;})
      ];
    };

    secureboot = {
      pkgs,
      lib,
      ...
    }: {
      imports = [
        (pareto {inherit pkgs lib;})
      ];
      # NixOS SecureBoot test VM configuration taken from
      # https://github.com/NixOS/nixpkgs/blob/master/nixos/tests/systemd-boot.nix

      virtualisation.useSecureBoot = true;
      virtualisation.useBootLoader = true;
      virtualisation.useEFIBoot = true;
      boot.loader.systemd-boot.enable = true;
      boot.loader.efi.canTouchEfiVariables = true;
      environment.systemPackages = [pkgs.efibootmgr pkgs.sbctl];
      system.switch.enable = true;
    };
  };

  interactive.nodes.regularboot = {...}:
    ssh {port = 2221;} {};

  interactive.nodes.secureboot = {...}:
    ssh {port = 2222;} {};

  testScript = {nodes, ...}: ''
    # Test 1: check fails with SecureBoot disabled
    out = regularboot.fail("paretosecurity check --only c96524f2-850b-4bb9-abc7-517051b6c14e")
    expected = (
        "  • Starting checks...\n"
        "  • System Integrity: SecureBoot is enabled > [FAIL] System is not running in UEFI mode\n"
        "  • Checks completed.\n"
    )
    assert out == expected, f"Expected did not match actual, got \n{out}"

    # Test 2: check succeeds with SecureBoot enabled
    secureboot.start(allow_reboot=True)
    secureboot.wait_for_unit("multi-user.target")

    secureboot.succeed("sbctl create-keys")
    secureboot.succeed("sbctl enroll-keys --yes-this-might-brick-my-machine")
    secureboot.succeed('sbctl sign /boot/EFI/systemd/systemd-boot*.efi')
    secureboot.succeed('sbctl sign /boot/EFI/BOOT/BOOT*.EFI')
    secureboot.succeed('sbctl sign /boot/EFI/nixos/*-linux-*Image.efi')

    secureboot.reboot()
    assert "Secure Boot: enabled (user)" in secureboot.succeed("bootctl status")

    out = secureboot.succeed("paretosecurity check --only c96524f2-850b-4bb9-abc7-517051b6c14e")
    expected = (
        "  • Starting checks...\n"
        "  • System Integrity: SecureBoot is enabled > [OK] SecureBoot is enabled\n"
        "  • Checks completed.\n"
    )
    assert out == expected, f"Expected did not match actual, got \n{out}"
  '';
}
