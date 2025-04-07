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

    sway = {
      pkgs,
      lib,
      ...
    }: {
      imports = [
        (pareto {inherit pkgs lib;})
      ];
      # enable Sway window manager
      programs.sway = {
        enable = true;
        wrapperFeatures.gtk = true;
      };
    };
    swaylock = {
      pkgs,
      lib,
      ...
    }: {
      imports = [
        (pareto {inherit pkgs lib;})
      ];
      # enable Sway window manager
      programs.sway = {
        enable = true;
        wrapperFeatures.gtk = true;
      };
      programs.swayidle = {
        enable = true;
        timeouts = [
          # Lock screen after 300 seconds of inactivity
          {
            timeout = 300;
            command = "${pkgs.swaylock}/bin/swaylock -f -c 000000";
          }

          # Turn off displays after 600 seconds of inactivity
          {
            timeout = 600;
            command = "${pkgs.sway}/bin/swaymsg 'output * dpms off'";
            resumeCommand = "${pkgs.sway}/bin/swaymsg 'output * dpms on'";
          }
        ];

        # Also lock screen before going to sleep
        events = [
          {
            event = "before-sleep";
            command = "${pkgs.swaylock}/bin/swaylock -f -c 000000";
          }
          {
            event = "lock";
            command = "${pkgs.swaylock}/bin/swaylock -f -c 000000";
          }
        ];
      };
      environment.systemPackages = with pkgs; [
        swaylock
      ];
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


    # Test sway, swaylock is disabled by default
    status, out = sway.execute("paretosecurity check --only 37dee029-605b-4aab-96b9-5438e5aa44d8")
    expected = (
        "  • Starting checks...\n"
        "  • Access Security: Password is required to unlock the screen > [FAIL] Password after sleep or screensaver is off\n"
        "  • Checks completed.\n"
    )
    assert out == expected, f"Expected did not match actual, got \n{out}"

    # Test swaylock
    status, out = swaylock.execute("paretosecurity check --only 37dee029-605b-4aab-96b9-5438e5aa44d8")
    expected = (
        "  • Starting checks...\n"
        "  • Access Security: Password is required to unlock the screen > [FAIL] Password after sleep or screensaver is off\n"
        "  • Checks completed.\n"
    )
    assert out == expected, f"Expected did not match actual, got \n{out}"
  '';
}
