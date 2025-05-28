let
  common = import ../common.nix;
  inherit (common) users paretoPatchedDash dashboard displayManager ssh;
in {
  name = "GNOME";

  nodes.dashboard = {
    imports = [
      (dashboard {})
    ];
  };

  nodes.gnome = {
    pkgs,
    lib,
    ...
  }: {
    imports = [
      (users {})
      (paretoPatchedDash {inherit pkgs lib;})
      (displayManager {inherit pkgs;})
    ];

    virtualisation.resolution = {
      x = 800;
      y = 600;
    };

    services.xserver.displayManager.lightdm.enable = true;
    services.xserver.displayManager.gdm.wayland = false;
    services.xserver.desktopManager.gnome.enable = true;

    environment.systemPackages = [pkgs.gnomeExtensions.appindicator];
    services.udev.packages = [pkgs.gnome-settings-daemon];

    # Enable AppIndicator extension for tray icons
    programs.dconf.profiles.user.databases = [
      {
        settings = {
          "org/gnome/shell" = {
            enabled-extensions = ["appindicatorsupport@rgcjonas.gmail.com"];
          };
        };
      }
    ];

    environment.gnome.excludePackages = with pkgs; [
      atomix # puzzle game
      cheese # webcam tool
      epiphany # web browser
      evince # document viewer
      geary # email reader
      gedit # text editor
      gnome-characters
      gnome-music
      gnome-photos
      gnome-terminal
      gnome-tour
      hitori # sudoku game
      iagno # go game
      tali # poker game
      totem # video player
    ];
  };

  interactive.nodes.gnome = {...}:
    ssh {port = 2221;} {};

  enableOCR = true;

  testScript = ''
    # Test setup
    for m in [gnome, dashboard]:
      m.systemctl("start network-online.target")
      m.wait_for_unit("network-online.target")

    gnome.wait_for_x()

    # Test: Test the tray icon
    for unit in [
        'paretosecurity-trayicon',
        'paretosecurity-user',
        'paretosecurity-user.timer'
    ]:
        status, out = gnome.systemctl("is-enabled " + unit, "alice")
        assert status == 0, f"Unit {unit} is not enabled (status: {status}): {out}"

    gnome.succeed("xdotool mousemove 670 10")
    gnome.succeed("sleep 3")
    gnome.succeed("xdotool click 1")
    gnome.wait_for_text("Run Checks", 30)

    # Hide the tray menu
    gnome.succeed("xdotool mousemove 10 100")
    gnome.succeed("xdotool click 1")

    # Test: Pareto Desktop entry
    gnome.succeed("xdotool key Super_r")
    gnome.succeed("sleep 3")
    gnome.wait_for_text("Type to search", 30)
    gnome.succeed("xdotool type 'Pareto'")
    gnome.wait_for_text("Pareto Security", 30)

    # Test: paretosecurity:// URL handler is registered
    gnome.execute("su - alice -c 'xdg-open paretosecurity://foo >/dev/null &'")
    gnome.wait_for_text("Failed to add device")
  '';
}
