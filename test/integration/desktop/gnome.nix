let
  common = import ../common.nix;
  inherit (common) users paretoPatchedDash dashboard displayManager ssh;
in {
  name = "GNOME";
  interactive.sshBackdoor.enable = true;

  nodes.dashboard = {
    imports = [
      (dashboard {})
    ];

    # Minimal resources for dashboard mock server
    virtualisation.memorySize = 512;
    virtualisation.cores = 1;
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

    # Optimize memory usage for desktop tests
    virtualisation.memorySize = 2048; # GNOME needs more memory
    virtualisation.cores = 2; # Limit CPU cores

    services.xserver.enable = true;

    services.displayManager.gdm = {
      enable = true;
      debug = true;
    };

    services.displayManager.autoLogin = {
      enable = true;
      user = "alice";
    };

    services.desktopManager.gnome = {
      enable = true;
      debug = true;
    };

    # Run GNOME Shell in unsafe mode for testing
    systemd.user.services."org.gnome.Shell@wayland" = {
      serviceConfig = {
        ExecStart = [
          ""
          "${pkgs.gnome-shell}/bin/gnome-shell --unsafe-mode"
        ];
      };
    };
  };

  enableOCR = true;

  testScript = {nodes, ...}: let
    user = nodes.gnome.users.users.alice;
    uid = toString user.uid;
    bus = "DBUS_SESSION_BUS_ADDRESS=unix:path=/run/user/${uid}/bus";
    run = command: "su - ${user.name} -c '${bus} ${command}'";

    eval = command:
      run "gdbus call --session -d org.gnome.Shell -o /org/gnome/Shell -m org.gnome.Shell.Eval ${command}";

    startingUp = eval "Main.layoutManager._startingUp";
  in ''
    # Test setup
    for m in [gnome, dashboard]:
      m.systemctl("start network-online.target")
      m.wait_for_unit("network-online.target")

    with subtest("Login to GNOME with GDM"):
        gnome.wait_for_unit("display-manager.service")
        gnome.wait_for_file("/run/user/${uid}/wayland-0")
        gnome.wait_for_unit("default.target", "${user.name}")

    with subtest("Wait for GNOME Shell to finish starting up"):
        gnome.wait_until_succeeds(
            "${startingUp} | grep -q 'false'"
        )

    with subtest("Check Pareto Security services are enabled"):
        for unit in [
            'paretosecurity-trayicon',
            'paretosecurity-user',
            'paretosecurity-user.timer'
        ]:
            status, out = gnome.systemctl("is-enabled " + unit, "${user.name}")
            assert status == 0, f"Unit {unit} is not enabled (status: {status}): {out}"

    with subtest("Check tray icon is visible"):
        # GNOME needs extension for tray icons, checking if service is running
        gnome.succeed("pgrep -f paretosecurity-tray")
        gnome.wait_for_text("Pareto Security", timeout=180)

    with subtest("Launch Pareto Security from Activities"):
        # Open Activities overview
        gnome.send_key("super")  # Open Application Launcher
        gnome.sleep(2)

        # Search for Pareto Security
        gnome.send_chars("pareto")
        gnome.wait_for_text("Pareto Security", timeout=30)
        gnome.screenshot("gnome-search-pareto")

    with subtest("Check desktop entry exists"):
        gnome.succeed("test -f /home/${user.name}/.local/share/applications/paretosecurity.desktop || test -f /usr/share/applications/paretosecurity.desktop")

    with subtest("Check URL handler registration"):
        # Open GNOME Terminal to capture output
        gnome.succeed("${run "gnome-terminal >&2 &"}")
        gnome.wait_for_window("gnome-terminal")
        gnome.wait_for_text(r"(${user.name}|machine)", timeout=30)
        gnome.screenshot("gnome-terminal-open")

        # Execute URL handler test in terminal to see output
        gnome.send_chars("xdg-open paretosecurity://enrollTeam/?token=gnome-integration-test-token\n")
        gnome.wait_for_text("invite_id not found", timeout=20)
        gnome.screenshot("gnome-url-handler-result")
        gnome.send_key("alt-f4")
  '';
}
