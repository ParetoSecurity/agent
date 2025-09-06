let
  # A simple web server for testing connectivity
  nginx =
    { pkgs, ... }:
    {
      services.nginx = {
        enable = true;
        virtualHosts."localhost" = {
          locations."/" = {
            root = pkgs.writeTextDir "index.html" "<html><body><h1>Test Server</h1></body></html>";
          };
        };
      };
    };
in
{
  name = "Firewall";
  interactive.sshBackdoor.enable = true;

  nodes = {
    wideopen =
      { pkgs, ... }:
      {
        imports = [
          (nginx { inherit pkgs; })
        ];
        services.paretosecurity.enable = true;
        networking.firewall.enable = false;
      };

    iptables =
      { pkgs, ... }:
      {
        imports = [
          (nginx { inherit pkgs; })
        ];
        services.paretosecurity.enable = true;
        networking.firewall.enable = true;
      };

    nftables =
      { pkgs, ... }:
      {
        imports = [
          (nginx { inherit pkgs; })
        ];
        services.paretosecurity.enable = true;
        networking.nftables.enable = true;
      };
  };

  testScript = ''
    # Test Setup
    for m in [wideopen, iptables, nftables]:
      m.systemctl("start network-online.target")
      m.wait_for_unit("network-online.target")
      m.wait_for_unit("nginx")
      m.wait_for_open_port(80)

    # Test 0: assert firewall is actually configured
    wideopen.fail("curl --fail --connect-timeout 2 http://iptables")
    wideopen.fail("curl --fail --connect-timeout 2 http://nftables")
    iptables.succeed("curl --fail --connect-timeout 2 http://wideopen")

    # Test 1: check fails with no firewall enabled
    out = wideopen.fail("paretosecurity check --only 2e46c89a-5461-4865-a92e-3b799c12034a")
    expected = (
        "  • Starting checks...\n"
        "  • [root] Firewall & Sharing: Firewall is configured > [FAIL] Firewall is off\n"
        "  • Checks completed.\n"
    )
    assert out == expected, f"{expected} did not match actual, got \n{out}"

    # Test 2: check succeeds with iptables enabled
    out = iptables.succeed("paretosecurity check --only 2e46c89a-5461-4865-a92e-3b799c12034a")
    expected = (
        "  • Starting checks...\n"
        "  • [root] Firewall & Sharing: Firewall is configured > [OK] Firewall is on\n"
        "  • Checks completed.\n"
    )
    assert out == expected, f"{expected} did not match actual, got \n{out}"

    # Test 3: check succeeds with nftables enabled
    out = nftables.succeed("paretosecurity check --only 2e46c89a-5461-4865-a92e-3b799c12034a")
    expected = (
        "  • Starting checks...\n"
        "  • [root] Firewall & Sharing: Firewall is configured > [OK] Firewall is on\n"
        "  • Checks completed.\n"
    )
    assert out == expected, f"{expected} did not match actual, got \n{out}"
  '';
}
