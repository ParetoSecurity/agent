let
  common = import ./common.nix;
  inherit (common) pareto ssh;
in {
  name = "SecureBoot";
  interactive.sshBackdoor.enable = true;

  nodes = {
    regularboot = {
      pkgs,
      lib,
      ...
    }: {
      imports = [
        (pareto {inherit pkgs lib;})
      ];

      # Optimize memory for non-secure boot test
      virtualisation.memorySize = 1024;
      virtualisation.cores = 1;
    };

    secureboot = {
      pkgs,
      lib,
      ...
    }: {
      imports = [
        (pareto {inherit pkgs lib;})
      ];
      # NixOS SecureBoot test VM configuration based on
      # https://github.com/NixOS/nixpkgs/blob/master/nixos/tests/systemd-boot.nix

      virtualisation.useSecureBoot = true;
      virtualisation.useBootLoader = true;
      virtualisation.useEFIBoot = true;
      boot.loader.systemd-boot.enable = true;
      boot.loader.efi.canTouchEfiVariables = true;
      environment.systemPackages = [pkgs.efibootmgr pkgs.sbctl];
      system.switch.enable = true;

      # SecureBoot with UEFI needs significant memory for disk image operations
      virtualisation.memorySize = 3072; # 3GB for SecureBoot operations
      virtualisation.cores = 2;

      # Reduce disk size to save memory
      virtualisation.diskSize = 2048; # 2GB disk instead of default

      # Ensure UEFI firmware is properly configured
      virtualisation.efi.firmware = pkgs.OVMF.fd;
    };
  };

  testScript = {nodes, ...}: let
    # Get the correct EFI architecture for signing
    efiArch = nodes.secureboot.pkgs.stdenv.hostPlatform.efiArch;
    kernelFile = nodes.secureboot.system.boot.loader.kernelFile;
  in ''
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

    # Create and enroll SecureBoot keys
    secureboot.succeed("sbctl create-keys")
    secureboot.succeed("sbctl enroll-keys --yes-this-might-brick-my-machine")

    # Sign bootloader and kernel with correct architecture paths
    secureboot.succeed('sbctl sign /boot/EFI/systemd/systemd-boot${efiArch}.efi')
    secureboot.succeed('sbctl sign /boot/EFI/BOOT/BOOT${builtins.toUpper efiArch}.EFI')
    secureboot.succeed('sbctl sign /boot/EFI/nixos/*${kernelFile}.efi')

    # Reboot to activate SecureBoot
    secureboot.reboot()

    # Wait for system to come back up and verify SecureBoot is enabled
    secureboot.wait_for_unit("multi-user.target")
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
