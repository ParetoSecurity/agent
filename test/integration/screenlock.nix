{
  name = "Screen Lock";
  interactive.sshBackdoor.enable = true;

  nodes = {
    gnome =
      { ... }:
      {
        services.paretosecurity.enable = true;
        # Install GNOME Desktop Environment
        services.xserver.desktopManager.gnome.enable = true;
        services.xserver.displayManager.gdm.enable = true;
      };

    kde =
      { ... }:
      {
        services.paretosecurity.enable = true;
        # Install KDE Plasma 5 Desktop Environment
        services.xserver.enable = true;
        services.xserver.desktopManager.plasma5.enable = true;
        services.xserver.displayManager.sddm.enable = true;
        services.colord.enable = false;
      };

    sway =
      { pkgs, ... }:
      {
        services.paretosecurity.enable = true;

        programs.sway.enable = true;
        programs.sway.wrapperFeatures.gtk = true;

        users.users.testuser.isNormalUser = true;
        users.users.testuser.extraGroups = [
          "wheel"
          "video"
          "input"
        ];

        # Autologin to Sway with testuser
        services.greetd.enable = true;
        services.greetd.settings.initial_session = {
          user = "testuser";
          command = "${pkgs.sway}/bin/sway";
        };

        # Run the check in the sway config on startup and log output.
        # In nixos tests, getting sytemd user services to work is tricky,
        # so we run graphical session and run the check directly from sway config
        environment.etc."sway/config.d/99-test.conf".text = ''
          exec paretosecurity check --only 37dee029-605b-4aab-96b9-5438e5aa44d8 > /tmp/paretosecurity-check.log 2>&1; swaymsg exit
        '';
      };

    swaylock =
      { pkgs, ... }:
      {
        services.paretosecurity.enable = true;

        programs.sway.enable = true;
        programs.sway.wrapperFeatures.gtk = true;

        users.users.testuser.isNormalUser = true;
        users.users.testuser.extraGroups = [
          "wheel"
          "video"
          "input"
        ];

        # Autologin to Sway with testuser
        services.greetd.enable = true;
        services.greetd.settings.initial_session = {
          user = "testuser";
          command = "${pkgs.sway}/bin/sway";
        };

        # Swayidle user service that locks/suspends using swaylock
        systemd.user.services.swayidle = {
          description = "Idle management daemon for Wayland";
          documentation = [ "man:swayidle(1)" ];
          wantedBy = [ "sway-session.target" ];
          partOf = [ "graphical-session.target" ];
          after = [ "graphical-session.target" ];
          serviceConfig = {
            Type = "simple";
            ExecStart =
              "${pkgs.swayidle}/bin/swayidle -w "
              + "timeout 300 '${pkgs.swaylock}/bin/swaylock -f' "
              + "timeout 600 'systemctl suspend' "
              + "before-sleep '${pkgs.swaylock}/bin/swaylock -f' "
              + "lock '${pkgs.swaylock}/bin/swaylock -f'";
            Restart = "on-failure";
            RestartSec = 1;
          };
        };

        # Run the check in the sway config on startup and log output.
        # In nixos tests, getting sytemd user services to work is tricky,
        # so we run graphical session and run the check directly from sway config
        environment.etc."sway/config.d/99-test.conf".text = ''
          exec paretosecurity check --only 37dee029-605b-4aab-96b9-5438e5aa44d8 > /tmp/paretosecurity-check.log 2>&1; swaymsg exit
        '';
      };
  };

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
    kde.succeed("kwriteconfig5 --file kscreenlockerrc --group Daemon --key LockOnResume true")
    out = kde.succeed("paretosecurity check --only 37dee029-605b-4aab-96b9-5438e5aa44d8")
    expected = (
        "  • Starting checks...\n"
        "  • Access Security: Password is required to unlock the screen > [OK] Password after sleep or screensaver is on\n"
        "  • Checks completed.\n"
    )
    assert out == expected, f"Expected did not match actual, got \n{out}"

    # Test 2: Check fails when lock is disabled
    kde.succeed("kwriteconfig5 --file kscreenlockerrc --group Daemon --key LockOnResume false")
    status, out = kde.execute("paretosecurity check --only 37dee029-605b-4aab-96b9-5438e5aa44d8")
    expected = (
        "  • Starting checks...\n"
        "  • Access Security: Password is required to unlock the screen > [FAIL] Password after sleep or screensaver is off\n"
        "  • Checks completed.\n"
    )
    assert out == expected, f"Expected did not match actual, got \n{out}"

    # Test Sway
    # Test 1: Check fails when swaylock is not configured (default sway without swayidle)
    # Wait for system to boot and auto-login to happen
    sway.wait_for_unit("multi-user.target")

    # Wait for the log file to be created by Sway startup script
    sway.wait_for_file("/tmp/paretosecurity-check.log", timeout=30)
    sway.sleep(1)  # Ensure the file is fully written

    # Read the log file
    out = sway.succeed("cat /tmp/paretosecurity-check.log")
    print(f"Sway (no swaylock) check output:\n{out}")

    # The check should fail because swayidle is not configured
    expected = (
        "  • Starting checks...\n"
        "  • Access Security: Password is required to unlock the screen > [FAIL] Password after sleep or screensaver is off\n"
        "  • Checks completed.\n"
    )
    assert out == expected, f"Expected did not match actual, got \n{out}"

    # Test 2: Check passes when swaylock is configured via swayidle service
    # Wait for system to boot and auto-login to happen
    swaylock.wait_for_unit("multi-user.target")

    # Wait for the log file to be created by Sway startup script
    swaylock.wait_for_file("/tmp/paretosecurity-check.log", timeout=30)
    swaylock.sleep(1)  # Ensure the file is fully written

    # Read the log file
    out = swaylock.succeed("cat /tmp/paretosecurity-check.log")
    print(f"Sway check output:\n{out}")

    # The check should pass because swayidle is configured with swaylock
    expected = (
        "  • Starting checks...\n"
        "  • Access Security: Password is required to unlock the screen > [OK] Password after sleep or screensaver is on\n"
        "  • Checks completed.\n"
    )
    assert out == expected, f"Expected did not match actual, got \n{out}"
  '';
}
