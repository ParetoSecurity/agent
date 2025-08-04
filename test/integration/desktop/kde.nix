let
  common = import ../common.nix;
  inherit (common) users paretoPatchedDash dashboard displayManager ssh;
in {
  name = "KDE";
  interactive.sshBackdoor.enable = true;

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

    services.xserver.enable = true;
    services.displayManager.sddm.enable = true;
    # Use plasmax11 for X11 session (plasma6 would be Wayland)
    services.displayManager.defaultSession = "plasmax11";
    services.desktopManager.plasma6.enable = true;

    # Alternative for Plasma 5
    # services.xserver.desktopManager.plasma5.enable = true;
    # services.displayManager.defaultSession = "plasma";

    services.displayManager.autoLogin = {
      enable = true;
      user = "alice";
    };
  };

  enableOCR = true;

  testScript = {nodes, ...}: let
    user = nodes.kde.users.users.alice;
    xdo = "${pkgs.xdotool}/bin/xdotool";
  in ''
    # Test setup
    for m in [kde, dashboard]:
      m.systemctl("start network-online.target")
      m.wait_for_unit("network-online.target")

    with subtest("Wait for login"):
        start_all()
        kde.wait_for_file("/tmp/xauth_*")
        kde.succeed("xauth merge /tmp/xauth_*")

    with subtest("Check plasmashell started"):
        kde.wait_until_succeeds("pgrep plasmashell")
        kde.wait_for_window("^Desktop ")

    with subtest("Check that KDED is running"):
        kde.succeed("pgrep kded6")

    with subtest("Check that logging in has given the user ownership of devices"):
        kde.succeed("getfacl -p /dev/snd/timer | grep -q ${user.name}")

    with subtest("Check Pareto Security services are enabled"):
        for unit in [
            'paretosecurity-trayicon',
            'paretosecurity-user',
            'paretosecurity-user.timer'
        ]:
            status, out = kde.systemctl("is-enabled " + unit, "${user.name}")
            assert status == 0, f"Unit {unit} is not enabled (status: {status}): {out}"

    kde.succeed("su - ${user.name} -c 'xauth merge /tmp/xauth_*'")

    with subtest("Check system tray for Pareto Security"):
        # Wait for system tray to be ready
        kde.sleep(10)
        # Click on system tray area
        kde.succeed("${xdo} mousemove 950 10")
        kde.succeed("${xdo} click 1")
        kde.wait_for_text("Pareto Security", timeout=60)
        kde.screenshot("kde-tray-pareto")

        # Click on Pareto Security tray icon
        kde.succeed("${xdo} click 1")
        kde.wait_for_text("Run Checks", timeout=30)
        kde.screenshot("kde-tray-menu")

    with subtest("Test Application Launcher"):
        kde.succeed("${xdo} key Escape")  # Close tray menu
        kde.send_key("super")  # Open Application Launcher
        kde.wait_for_text("Search", timeout=30)
        kde.send_chars("pareto")
        kde.wait_for_text("Pareto Security", timeout=30)
        kde.screenshot("kde-launcher-pareto")

    with subtest("Run Dolphin to test desktop integration"):
        kde.execute("su - ${user.name} -c 'DISPLAY=:0.0 dolphin >&2 &'")
        kde.wait_for_window(" Dolphin")

    with subtest("Run Konsole"):
        kde.execute("su - ${user.name} -c 'DISPLAY=:0.0 konsole >&2 &'")
        kde.wait_for_window("Konsole")

    with subtest("Check desktop entry exists"):
        kde.succeed("test -f /home/${user.name}/.local/share/applications/paretosecurity.desktop || test -f /usr/share/applications/paretosecurity.desktop")

    with subtest("Take a screenshot"):
        kde.screenshot("kde-desktop")

    with subtest("Check URL handler registration"):
        # Open Konsole to capture output
        kde.execute("su - ${user.name} -c 'DISPLAY=:0.0 konsole >&2 &'")
        kde.wait_for_window("Konsole")
        kde.wait_for_text(r"(${user.name}|machine)", timeout=30)
        kde.screenshot("kde-terminal-open")

        # Execute URL handler test in terminal to see output
        kde.send_chars("xdg-open paretosecurity://enrollTeam/?token=kde-integration-test-token\n")
        kde.wait_for_text("invite_id not found", timeout=20)
        kde.screenshot("kde-url-handler-result")
        kde.send_key("alt-f4")
  '';
}
