{
  # Override paretosecurity to use the local codebase
  pareto = {
    pkgs,
    lib,
    ...
  }: {
    services.paretosecurity = {
      enable = true;
      package = pkgs.callPackage ../../package.nix {inherit lib;};
    };
    environment.systemPackages = with pkgs; [
      iptables
      ip6tables
    ];
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
