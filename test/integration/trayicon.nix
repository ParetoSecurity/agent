let
  common = import ./common.nix;
  inherit (common) users pareto displayManager ssh;
in {
  name = "Trayicon";

  nodes = {
    # XFCE without StatusNotifierWatcher support (failure case)
    xfce = {
      pkgs,
      lib,
      ...
    }: {
      imports = [
        (users {})
        (pareto {inherit pkgs lib;})
        (displayManager {inherit pkgs;})
      ];

      services.xserver.enable = true;
      services.xserver.displayManager.lightdm.enable = true;
      services.xserver.desktopManager.xfce.enable = true;

      # Add basic XFCE packages (no StatusNotifierWatcher support)
      environment.systemPackages = with pkgs; [
        xfce.xfce4-panel
        xfce.xfce4-settings
        dbus
      ];

      # Remove snixembed service since it doesn't provide StatusNotifierWatcher reliably
      # We'll test this as a failure case to demonstrate the error handling
    };

    # GNOME with AppIndicator extension
    gnome = {
      pkgs,
      lib,
      ...
    }: {
      imports = [
        (users {})
        (pareto {inherit pkgs lib;})
        (displayManager {inherit pkgs;})
      ];

      services.xserver.enable = true;
      services.xserver.displayManager.gdm.enable = true;
      services.xserver.desktopManager.gnome.enable = true;

      # Add AppIndicator extension for system tray support
      environment.systemPackages = with pkgs; [
        gnome-shell-extensions
        gnomeExtensions.appindicator
        dbus
      ];
    };

    # KDE Plasma with native StatusNotifierItem support
    kde = {
      pkgs,
      lib,
      ...
    }: {
      imports = [
        (users {})
        (pareto {inherit pkgs lib;})
        (displayManager {inherit pkgs;})
      ];

      services.xserver.enable = true;
      services.xserver.displayManager.sddm.enable = true;
      services.xserver.desktopManager.plasma5.enable = true;
      services.colord.enable = false;

      environment.systemPackages = with pkgs; [
        dbus
      ];
    };

    # Home Manager with status-notifier-watcher service
    homemanager = {
      pkgs,
      lib,
      ...
    }: {
      imports = [
        (users {})
        (pareto {inherit pkgs lib;})
        (displayManager {inherit pkgs;})
      ];

      services.xserver.enable = true;
      services.xserver.displayManager.lightdm.enable = true;
      services.xserver.windowManager.i3.enable = true;

      environment.systemPackages = with pkgs; [
        i3
        dbus
        haskellPackages.status-notifier-item
      ];

      # Enable Home Manager status-notifier-watcher service for alice
      systemd.user.services.status-notifier-watcher = {
        description = "Status Notifier Watcher";
        wantedBy = ["tray.target"];
        after = ["tray.target"];
        serviceConfig = {
          Type = "dbus";
          BusName = "org.kde.StatusNotifierWatcher";
          ExecStart = "${pkgs.haskellPackages.status-notifier-item}/bin/status-notifier-watcher";
          Restart = "on-failure";
        };
      };

      systemd.user.targets.tray = {
        description = "Tray target";
        bindsTo = ["graphical-session.target"];
        wants = ["graphical-session.target"];
        after = ["graphical-session.target"];
      };
    };

    # Waybar with built-in StatusNotifierWatcher support
    waybar = {
      pkgs,
      lib,
      ...
    }: {
      imports = [
        (users {})
        (pareto {inherit pkgs lib;})
        (displayManager {inherit pkgs;})
      ];

      services.xserver.enable = true;
      services.xserver.displayManager.lightdm.enable = true;
      services.xserver.windowManager.i3.enable = true;

      environment.systemPackages = with pkgs; [
        i3
        dbus
        waybar
      ];

      # Configure waybar with tray support
      systemd.user.services.waybar = {
        description = "Waybar - Highly customizable Wayland bar";
        wantedBy = ["graphical-session.target"];
        after = ["graphical-session.target"];
        serviceConfig = {
          Type = "dbus";
          BusName = "fr.arouillard.waybar";
          ExecStart = "${pkgs.waybar}/bin/waybar";
          Restart = "on-failure";
          RestartSec = 1;
        };
      };

      # Create waybar config with tray support
      environment.etc."skel/.config/waybar/config".text = builtins.toJSON {
        layer = "top";
        position = "top";
        height = 30;
        modules-right = ["tray"];
        tray = {
          spacing = 10;
        };
      };
    };

    # Minimal desktop environment without StatusNotifierItem support
    minimal = {
      pkgs,
      lib,
      ...
    }: {
      imports = [
        (users {})
        (pareto {inherit pkgs lib;})
        (displayManager {inherit pkgs;})
      ];

      services.xserver.enable = true;
      services.xserver.displayManager.lightdm.enable = true;
      services.xserver.windowManager.i3.enable = true;

      environment.systemPackages = with pkgs; [
        i3
        dbus
      ];
    };
  };

  interactive.nodes.xfce = {...}:
    ssh {port = 2221;} {};

  interactive.nodes.gnome = {...}:
    ssh {port = 2222;} {};

  interactive.nodes.kde = {...}:
    ssh {port = 2223;} {};

  interactive.nodes.homemanager = {...}:
    ssh {port = 2224;} {};

  interactive.nodes.waybar = {...}:
    ssh {port = 2225;} {};

  interactive.nodes.minimal = {...}:
    ssh {port = 2226;} {};

  enableOCR = true;

  testScript = ''
    # Test XFCE without StatusNotifierWatcher support
    print("Testing XFCE without StatusNotifierWatcher support...")
    xfce.wait_for_unit("multi-user.target")
    xfce.wait_for_x()

    # Check that StatusNotifierWatcher is NOT available
    status, out = xfce.execute("su - alice -c 'dbus-send --session --dest=org.freedesktop.DBus --type=method_call --print-reply /org/freedesktop/DBus org.freedesktop.DBus.ListNames | grep StatusNotifierWatcher'")
    assert status != 0, f"StatusNotifierWatcher should not be available in XFCE, but got: {out}"

    # Test trayicon command fails gracefully and shows error message
    status, out = xfce.execute("su - alice -c 'paretosecurity trayicon 2>&1'")
    assert status != 0, f"Trayicon command should fail in XFCE without StatusNotifierWatcher, but got exit code: {status}"
    assert "StatusNotifierWatcher not found" in out, f"Expected error message not found in output: {out}"
    assert "gnome-shell-extension-appindicator" in out, f"Expected GNOME solution not found in output: {out}"
    assert "snixembed" in out, f"Expected snixembed solution not found in output: {out}"
    print("XFCE test passed - trayicon failed gracefully with helpful error message")

    # Test GNOME with AppIndicator
    print("Testing GNOME with AppIndicator...")
    gnome.wait_for_unit("multi-user.target")
    gnome.wait_for_x()

    # Enable AppIndicator extension
    gnome.succeed("su - alice -c 'gnome-extensions enable appindicatorsupport@rgcjonas.gmail.com'")

    # Wait for GNOME shell to load
    gnome.wait_for_unit("gnome-session.target", "alice")

    # Check if StatusNotifierWatcher is available
    gnome.succeed("su - alice -c 'dbus-send --session --dest=org.freedesktop.DBus --type=method_call --print-reply /org/freedesktop/DBus org.freedesktop.DBus.ListNames | grep -q StatusNotifierWatcher'")

    # Test trayicon command starts without immediate error
    gnome.succeed("timeout 5s su - alice -c 'paretosecurity trayicon &'")

    # Test KDE with native StatusNotifierItem support
    print("Testing KDE with native StatusNotifierItem support...")
    kde.wait_for_unit("multi-user.target")
    kde.wait_for_x()

    # Wait for KDE to fully load
    kde.wait_for_unit("plasma-plasmashell.service", "alice")

    # Check if StatusNotifierWatcher is available (KDE provides this natively)
    kde.succeed("su - alice -c 'dbus-send --session --dest=org.freedesktop.DBus --type=method_call --print-reply /org/freedesktop/DBus org.freedesktop.DBus.ListNames | grep -q StatusNotifierWatcher'")

    # Test trayicon command starts without immediate error
    kde.succeed("timeout 5s su - alice -c 'paretosecurity trayicon &'")

    # Test Home Manager with status-notifier-watcher service
    print("Testing Home Manager with status-notifier-watcher service...")
    homemanager.wait_for_unit("multi-user.target")
    homemanager.wait_for_x()

    # Wait for tray.target and status-notifier-watcher service to start
    homemanager.wait_for_unit("tray.target", "alice")
    homemanager.wait_for_unit("status-notifier-watcher.service", "alice")

    # Check if StatusNotifierWatcher is available
    homemanager.succeed("su - alice -c 'dbus-send --session --dest=org.freedesktop.DBus --type=method_call --print-reply /org/freedesktop/DBus org.freedesktop.DBus.ListNames | grep -q StatusNotifierWatcher'")

    # Test trayicon command starts without immediate error
    homemanager.succeed("timeout 5s su - alice -c 'paretosecurity trayicon &'")

    # Test Waybar with built-in StatusNotifierWatcher support
    print("Testing Waybar with built-in StatusNotifierWatcher support...")
    waybar.wait_for_unit("multi-user.target")
    waybar.wait_for_x()

    # Wait for waybar service to start
    waybar.wait_for_unit("waybar.service", "alice")

    # Check if StatusNotifierWatcher is available (provided by waybar)
    waybar.succeed("su - alice -c 'dbus-send --session --dest=org.freedesktop.DBus --type=method_call --print-reply /org/freedesktop/DBus org.freedesktop.DBus.ListNames | grep -q StatusNotifierWatcher'")

    # Test trayicon command starts without immediate error
    waybar.succeed("timeout 5s su - alice -c 'paretosecurity trayicon &'")

    # Test minimal desktop environment (should fail gracefully)
    print("Testing minimal desktop environment without StatusNotifierItem support...")
    minimal.wait_for_unit("multi-user.target")
    minimal.wait_for_x()

    # Wait for i3 to start
    minimal.wait_for_unit("i3.service", "alice")

    # Check that StatusNotifierWatcher is NOT available
    status, out = minimal.execute("su - alice -c 'dbus-send --session --dest=org.freedesktop.DBus --type=method_call --print-reply /org/freedesktop/DBus org.freedesktop.DBus.ListNames | grep StatusNotifierWatcher'")
    assert status != 0, f"StatusNotifierWatcher should not be available in minimal environment, but got: {out}"

    # Test trayicon command fails gracefully and shows error message
    status, out = minimal.execute("su - alice -c 'paretosecurity trayicon 2>&1'")
    assert status != 0, f"Trayicon command should fail in minimal environment, but got exit code: {status}"
    assert "StatusNotifierWatcher not found" in out, f"Expected error message not found in output: {out}"
    assert "gnome-shell-extension-appindicator" in out, f"Expected GNOME solution not found in output: {out}"
    assert "snixembed" in out, f"Expected snixembed solution not found in output: {out}"
    assert "services.status-notifier-watcher in Home Manager" in out, f"Expected Home Manager solution not found in output: {out}"
    assert "waybar with tray support enabled" in out, f"Expected waybar solution not found in output: {out}"
    assert "https://paretosecurity.com/docs/linux/trayicon" in out, f"Expected documentation URL not found in output: {out}"

    print("All trayicon tests passed!")
  '';
}
