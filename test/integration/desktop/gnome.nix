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

    services.xserver.enable = true;
    services.xserver.desktopManager.gnome.enable = true;

    environment.systemPackages = [pkgs.gnomeExtensions.appindicator];
    services.udev.packages = [pkgs.gnome-settings-daemon];

    # Enable AppIndicator extension automatically
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
    # terminal.succeed("su - alice -c 'mkdir -p /home/alice/.config'")
    for m in [gnome, dashboard]:
      m.systemctl("start network-online.target")
      m.wait_for_unit("network-online.target")


    # Test 4: Test the tray icon
    gnome.wait_for_x()
    for unit in [
        'paretosecurity-trayicon',
        'paretosecurity-user',
        'paretosecurity-user.timer'
    ]:
        status, out = gnome.systemctl("is-enabled " + unit, "alice")
        assert status == 0, f"Unit {unit} is not enabled (status: {status}): {out}"

    # GNOME system tray is in the top-right corner
    gnome.succeed("xdotool mousemove 670 10")
    gnome.screenshot("1-mouse-at-670-10.png")
    # gnome.succeed("xdotool click 1")
    gnome.succeed("sleep 3")
    gnome.screenshot("2-pareto-trayicon-after-1-click.png")
    gnome.succeed("xdotool click 1")
    # gnome.succeed("sleep 3")
    gnome.screenshot("3-pareto-trayicon-after-2-clicks.png")
    gnome.wait_for_text("Run Checks", timeout=30)

    # Test 5: Desktop entry
    gnome.succeed("xdotool mousemove 10 100")
    gnome.succeed("xdotool click 1")  # hide the tray icon window
    gnome.succeed("xdotool key Super_r")
    gnome.succeed("sleep 3")
    gnome.succeed("xdotool type 'Pareto'")
    gnome.wait_for_text("Pareto Security", timeout=30)

    # Test 6: paretosecurity:// URL handler is registered
    gnome.execute("su - alice -c 'xdg-open paretosecurity://foo >/dev/null &'")
    gnome.wait_for_text("Failed to add device")

  '';
}
