{
  name = "ZFS Encryption";
  interactive.sshBackdoor.enable = true;

  nodes = {
    zfs =
      { pkgs, ... }:
      {
        services.paretosecurity.enable = true;

        # Enable ZFS
        boot.supportedFilesystems = [ "zfs" ];
        boot.zfs.forceImportRoot = false;
        networking.hostId = "8425e349"; # Required for ZFS

        environment.systemPackages = with pkgs; [
          zfs
        ];

        # Add extra disks for ZFS pools
        virtualisation.emptyDiskImages = [
          1024
          1024
        ];
      };
  };

  testScript = ''
    zfs.wait_for_unit("multi-user.target")

    # Test 1: `Filesystem encryption` check fails with unencrypted ZFS pool
    zfs.succeed("zpool create -f plainpool /dev/vdb")
    zfs.succeed("zfs create plainpool/test")

    # Verify pool exists and has no encryption
    assert "off" in zfs.succeed("zfs get -H -o value encryption plainpool") or "none" in zfs.succeed("zfs get -H -o value encryption plainpool")

    # Check should fail with unencrypted ZFS pool
    zfs.fail("paretosecurity check --only 21830a4e-84f1-48fe-9c5b-beab436b2cdb")

    # Test 2: `Filesystem encryption` check passes when any ZFS pool has encryption enabled
    zfs.succeed("echo 'supersecret' | zpool create -f -o ashift=12 -O encryption=aes-256-gcm -O keylocation=prompt -O keyformat=passphrase cryptpool /dev/vdc")
    zfs.succeed("zfs create cryptpool/test")

    # Verify encryption is enabled
    assert "aes-256-gcm" in zfs.succeed("zfs get -H -o value encryption cryptpool")
    keystatus = zfs.succeed("zfs get -H -o value keystatus cryptpool")
    assert "available" in keystatus, f"Expected encryption key to be available, got {keystatus}"

    # Verify both pools exist
    pools = zfs.succeed("zpool list -H -o name")
    assert "plainpool" in pools
    assert "cryptpool" in pools

    # Verify ZFS encryption is detected by the check - should pass now
    # Note: The check passes if ANY ZFS pool has encryption enabled,
    # even if other unencrypted pools exist on the system
    zfs.succeed("paretosecurity check --only 21830a4e-84f1-48fe-9c5b-beab436b2cdb")
  '';
}
