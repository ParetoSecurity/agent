{
  # Create dummy user account
  users = {}: {
    users.users.alice = {
      isNormalUser = true;
      description = "Alice Foobar";
      password = "foobar";
      uid = 1000;
    };
  };

  # Paretosecurity that uses local codebase
  pareto = {
    config,
    lib,
    pkgs,
    ...
  }: {
    nixpkgs.overlays = [
      (final: prev: {
        paretosecurity = final.callPackage ../../package.nix {};
      })
    ];

    services.paretosecurity = {
      enable = true;
      package = pkgs.paretosecurity;
    };
  };

  # Paretosecurity with local codebase and patched dashboard URL
  paretoPatchedDash = {
    config,
    lib,
    pkgs,
    ...
  }: {
    nixpkgs.overlays = [
      (final: prev: {
        paretosecurity = (final.callPackage ../../package.nix {}).overrideAttrs (oldAttrs: {
          postPatch =
            oldAttrs.postPatch or ""
            + ''
              substituteInPlace team/report.go \
                --replace-warn 'const reportURL = "https://cloud.paretosecurity.com"' \
                               'const reportURL = "http://dashboard"'
            '';
        });
      })
    ];

    services.paretosecurity = {
      enable = true;
      package = pkgs.paretosecurity;
    };
  };

  # Dashboard mockup server
  dashboard = {}: {
    networking.firewall.allowedTCPPorts = [80];

    services.nginx = {
      enable = true;
      virtualHosts."dashboard" = {
        locations."/api/v1/team/".extraConfig = ''
          add_header Content-Type application/json;
          return 200 '{"message": "Linked device."}';
        '';
      };
    };
  };

  # Common configuration for Display Manager
  displayManager = {pkgs}: {
    services.displayManager.autoLogin = {
      enable = true;
      user = "alice";
    };

    virtualisation.resolution = {
      x = 800;
      y = 600;
    };

    environment.systemPackages = [pkgs.xdotool];
    environment.variables.XAUTHORITY = "/home/alice/.Xauthority";
  };

  # Easier tests debugging by SSH-ing into nodes
  ssh = {port}: {...}: {
    services.openssh = {
      enable = true;
      settings = {
        PermitRootLogin = "yes";
        PermitEmptyPasswords = "yes";
      };
    };
    security.pam.services.sshd.allowNullPassword = true;
    virtualisation.forwardPorts = [
      {
        from = "host";
        host.port = port;
        guest.port = 22;
      }
    ];
  };

  # Common test utilities
  testHelpers = let
    # Common check messages
    checkMessages = {
      firewall = {
        ok = "[root] Firewall & Sharing: Firewall is configured > [OK] Firewall is on";
        fail = "[root] Firewall & Sharing: Firewall is configured > [FAIL] Firewall is off";
      };
      screenlock = {
        ok = "Access Security: Password is required to unlock the screen > [OK] Password after sleep or screensaver is on";
        fail = "Access Security: Password is required to unlock the screen > [FAIL] Password after sleep or screensaver is off";
      };
      secureboot = {
        ok = "System Integrity: SecureBoot is enabled > [OK] Secure Boot is enabled";
        fail = "System Integrity: SecureBoot is enabled > [FAIL] System is not running in UEFI mode";
      };
      luks = {
        ok = "[root] System Integrity: Filesystem encryption is enabled > [OK] Disk encryption is enabled";
        fail = "[root] System Integrity: Filesystem encryption is enabled > [FAIL] Block device encryption is disabled";
      };
      autologin = {
        ok = "Access Security: Automatic login is disabled > [OK] Automatic login is off";
        fail = "Access Security: Automatic login is disabled > [FAIL] Automatic login is on";
      };
    };

    # Format expected output for easier assertions
    formatCheckOutput = lines: let
      header = "  • Starting checks...";
      footer = "  • Checks completed.";
      formattedLines = map (line: "  • ${line}") lines;
    in
      builtins.concatStringsSep "\n" ([header] ++ formattedLines ++ [footer]) + "\n";
  in {
    inherit checkMessages formatCheckOutput;

    # Setup network for test nodes
    setupNetwork = nodes: ''
      # Setup network for all nodes
      ${builtins.concatStringsSep "\n" (map (node: ''
          ${node}.systemctl("start network-online.target")
          ${node}.wait_for_unit("network-online.target")
        '')
        nodes)}
    '';

    # Wait for service with optional port check
    waitForService = {
      node,
      service,
      port ? null,
    }: ''
      ${node}.wait_for_unit("${service}")
      ${
        if port != null
        then "${node}.wait_for_open_port(${toString port})"
        else ""
      }
    '';

    # Run Pareto check with specific UUID
    runCheck = {
      node,
      uuid,
      user ? null,
    }: let
      cmd = "paretosecurity check --only ${uuid}";
      fullCmd =
        if user != null
        then "su - ${user} -c '${cmd}'"
        else cmd;
    in
      fullCmd;

    # Assert check output matches expected
    assertCheckOutput = {
      actual,
      expected,
    }: ''
      expected = ${builtins.toJSON expected}
      assert ${actual} == expected, f"Expected:\n{expected}\n\nActual:\n{${actual}}"
    '';

    # Retry operation with timeout
    retryOperation = {
      operation,
      timeout ? 30,
      interval ? 1,
    }: ''
      retry(lambda: ${operation}, timeout=${toString timeout}, interval=${toString interval})
    '';

    # Common timeout values (in seconds)
    timeouts = {
      service = 30;
      network = 60;
      desktop = 120;
      boot = 180;
    };

    # Common port numbers
    ports = {
      http = 80;
      https = 443;
      ssh = 22;
      sshBackdoor = 2222;
    };
  };
}
