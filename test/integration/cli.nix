let
  common = import ./common.nix;
  inherit (common) users pareto testHelpers;
  inherit (testHelpers) setupNetwork;

  # UUIDs for tests
  checks = {
    screenlock = "37dee029-605b-4aab-96b9-5438e5aa44d8";
    secureboot = "c96524f2-850b-4bb9-abc7-517051b6c14e";
    luks = "21830a4e-84f1-48fe-9c5b-beab436b2cdb";
    passwordManager = "f962c423-fdf5-428a-a57a-827abc9b253e";
  };
in {
  name = "CLI";
  interactive.sshBackdoor.enable = true;

  nodes = {
    agent = {
      config,
      lib,
      pkgs,
      ...
    }: {
      imports = [
        (users {})
        (pareto {inherit pkgs lib;})
      ];
    };
  };

  testScript = ''
    # Test setup
    agent.succeed("su - alice -c 'mkdir -p /home/alice/.config'")
    ${setupNetwork ["agent"]}

    with subtest("Verify systemd socket is installed & enabled"):
      agent.succeed('systemctl is-enabled paretosecurity.socket')

    with subtest("Run all checks with expected failures"):
      out = agent.fail("su - alice -c 'paretosecurity check'")

      # Define expected check lines
      expectedLines = [
      "Access Security: Automatic login is disabled > [OK] Automatic login is off"
      "Access Security: Access to Docker is restricted > [DISABLED] Docker is not installed"
      "Access Security: Password is required to unlock the screen > [FAIL] Password after sleep or screensaver is off"
      "Access Security: SSH keys have password protection > [DISABLED] No private keys found in .ssh directory"
      "Access Security: SSH keys have sufficient algorithm strength > [DISABLED] No private keys found in the .ssh directory"
      "System Integrity: SecureBoot is enabled > [FAIL] System is not running in UEFI mode"
      "Application Updates: Apps are up to date > [OK] All packages are up to date"
      "Firewall & Sharing: Sharing printers is off > [OK] Sharing printers is off"
      "[root] System Integrity: Filesystem encryption is enabled > [FAIL] Block device encryption is disabled"
      "Firewall & Sharing: Remote login is disabled > [OK] No remote access services found running"
      "Firewall & Sharing: File Sharing is disabled > [OK] No file sharing services found running"
      "Application Updates: Pareto Security is up to date > [ERROR] ErrTransport: Get \"https://api.github.com/repos/ParetoSecurity/agent/releases\": dial tcp: lookup api.github.com: no such host"
      "Access Security: Password Manager Presence > [FAIL] No password manager found"
      "[root] Firewall & Sharing: Firewall is configured > [OK] Firewall is on"
    ]

    # Build expected output
    header = "  • Starting checks..."
    footer = "  • Checks completed."
    formattedLines = [f"  • {line}" for line in expectedLines]
    expected = sorted([header] + formattedLines + [footer] + [""])

      assert sorted(out.split("\n")) == expected, f"Expected did not match actual, got \n{out}"

    with subtest("Disable failing checks and verify success"):
      agent.succeed("su - alice -c 'paretosecurity config disable ${checks.screenlock}'")
      agent.succeed("su - alice -c 'paretosecurity config disable ${checks.secureboot}'")
      agent.succeed("su - alice -c 'paretosecurity config disable ${checks.luks}'")
      agent.succeed("su - alice -c 'paretosecurity config disable ${checks.passwordManager}'")

      # Should succeed now with disabled checks
      out = agent.succeed("su - alice -c 'paretosecurity check'")
  '';
}
