let
  common = import ./common.nix;
  inherit (common) users pareto displayManager ssh;
in {
  name = "Trayicon";

  nodes = {
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
      services.displayManager.defaultSession = "gnome";

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
      services.displayManager.defaultSession = "plasma5";
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
      services.displayManager.defaultSession = "none+i3";

      # Enable Home Manager for alice
      users.users.alice.isNormalUser = true;
      home-manager.users.alice = {
        services.status-notifier-watcher.enable = true;
      };

      environment.systemPackages = with pkgs; [
        i3
        dbus
        home-manager
      ];
    };

    # XFCE desktop environment
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
      services.displayManager.defaultSession = "xfce";

      services.displayManager.autoLogin = {
        enable = true;
        user = "alice";
      };

      environment.systemPackages = with pkgs; [
        xfce.xfce4-whiskermenu-plugin
        dbus
        snixembed # Provides StatusNotifierWatcher for XFCE
      ];

      programs.thunar.plugins = [pkgs.xfce.thunar-archive-plugin];
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
      services.displayManager.defaultSession = "none+i3";

      environment.systemPackages = with pkgs; [
        i3
        dbus
      ];
    };
  };

  testScript = {nodes, ...}: let
    user = nodes.gnome.users.users.alice;
    bus = "DBUS_SESSION_BUS_ADDRESS=unix:path=/run/user/${toString user.uid}/bus";
  in ''
    with subtest("GNOME with AppIndicator"):
      gnome.wait_for_unit("multi-user.target")
      gnome.wait_for_x()

      # Enable AppIndicator extension
      gnome.succeed("su - alice -c 'gnome-extensions enable appindicatorsupport@rgcjonas.gmail.com'")

      # Wait for GNOME shell to load
      gnome.wait_for_unit("gnome-session.target", "alice")

      # Check if StatusNotifierWatcher is available
      gnome.succeed("su - alice -c 'DISPLAY=:0 ${bus} dbus-send --session --dest=org.freedesktop.DBus --type=method_call --print-reply /org/freedesktop/DBus org.freedesktop.DBus.ListNames | grep -q StatusNotifierWatcher'")

      # Test trayicon command starts without immediate error
      gnome.succeed("timeout 5s su - alice -c 'DISPLAY=:0 ${bus} paretosecurity trayicon &'")

    with subtest("KDE with native StatusNotifierItem support"):
      kde.wait_for_unit("multi-user.target")
      kde.wait_for_x()

      # Wait for KDE to fully load
      kde.wait_for_unit("plasma-plasmashell.service", "alice")

      # Check if StatusNotifierWatcher is available (KDE provides this natively)
      kde.succeed("su - alice -c 'DISPLAY=:0 ${bus} dbus-send --session --dest=org.freedesktop.DBus --type=method_call --print-reply /org/freedesktop/DBus org.freedesktop.DBus.ListNames | grep -q StatusNotifierWatcher'")

      # Test trayicon command starts without immediate error
      kde.succeed("timeout 5s su - alice -c 'DISPLAY=:0 ${bus} paretosecurity trayicon &'")

    with subtest("Home Manager with status-notifier-watcher service"):
      homemanager.wait_for_unit("multi-user.target")
      homemanager.wait_for_x()

      # Wait for tray.target and status-notifier-watcher service to start
      homemanager.wait_for_unit("tray.target", "alice")
      homemanager.wait_for_unit("status-notifier-watcher.service", "alice")

      # Check if StatusNotifierWatcher is available
      homemanager.succeed("su - alice -c 'DISPLAY=:0 ${bus} dbus-send --session --dest=org.freedesktop.DBus --type=method_call --print-reply /org/freedesktop/DBus org.freedesktop.DBus.ListNames | grep -q StatusNotifierWatcher'")

      # Test trayicon command starts without immediate error
      homemanager.succeed("timeout 5s su - alice -c 'DISPLAY=:0 ${bus} paretosecurity trayicon &'")

    with subtest("XFCE desktop environment"):
      xfce.wait_for_unit("multi-user.target")
      xfce.wait_for_x()

      # Wait for XFCE to fully load
      xfce.wait_for_file("/run/user/${toString user.uid}/bus")
      xfce.wait_for_window("xfce4-panel")
      xfce.wait_for_window("Desktop")

      # Check if XFCE components are running
      for component in ["xfwm4", "xfsettingsd", "xfdesktop", "xfce4-notifyd", "xfconfd"]:
        xfce.wait_until_succeeds(f"pgrep -f {component}")

      # Start snixembed to provide StatusNotifierWatcher
      xfce.succeed("su - alice -c 'DISPLAY=:0 ${bus} snixembed >&2 &'")
      xfce.sleep(2)  # Give snixembed time to start

      # Check if StatusNotifierWatcher is available
      xfce.succeed("su - alice -c 'DISPLAY=:0 ${bus} dbus-send --session --dest=org.freedesktop.DBus --type=method_call --print-reply /org/freedesktop/DBus org.freedesktop.DBus.ListNames | grep -q StatusNotifierWatcher'")

      # Test trayicon command starts without immediate error
      xfce.succeed("timeout 5s su - alice -c 'DISPLAY=:0 ${bus} paretosecurity trayicon &'")

    with subtest("Minimal desktop environment failure handling"):
      minimal.wait_for_unit("multi-user.target")
      minimal.wait_for_x()

      # Wait for i3 to start
      minimal.wait_for_unit("i3.service", "alice")

      # Check that StatusNotifierWatcher is NOT available
      status, out = minimal.execute("su - alice -c 'DISPLAY=:0 ${bus} dbus-send --session --dest=org.freedesktop.DBus --type=method_call --print-reply /org/freedesktop/DBus org.freedesktop.DBus.ListNames | grep StatusNotifierWatcher'")
      assert status != 0, f"StatusNotifierWatcher should not be available in minimal environment, but got: {out}"

      # Test trayicon command fails gracefully and shows error message
      status, out = minimal.execute("su - alice -c 'DISPLAY=:0 ${bus} paretosecurity trayicon 2>&1'")
      assert status != 0, f"Trayicon command should fail in minimal environment, but got exit code: {status}"
      assert "StatusNotifierWatcher not found" in out, f"Expected error message not found in output: {out}"
      assert "gnome-shell-extension-appindicator" in out, f"Expected GNOME solution not found in output: {out}"
      assert "snixembed" in out, f"Expected snixembed solution not found in output: {out}"
      assert "services.status-notifier-watcher in Home Manager" in out, f"Expected Home Manager solution not found in output: {out}"
      assert "waybar with tray support enabled" in out, f"Expected waybar solution not found in output: {out}"
      assert "https://paretosecurity.com/docs/linux/trayicon" in out, f"Expected documentation URL not found in output: {out}"
  '';
}
