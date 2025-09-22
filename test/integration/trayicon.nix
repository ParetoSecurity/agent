let
  common = import ./common.nix;
  inherit (common) users displayManager;
in
{
  name = "Trayicon";

  nodes = {
    # GNOME with AppIndicator extension
    gnome =
      { pkgs, ... }:
      {
        imports = [
          (users { })
          (displayManager { inherit pkgs; })
        ];
        services.paretosecurity.enable = true;

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

    # Minimal desktop environment without StatusNotifierItem support
    minimal =
      { pkgs, ... }:
      {
        imports = [
          (users { })
          (displayManager { inherit pkgs; })
        ];
        services.paretosecurity.enable = true;

        services.xserver.enable = true;
        services.xserver.displayManager.lightdm.enable = true;
        services.xserver.windowManager.i3.enable = true;
        services.displayManager.defaultSession = "none+i3";

        environment.systemPackages = with pkgs; [
          i3
        ];
      };
  };

  testScript =
    { nodes, ... }:
    let
      user = nodes.gnome.users.users.alice;
      bus = "DBUS_SESSION_BUS_ADDRESS=unix:path=/run/user/${toString user.uid}/bus";
    in
    ''
      with subtest("GNOME with AppIndicator"):
        gnome.start()
        gnome.wait_for_unit("multi-user.target")
        gnome.wait_for_x()

        # Enable AppIndicator extension
        gnome.succeed("su - alice -c 'gnome-extensions enable appindicatorsupport@rgcjonas.gmail.com'")

        # Wait for GNOME shell to load
        gnome.wait_for_unit("gnome-session.target", "alice")

        # Test trayicon command starts without immediate error
        gnome.succeed("timeout 5s su - alice -c 'DISPLAY=:0 ${bus} paretosecurity trayicon &'")

      # Shutdown GNOME before starting KDE
      gnome.shutdown()

      with subtest("Minimal desktop environment failure handling"):
        minimal.start()
        minimal.wait_for_unit("multi-user.target")
        minimal.wait_for_x()

        # Test trayicon command fails gracefully and shows error message
        minimal.execute("su - alice -c 'DISPLAY=:0 ${bus} systemctl --user stop paretosecurity-tray.service'")
        minimal.execute("systemctl stop paretosecurity.service")
        minimal.execute("pkill -9 paretosecurity-tray")
        print(minimal.execute("sleep 5 && ps axf")) # give some time to shutdown
        status, out = minimal.execute("su - alice -c 'DISPLAY=:0 ${bus} paretosecurity trayicon 2>&1'")
        assert status != 0, f"Trayicon command should fail in minimal environment, but got exit code: {status}"
        assert "StatusNotifierWatcher not found" in out, f"Expected error message not found in output: {out}"

      # Shutdown minimal when done
      minimal.shutdown()
    '';
}
