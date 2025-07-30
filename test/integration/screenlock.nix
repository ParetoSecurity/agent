let
  common = import ./common.nix;
  inherit (common) pareto testHelpers;
  inherit (testHelpers) formatCheckOutput checkMessages;

  screenlockUuid = "37dee029-605b-4aab-96b9-5438e5aa44d8";
in {
  name = "Screen Lock";
  interactive.sshBackdoor.enable = true;

  nodes = {
    gnome = {
      pkgs,
      lib,
      ...
    }: {
      imports = [
        (pareto {inherit pkgs lib;})
      ];
      # Install GNOME Desktop Environment
      services.xserver.desktopManager.gnome.enable = true;
      services.xserver.displayManager.gdm.enable = true;
    };

    kde = {
      pkgs,
      lib,
      ...
    }: {
      imports = [
        (pareto {inherit pkgs lib;})
      ];
      # Install KDE Plasma 5 Desktop Environment
      services.xserver.enable = true;
      services.xserver.desktopManager.plasma5.enable = true;
      services.xserver.displayManager.sddm.enable = true;
      services.colord.enable = false;
    };
  };

  testScript = ''
    with subtest("GNOME screenlock tests"):
      # Test 1: Check passes by default
      out = gnome.succeed("paretosecurity check --only ${screenlockUuid}")
      expected = ${builtins.toJSON (formatCheckOutput [checkMessages.screenlock.ok])}
      assert out == expected, f"Expected:\n{expected}\n\nActual:\n{out}"

      # Test 2: Check fails when lock is disabled
      gnome.succeed("dbus-run-session -- gsettings set org.gnome.desktop.screensaver lock-enabled false")
      out = gnome.fail("paretosecurity check --only ${screenlockUuid}")
      expected = ${builtins.toJSON (formatCheckOutput [checkMessages.screenlock.fail])}
      assert out == expected, f"Expected:\n{expected}\n\nActual:\n{out}"

    with subtest("KDE screenlock tests"):
      # Test 1: Check passes with lock enabled
      kde.succeed("kwriteconfig5 --file kscreenlockerrc --group Daemon --key LockOnResume true")
      out = kde.succeed("paretosecurity check --only ${screenlockUuid}")
      expected = ${builtins.toJSON (formatCheckOutput [checkMessages.screenlock.ok])}
      assert out == expected, f"Expected:\n{expected}\n\nActual:\n{out}"

      # Test 2: Check fails when lock is disabled
      kde.succeed("kwriteconfig5 --file kscreenlockerrc --group Daemon --key LockOnResume false")
      out = kde.fail("paretosecurity check --only ${screenlockUuid}")
      expected = ${builtins.toJSON (formatCheckOutput [checkMessages.screenlock.fail])}
      assert out == expected, f"Expected:\n{expected}\n\nActual:\n{out}"
  '';
}
