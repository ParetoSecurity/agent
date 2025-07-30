let
  common = import ./common.nix;
  inherit (common) pareto testHelpers;
  inherit (testHelpers) formatCheckOutput checkMessages timeouts;

  luksUuid = "21830a4e-84f1-48fe-9c5b-beab436b2cdb";

  # LUKS configuration constants
  luks = {
    diskSizeMB = 512; # Size of virtual disks for encryption
    passphrase = "supersecret";
    iterTime = 1; # Fast iteration for testing
  };
in {
  name = "FS Encryption";
  interactive.sshBackdoor.enable = true;

  nodes = {
    plaindisk = {
      pkgs,
      lib,
      ...
    }: {
      imports = [
        (pareto {inherit pkgs lib;})
      ];
    };

    luks = {
      pkgs,
      lib,
      config,
      ...
    }: {
      imports = [
        (pareto {inherit pkgs lib;})
      ];

      # NixOS LUKS test VM configuration taken from
      # https://github.com/NixOS/nixpkgs/blob/master/nixos/tests/luks.nix

      boot.initrd.extraUtilsCommands = ''
        # We need mke2fs in the initrd.
        copy_bin_and_libs ${pkgs.e2fsprogs}/bin/mke2fs
      '';

      boot.initrd.postDeviceCommands = ''
        # If the disk image appears to be empty, run mke2fs to
        # initialise.
        FSTYPE=$(blkid -o value -s TYPE ${config.virtualisation.rootDevice} || true)
        PARTTYPE=$(blkid -o value -s PTTYPE ${config.virtualisation.rootDevice} || true)
        if test -z "$FSTYPE" -a -z "$PARTTYPE"; then
            mke2fs -t ext4 ${config.virtualisation.rootDevice}
        fi
      '';

      # Use systemd-boot
      virtualisation = {
        emptyDiskImages = [
          luks.diskSizeMB
          luks.diskSizeMB
        ];
        useBootLoader = true;
        useEFIBoot = true;
        # To boot off the encrypted disk, we need to have a init script which comes from the Nix store
        mountHostNixStore = true;
      };
      boot.loader.systemd-boot.enable = true;

      boot.kernelParams = lib.mkOverride 5 ["console=tty1"];

      environment.systemPackages = with pkgs; [cryptsetup];

      specialisation = rec {
        boot-luks.configuration = {
          boot.initrd.luks.devices = lib.mkVMOverride {
            # We have two disks and only type one password - key reuse is in place
            cryptroot.device = "/dev/vdb";
            cryptroot2.device = "/dev/vdc";
          };
          virtualisation.rootDevice = "/dev/mapper/cryptroot";
        };
        boot-luks-custom-keymap.configuration = lib.mkMerge [
          boot-luks.configuration
          {
            console.keyMap = "neo";
          }
        ];
      };
    };
  };

  enableOCR = true;

  testScript = {nodes, ...}: ''
    with subtest("Check fails with LUKS not configured"):
      out = plaindisk.fail("paretosecurity check --only ${luksUuid}")
      expected = ${builtins.toJSON (formatCheckOutput [checkMessages.luks.fail])}
      assert out == expected, f"Expected:\n{expected}\n\nActual:\n{out}"

    with subtest("Check succeeds with LUKS configured"):
      luks.wait_for_unit("multi-user.target")

      # Setup LUKS encryption on both virtual disks
      luks.succeed("echo -n ${luks.passphrase} | cryptsetup luksFormat -q --iter-time=${toString luks.iterTime} /dev/vdb -")
      luks.succeed("echo -n ${luks.passphrase} | cryptsetup luksFormat -q --iter-time=${toString luks.iterTime} /dev/vdc -")

      # Configure boot to use LUKS specialisation
      luks.succeed("bootctl set-default nixos-generation-1-specialisation-boot-luks.conf")
      luks.succeed("sync")

      # Reboot into encrypted system
      luks.crash()
      luks.start()

      # Enter passphrase at boot prompt
      luks.wait_for_text("Passphrase for")
      luks.send_chars("${luks.passphrase}\n")
      luks.wait_for_unit("multi-user.target")

      # Verify encryption is active
      assert "/dev/mapper/cryptroot on / type ext4" in luks.succeed("mount")

      # Check should now pass
      out = luks.succeed("paretosecurity check --only ${luksUuid}")
      expected = ${builtins.toJSON (formatCheckOutput [checkMessages.luks.ok])}
      assert out == expected, f"Expected:\n{expected}\n\nActual:\n{out}"
  '';
}
