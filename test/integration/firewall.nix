let
  common = import ./common.nix;
  inherit (common) pareto testHelpers;
  inherit (testHelpers) setupNetwork formatCheckOutput checkMessages waitForService ports;

  firewallUuid = "2e46c89a-5461-4865-a92e-3b799c12034a";

  # A simple web server for testing connectivity
  nginx = {pkgs, ...}: {
    services.nginx = {
      enable = true;
      virtualHosts."localhost" = {
        locations."/" = {
          root = pkgs.writeTextDir "index.html" "<html><body><h1>Test Server</h1></body></html>";
        };
      };
    };
  };
in {
  name = "Firewall";
  interactive.sshBackdoor.enable = true;

  nodes = {
    wideopen = {
      config,
      lib,
      pkgs,
      ...
    }: {
      imports = [
        (pareto {inherit config lib pkgs;})
        (nginx {inherit pkgs;})
      ];
      networking.firewall.enable = false;
    };

    iptables = {
      config,
      lib,
      pkgs,
      ...
    }: {
      imports = [
        (pareto {inherit config lib pkgs;})
        (nginx {inherit pkgs;})
      ];
      networking.firewall.enable = true;
    };

    nftables = {
      config,
      lib,
      pkgs,
      ...
    }: {
      imports = [
        (pareto {inherit config lib pkgs;})
        (nginx {inherit pkgs;})
      ];
      networking.nftables.enable = true;
    };
  };

  testScript = ''
    # Test Setup
    ${setupNetwork ["wideopen" "iptables" "nftables"]}

    # Wait for nginx service and port on all nodes
    ${waitForService {
      node = "wideopen";
      service = "nginx";
      port = ports.http;
    }}
    ${waitForService {
      node = "iptables";
      service = "nginx";
      port = ports.http;
    }}
    ${waitForService {
      node = "nftables";
      service = "nginx";
      port = ports.http;
    }}

    # Test 0: Verify firewall is actually configured
    with subtest("Verify firewall configurations"):
      wideopen.fail("curl --fail --connect-timeout 2 http://iptables")
      wideopen.fail("curl --fail --connect-timeout 2 http://nftables")
      iptables.succeed("curl --fail --connect-timeout 2 http://wideopen")

    with subtest("Check fails with no firewall enabled"):
      out = wideopen.fail("paretosecurity check --only ${firewallUuid}")
      expected = ${builtins.toJSON (formatCheckOutput [checkMessages.firewall.fail])}
      assert out == expected, f"Expected:\n{expected}\n\nActual:\n{out}"

    with subtest("Check succeeds with iptables enabled"):
      out = iptables.succeed("paretosecurity check --only ${firewallUuid}")
      expected = ${builtins.toJSON (formatCheckOutput [checkMessages.firewall.ok])}
      assert out == expected, f"Expected:\n{expected}\n\nActual:\n{out}"

    with subtest("Check succeeds with nftables enabled"):
      out = nftables.succeed("paretosecurity check --only ${firewallUuid}")
      expected = ${builtins.toJSON (formatCheckOutput [checkMessages.firewall.ok])}
      assert out == expected, f"Expected:\n{expected}\n\nActual:\n{out}"
  '';
}
