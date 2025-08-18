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

  # Paretosecurity that uses local codebase from overlay
  pareto = {
    pkgs,
    lib,
    ...
  }: {
    services.paretosecurity = {
      enable = true;
      package = pkgs.paretosecurity;
    };
  };

  # Paretosecurity with local codebase and patched dashboard URL
  paretoPatchedDash = {
    pkgs,
    lib,
    ...
  }: {
    services.paretosecurity = {
      enable = true;
      package = pkgs.paretosecurity.overrideAttrs (oldAttrs: {
        postPatch =
          oldAttrs.postPatch or ""
          + ''
            substituteInPlace team/report.go \
              --replace-warn 'const reportURL = "https://cloud.paretosecurity.com"' \
                             'const reportURL = "http://dashboard"'
          '';
      });
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
}
