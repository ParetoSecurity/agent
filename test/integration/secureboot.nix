let
  common = import ./common.nix;
  inherit (common) pareto testHelpers;
  inherit (testHelpers) formatCheckOutput checkMessages;

  securebootUuid = "c96524f2-850b-4bb9-abc7-517051b6c14e";
in {
  name = "SecureBoot";
  interactive.sshBackdoor.enable = true;

  nodes = {
    regularboot = {
      config,
      lib,
      pkgs,
      ...
    }: {
      imports = [
        (pareto {inherit config lib pkgs;})
      ];
    };

    secureboot = {
      config,
      lib,
      pkgs,
      ...
    }: {
      imports = [
        (pareto {inherit config lib pkgs;})
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

  testScript = {nodes, ...}: ''
    with subtest("Check fails with SecureBoot disabled"):
      out = regularboot.fail("paretosecurity check --only ${securebootUuid}")
      expected = ${builtins.toJSON (formatCheckOutput [checkMessages.secureboot.fail])}
      assert out == expected, f"Expected:\n{expected}\n\nActual:\n{out}"

    with subtest("Check succeeds with SecureBoot enabled"):
      secureboot.start(allow_reboot=True)
      secureboot.wait_for_unit("multi-user.target")

      # Setup SecureBoot keys and sign bootloader
      secureboot.succeed("sbctl create-keys")
      secureboot.succeed("sbctl enroll-keys --yes-this-might-brick-my-machine")
      secureboot.succeed('sbctl sign /boot/EFI/systemd/systemd-boot*.efi')
      secureboot.succeed('sbctl sign /boot/EFI/BOOT/BOOT*.EFI')
      secureboot.succeed('sbctl sign /boot/EFI/nixos/*-linux-*Image.efi')

      # Reboot to activate SecureBoot
      secureboot.reboot()
      assert "Secure Boot: enabled (user)" in secureboot.succeed("bootctl status")

      # Check should now pass
      out = secureboot.succeed("paretosecurity check --only ${securebootUuid}")
      expected = ${builtins.toJSON (formatCheckOutput ["System Integrity: SecureBoot is enabled > [OK] Secure Boot is enabled"])}
      assert out == expected, f"Expected:\n{expected}\n\nActual:\n{out}"
  '';
}
