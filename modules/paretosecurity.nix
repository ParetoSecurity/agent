{
  config,
  lib,
  pkgs,
  ...
}: {
  options.paretosecurity.paretosecurityBin = mkOption {
    type = types.str;
    default = "${pkgs.paretosecurity}/bin/paretosecurity";
    defaultText = literalExpression ''
      "''${pkgs.paretosecurity}/bin/paretosecurity"
    '';
    description = ''
      The paretosecurity executable to use.
    '';
  };
  options.paretosecurity.enable = lib.mkOption {
    type = lib.types.bool;
    default = false;
    description = "Enable ParetoSecurity.";
  };
  config = lib.mkIf config.paretosecurity.enable {
    environment.systemPackages = with pkgs; [config.paretosecurity.paretosecurityBin];

    systemd.sockets."pareto-core" = {
      wantedBy = ["sockets.target"];
      socketConfig = {
        ListenStream = "/var/run/pareto-core.sock";
        SocketMode = "0666";
      };
    };

    systemd.services."pareto-core" = {
      requires = ["pareto-core.socket"];
      after = ["pareto-core.socket"];
      wantedBy = ["multi-user.target"];
      serviceConfig = {
        ExecStart = ["${config.paretosecurity.paretosecurityBin}" "helper" "--verbose" "--socket" "/var/run/pareto-core.sock"];
        User = "root";
        Group = "root";
        StandardInput = "socket";
        Type = "oneshot";
        RemainAfterExit = "no";
        StartLimitInterval = "1s";
        StartLimitBurst = 100;
        ProtectSystem = "full";
        ProtectHome = true;
        StandardOutput = "journal";
        StandardError = "journal";
      };
    };

    systemd.user.services."pareto-core-hourly" = {
      wantedBy = ["timers.target"];
      serviceConfig = {
        Type = "oneshot";
        ExecStart = ["${config.paretosecurity.paretosecurityBin}" "check"];
        StandardInput = "null";
      };
    };

    systemd.user.timers."pareto-core-hourly" = {
      wantedBy = ["timers.target"];
      timerConfig = {
        OnCalendar = "hourly";
        Persistent = true;
      };
    };
  };
}
