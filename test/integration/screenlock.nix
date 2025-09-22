{
  name = "Screen Lock";
  interactive.sshBackdoor.enable = true;

  nodes = {
    gnome =
      { ... }:
      {
        services.paretosecurity.enable = true;
        # Install GNOME Desktop Environment
        services.desktopManager.gnome.enable = true;
        services.displayManager.gdm.enable = true;
      };

    kde =
      { pkgs, ... }:
      {
        services.paretosecurity.enable = true;
        # Install KDE Plasma 6 Desktop Environment
        services.xserver.enable = true;
        services.desktopManager.plasma6.enable = true;
        services.displayManager.sddm.enable = true;
        services.displayManager.autoLogin.enable = true;
        services.displayManager.autoLogin.user = "alice";
        services.colord.enable = false;

        # Increase memory for Plasma 6
        virtualisation.memorySize = 2048;

        # Create alice user
        users.users.alice = {
          isNormalUser = true;
          extraGroups = [ "wheel" ];
          password = "alice";
        };

        # Add kconfig package which includes kwriteconfig6 and kreadconfig6
        environment.systemPackages = with pkgs; [
          kdePackages.kconfig
        ];
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

    # Test KDE Plasma 6
    # Wait for KDE to start and alice to be logged in
    kde.wait_for_unit("graphical.target")

    # First ensure lock is enabled (Plasma 6 might have different defaults)
    # Run as alice user
    kde.succeed("su - alice -c 'kwriteconfig6 --file kscreenlockerrc --group Daemon --key LockOnResume true'")
    kde.succeed("su - alice -c 'kwriteconfig6 --file kscreenlockerrc --group Daemon --key Autolock true'")

    # Test 1: Check passes with lock enabled (run as alice)
    out = kde.succeed("su - alice -c 'paretosecurity check --only 37dee029-605b-4aab-96b9-5438e5aa44d8'")
    assert "[OK] Password after sleep or screensaver is on" in out, f"Expected check to pass, got \n{out}"

    # Test 2: Check fails when lock is disabled
    kde.succeed("su - alice -c 'kwriteconfig6 --file kscreenlockerrc --group Daemon --key LockOnResume false'")
    kde.succeed("su - alice -c 'kwriteconfig6 --file kscreenlockerrc --group Daemon --key Autolock false'")
    status, out = kde.execute("su - alice -c 'paretosecurity check --only 37dee029-605b-4aab-96b9-5438e5aa44d8'")
    assert "[FAIL] Password after sleep or screensaver is off" in out, f"Expected check to fail, got \n{out}"

    # Test Sway
    # Test 1: Check fails when swaylock is not configured (default sway without swayidle)
    # Wait for system to boot and auto-login to happen
    sway.wait_for_unit("multi-user.target")

    # Wait for the log file to be created by Sway startup script
    sway.wait_for_file("/tmp/paretosecurity-check.log", timeout=30)
    sway.wait_until_succeeds("test $(wc -l < /tmp/paretosecurity-check.log) -ge 3", timeout=30)

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
    swaylock.wait_until_succeeds("test $(wc -l < /tmp/paretosecurity-check.log) -ge 3", timeout=30)

    # Read the log file
    out = swaylock.succeed("cat /tmp/paretosecurity-check.log")
    print(f"Sway (with swaylock) check output:\n{out}")

    # The check should pass because swayidle is configured with swaylock
    # Check that the output contains the expected success message
    assert "[OK] Password after sleep or screensaver is on" in out, f"Expected check to pass, got \n{out}"
  '';
}
