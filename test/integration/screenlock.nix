let
  common = import ./common.nix;
  inherit (common) pareto ssh;
in {
  name = "Screen Lock";

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

  interactive.nodes.gnome = {...}:
    ssh {port = 2221;} {};

  interactive.nodes.kde = {...}:
    ssh {port = 2222;} {};

  testScript = ''
    # Test GNOME
    # Test 1: Check passes by default
    out = gnome.succeed("paretosecurity check --only 37dee029-605b-4aab-96b9-5438e5aa44d8")
    expected = (
        "  • Starting checks...\n"
        "  • Access Security: Password is required to unlock the screen > [OK] Password after sleep or screensaver is on\n"
        "  • Checks completed.\n"
    )
    assert out == expected, f"Expected did not match actual, got \n{out}"

    # Test 2: Check fails when lock is disabled
    gnome.succeed("dbus-run-session -- gsettings set org.gnome.desktop.screensaver lock-enabled false")
    status, out = gnome.execute("paretosecurity check --only 37dee029-605b-4aab-96b9-5438e5aa44d8")
    expected = (
        "  • Starting checks...\n"
        "  • Access Security: Password is required to unlock the screen > [FAIL] Password after sleep or screensaver is off\n"
        "  • Checks completed.\n"
    )
    assert out == expected, f"Expected did not match actual, got \n{out}"

    # Test KDE
    # Test 1: Check passes with lock enabled
    kde.succeed("kwriteconfig5 --file kscreenlockerrc --group Daemon --key Autolock true")
    out = kde.succeed("paretosecurity check --only 37dee029-605b-4aab-96b9-5438e5aa44d8")
    expected = (
        "  • Starting checks...\n"
        "  • Access Security: Password is required to unlock the screen > [OK] Password after sleep or screensaver is on\n"
        "  • Checks completed.\n"
    )
    assert out == expected, f"Expected did not match actual, got \n{out}"

    # Test 2: Check fails when lock is disabled
    kde.succeed("kwriteconfig5 --file kscreenlockerrc --group Daemon --key Autolock false")
    status, out = kde.execute("paretosecurity check --only 37dee029-605b-4aab-96b9-5438e5aa44d8")
    expected = (
        "  • Starting checks...\n"
        "  • Access Security: Password is required to unlock the screen > [FAIL] Password after sleep or screensaver is off\n"
        "  • Checks completed.\n"
    )
    assert out == expected, f"Expected did not match actual, got \n{out}"
  '';
}
