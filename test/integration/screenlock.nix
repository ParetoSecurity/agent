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
        services.greetd.settings.default_session = {
          user = "testuser";
          command = "${pkgs.sway}/bin/sway";
        };

        # Import Home Manager module
        imports = [
          "${
            builtins.fetchTarball {
              url = "https://github.com/nix-community/home-manager/archive/a504aee7d452ecf8881b933b1eb573535a543a03.tar.gz";
              sha256 = "1578zlv3p77djrc7vfrs0fps2xr6a9cxfqh2habvspai8r6hr2p1";
            }
          }/nixos"
        ];

        # Configure Home Manager for the test user
        home-manager.users.testuser =
          { ... }:
          {
            home.stateVersion = "25.11";
            home.username = "testuser";
            home.homeDirectory = "/home/testuser";

            # Swayidle disabled - no screen locking configured
            services.swayidle = {
              enable = false;
            };

            # Swaylock disabled
            programs.swaylock = {
              enable = false;
            };
          };

        # In nixos tests, getting systemd user services to work is tricky,
        # so we run graphical session and run the check directly from sway config
        environment.etc."sway/config.d/99-test.conf".text = ''
          exec sh -c "paretosecurity check --only 37dee029-605b-4aab-96b9-5438e5aa44d8 > /tmp/paretosecurity-check.log 2>&1; swaymsg exit"
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
        services.greetd.settings.default_session = {
          user = "testuser";
          command = "${pkgs.sway}/bin/sway";
        };

        # Import Home Manager module
        imports = [
          "${
            builtins.fetchTarball {
              url = "https://github.com/nix-community/home-manager/archive/a504aee7d452ecf8881b933b1eb573535a543a03.tar.gz";
              sha256 = "1578zlv3p77djrc7vfrs0fps2xr6a9cxfqh2habvspai8r6hr2p1";
            }
          }/nixos"
        ];

        # Configure Home Manager for the test user
        home-manager.users.testuser =
          { ... }:
          {
            home.stateVersion = "25.11";
            home.username = "testuser";
            home.homeDirectory = "/home/testuser";

            # Configure swayidle with swaylock using Home Manager
            services.swayidle = {
              enable = true;
              timeouts = [
                {
                  timeout = 300;
                  command = "${pkgs.swaylock}/bin/swaylock -f";
                }
                {
                  timeout = 600;
                  command = "systemctl suspend";
                }
              ];
              events = [
                {
                  event = "before-sleep";
                  command = "${pkgs.swaylock}/bin/swaylock -f";
                }
                {
                  event = "lock";
                  command = "${pkgs.swaylock}/bin/swaylock -f";
                }
              ];
            };

            # Configure swaylock
            programs.swaylock = {
              enable = true;
            };
          };

        # In nixos tests, getting systemd user services to work is tricky,
        # so we run graphical session and run the check directly from sway config
        environment.etc."sway/config.d/99-test.conf".text = ''
          exec sh -c "paretosecurity check --only 37dee029-605b-4aab-96b9-5438e5aa44d8 > /tmp/paretosecurity-check.log 2>&1; swaymsg exit"
        '';
      };
  };

  testScript = ''
    # Test GNOME
    # Test 1: Check passes by default
    out = gnome.succeed("paretosecurity check --only 37dee029-605b-4aab-96b9-5438e5aa44d8")
    assert "[OK] Password after sleep or screensaver is on" in out, f"Expected check to pass, got \n{out}"

    # Test 2: Check fails when lock is disabled
    gnome.succeed("dbus-run-session -- gsettings set org.gnome.desktop.screensaver lock-enabled false")
    status, out = gnome.execute("paretosecurity check --only 37dee029-605b-4aab-96b9-5438e5aa44d8")
    assert "[FAIL] Password after sleep or screensaver is off" in out, f"Expected check to fail, got \n{out}"

    # Test KDE
    # Test 1: Check passes with lock enabled
    kde.succeed("kwriteconfig5 --file kscreenlockerrc --group Daemon --key LockOnResume true")
    out = kde.succeed("paretosecurity check --only 37dee029-605b-4aab-96b9-5438e5aa44d8")
    assert "[OK] Password after sleep or screensaver is on" in out, f"Expected check to pass, got \n{out}"

    # Test 2: Check fails when lock is disabled
    kde.succeed("kwriteconfig5 --file kscreenlockerrc --group Daemon --key LockOnResume false")
    status, out = kde.execute("paretosecurity check --only 37dee029-605b-4aab-96b9-5438e5aa44d8")
    assert "[FAIL] Password after sleep or screensaver is off" in out, f"Expected check to fail, got \n{out}"

    # Test Sway
    # Test 1: Check fails when swaylock is not configured (default sway without swayidle)
    # Wait for system to boot and auto-login to happen
    sway.wait_for_unit("multi-user.target")

    # Wait for the log file to be created by Sway startup script
    sway.wait_for_file("/tmp/paretosecurity-check.log", timeout=30)
    sway.sleep(3)

    # Read the log file
    out = sway.succeed("cat /tmp/paretosecurity-check.log")
    print(f"Sway (no swaylock) check output:\n{out}")

    # The check should fail because swayidle is not configured
    # Check that the output contains the expected failure message
    assert "[FAIL] Password after sleep or screensaver is off" in out, f"Expected check to fail, got \n{out}"

    # Test 2: Check passes when swaylock is configured via swayidle service
    # Wait for system to boot and auto-login to happen
    swaylock.wait_for_unit("multi-user.target")

    # Wait for the log file to be created by Sway startup script
    swaylock.wait_for_file("/tmp/paretosecurity-check.log", timeout=30)
    sway.sleep(3)

    # Read the log file
    out = swaylock.succeed("cat /tmp/paretosecurity-check.log")
    print(f"Sway (with swaylock) check output:\n{out}")

    # The check should pass because swayidle is configured with swaylock
    # Check that the output contains the expected success message
    assert "[OK] Password after sleep or screensaver is on" in out, f"Expected check to pass, got \n{out}"
  '';
}
