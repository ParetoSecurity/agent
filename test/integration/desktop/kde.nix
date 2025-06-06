let
  common = import ../common.nix;
  inherit (common) users paretoPatchedDash dashboard displayManager ssh;
in {
  name = "KDE";

  nodes.terminal = {
    pkgs,
    lib,
    ...
  }: {
    imports = [
      (users {})
      (paretoPatchedDash {inherit pkgs lib;})
    ];

    networking.firewall.enable = true;
  };

  nodes.dashboard = {
    imports = [
      (dashboard {})
    ];
  };

  nodes.kde = {
    pkgs,
    lib,
    ...
  }: {
    imports = [
      (users {})
      (paretoPatchedDash {inherit pkgs lib;})
      (displayManager {inherit pkgs;})
    ];
    environment.systemPackages = [pkgs.xorg.xev];

    programs.ydotool.enable = true;
    services.displayManager.sddm = {
      enable = true;
    };
    services.desktopManager.plasma6.enable = true;

    virtualisation.memorySize = 4096;
    virtualisation.cores = 8;

    time.timeZone = "UTC";
  };

  interactive.nodes.kde = {...}:
    ssh {port = 2221;} {};

  enableOCR = true;

  testScript = {nodes, ...}: let
    user = nodes.kde.users.users.alice;
  in ''
    # Test setup
    for m in [kde, dashboard]:
      m.systemctl("start network-online.target")
      m.wait_for_unit("network-online.target")

    # Test: Test the tray icon
    for unit in [
        'paretosecurity-trayicon',
        'paretosecurity-user',
        'paretosecurity-user.timer'
    ]:
        status, out = kde.systemctl("is-enabled " + unit, "alice")
        assert status == 0, f"Unit {unit} is not enabled (status: {status}): {out}"

    kde.wait_for_unit("display-manager.service")
    kde.wait_for_text("(AM|PM)",  120)

    # Test: Test the tray icon
    kde.execute("ydotool mousemove --absolute --x 0 --y 0")
    kde.execute("ydotool mousemove 210 400")
    kde.succeed("sleep 3")
    kde.execute("ydotool click 0xC0")
    kde.wait_for_text("Run Checks", 30)

    # Hide the tray menu
    kde.execute("ydotool mousemove --absolute --x 0 --y 0")
    kde.execute("ydotool mousemove 10 10")
    kde.execute("ydotool click 0xC0")

    # Test: Pareto Desktop entry
    # `126` is the evdev scancode for KEY_RIGHTMETA (Super_R).
    # Send `press(126):release(126)` in one command.
    kde.succeed("ydotool key 126:1 126:0")
    kde.succeed("sleep 3")
    kde.wait_for_text("Applications", 30)
    kde.succeed("ydotool type 'Pareto'")
    kde.wait_for_text("Pareto Security", 30)

    # Test: paretosecurity:// URL handler is registered
    kde.execute("su - alice -c 'xdg-open paretosecurity://foo >/dev/null &'")
    kde.wait_for_text("Failed to add device")  '';
}
