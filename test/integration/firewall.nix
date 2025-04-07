let
  common = import ./common.nix;
  inherit (common) pareto ssh;

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

  nodes = {
    wideopen = {
      pkgs,
      lib,
      ...
    }: {
      imports = [
        (pareto {inherit pkgs lib;})
        (nginx {inherit pkgs;})
      ];
      networking.firewall.enable = false;
    };

    walled = {
      pkgs,
      lib,
      ...
    }: {
      imports = [
        (pareto {inherit pkgs lib;})
        (nginx {inherit pkgs;})
      ];
      networking.firewall.enable = true;
    };
  };

  interactive.nodes.wideopen = {...}:
    ssh {port = 2221;} {};

  interactive.nodes.walled = {...}:
    ssh {port = 2222;} {};

  testScript = ''
    # Test Setup
    for m in [wideopen, walled]:
      m.systemctl("start network-online.target")
      m.wait_for_unit("network-online.target")
      m.wait_for_unit("nginx")

    # Test 0: assert firewall is actually configured
    wideopen.fail("curl --fail --connect-timeout 2 http://walled")
    walled.succeed("curl --fail --connect-timeout 2 http://wideopen")

    # Test 1: check fails with iptables disabled
    out = wideopen.fail("paretosecurity check --only 2e46c89a-5461-4865-a92e-3b799c12034a")
    expected = (
        "  • Starting checks...\n"
        "  • [root] Firewall & Sharing: Firewall is configured > [FAIL] Firewall is off\n"
        "  • Checks completed.\n"
    )
    assert out == expected, f"{expected} did not match actual, got \n{out}"

    # Test 2: check succeeds with iptables enabled
    out = walled.succeed("paretosecurity check --only 2e46c89a-5461-4865-a92e-3b799c12034a")
    expected = (
        "  • Starting checks...\n"
        "  • [root] Firewall & Sharing: Firewall is configured > [OK] Firewall is on\n"
        "  • Checks completed.\n"
    )
    assert out == expected, f"{expected} did not match actual, got \n{out}"
  '';
}
