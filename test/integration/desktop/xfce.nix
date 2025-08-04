let
  common = import ../common.nix;
  inherit (common) users paretoPatchedDash dashboard displayManager ssh;
in {
  name = "XFCE";
  interactive.sshBackdoor.enable = true;

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
    xfce.wait_for_text("Pareto Security", timeout=180)
    xfce.succeed("xdotool click 1")
    xfce.wait_for_text("Run Checks", timeout=180)

    # Test: Desktop entry
    xfce.succeed("xdotool mousemove 10 10")
    xfce.succeed("xdotool click 1")  # hide the tray icon window
    xfce.succeed("xdotool click 1")  # show the Applications menu
    xfce.succeed("xdotool mousemove 10 200")
    xfce.succeed("xdotool click 1")
    xfce.wait_for_text("Pareto Security", timeout=20)

    # Test: paretosecurity:// URL handler is registered
    # Open terminal to capture output
    xfce.send_key("ctrl-alt-t")
    xfce.wait_for_text("alice@", timeout=30)
    xfce.screenshot("xfce-terminal-open")

    # Execute URL handler test in terminal to see output
    xfce.send_chars("xdg-open paretosecurity://enrollTeam/?token=xfce-integration-test-token\n")
    xfce.wait_for_text("invite_id not found", timeout=20)
    xfce.screenshot("xfce-url-handler-result")
    xfce.send_key("alt-f4")
  '';
}
