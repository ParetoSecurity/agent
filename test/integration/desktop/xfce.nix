let
  common = import ../common.nix;
  inherit (common) users paretoPatchedDash dashboard displayManager ssh;
in {
  name = "XFCE";

  nodes.dashboard = {
    imports = [
      (dashboard {})
    ];
  };

  nodes.xfce = {
    pkgs,
    lib,
    ...
  }: {
    imports = [
      (users {})
      (paretoPatchedDash {inherit pkgs lib;})
      (displayManager {inherit pkgs;})
    ];

    services.xserver.enable = true;
    services.xserver.displayManager.lightdm.enable = true;
    services.xserver.desktopManager.xfce.enable = true;
  };

  interactive.nodes.xfce = {...}:
    ssh {port = 2221;} {};

  enableOCR = true;

  testScript = ''
    # Test setup
    for m in [xfce, dashboard]:
      m.systemctl("start network-online.target")
      m.wait_for_unit("network-online.target")

    # Test: Tray icon
    xfce.wait_for_x()
    for unit in [
        'paretosecurity-trayicon',
        'paretosecurity-user',
        'paretosecurity-user.timer'
    ]:
        status, out = xfce.systemctl("is-enabled " + unit, "alice")
        assert status == 0, f"Unit {unit} is not enabled (status: {status}): {out}"
    xfce.succeed("xdotool mousemove 630 10")
    xfce.wait_for_text("Pareto Security")
    xfce.succeed("xdotool click 1")
    xfce.wait_for_text("Run Checks")

    # Test: Desktop entry
    # TODO: python function find_icon, that returns x,y position of icon so I know where
    # to move the mouse to
    xfce.succeed("xdotool mousemove 10 10")
    xfce.succeed("xdotool click 1")  # hide the tray icon window
    xfce.succeed("xdotool click 1")  # show the Applications menu
    xfce.succeed("xdotool mousemove 10 200")
    xfce.succeed("xdotool click 1")
    xfce.wait_for_text("Pareto Security", timeout=20)

    # Test: paretosecurity:// URL handler is registered
    xfce.succeed("su - alice -c 'xdg-open"
    + " paretosecurity://enrollTeam/?token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9."
    + "eyJ0b2tlbiI6ImR1bW15LXRva2VuIiwidGVhbUlEIjoiZHVtbXktdGVhbS1pZCIsImlhdCI6"
    + "MTcwMDAwMDAwMCwiZXhwIjoxOTAwMDAwMDAwfQ.WgnL6_S0EBJHwF1wEVUG8GtIcoVvK5IjWbZpUeZr4Qw'"
    + " >/dev/null &")

    xfce.wait_for_text("Device successfully linked", timeout=20)
  '';
}
