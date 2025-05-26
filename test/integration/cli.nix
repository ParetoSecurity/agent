let
  common = import ./common.nix;
  inherit (common) users pareto ssh;
in {
  name = "CLI";

  nodes = {
    agent = {
      pkgs,
      lib,
      ...
    }: {
      imports = [
        (users {})
        (pareto {inherit pkgs lib;})
      ];
    };
  };

  interactive.nodes.vanilla = {...}:
    ssh {port = 2221;} {};

  testScript = ''
    # Test setup
    agent.succeed("su - alice -c 'mkdir -p /home/alice/.config'")
    agent.systemctl("start network-online.target")
    agent.wait_for_unit("network-online.target")

    # Test 1: Test the systemd socket is installed & enabled
    agent.succeed('systemctl is-enabled paretosecurity.socket')

    # Test 2: run all checks, with failures
    out = agent.fail("su - alice -c 'paretosecurity check'")
    expected = sorted((
        "  • Starting checks...\n"
        "  • Access Security: Automatic login is disabled > [OK] Automatic login is off\n"
        "  • Access Security: Access to Docker is restricted > [DISABLED] Docker is not installed\n"
        "  • Access Security: Password is required to unlock the screen > [FAIL] Password after sleep or screensaver is off\n"
        "  • Access Security: SSH keys have password protection > [DISABLED] No private keys found in .ssh directory\n"
        "  • Access Security: SSH keys have sufficient algorithm strength > [DISABLED] No private keys found in the .ssh directory\n"
        "  • System Integrity: SecureBoot is enabled > [FAIL] System is not running in UEFI mode\n"
        "  • Application Updates: Apps are up to date > [OK] All packages are up to date\n"
        "  • Firewall & Sharing: Sharing printers is off > [OK] Sharing printers is off\n"
        "  • [root] System Integrity: Filesystem encryption is enabled > [FAIL] Block device encryption is disabled\n"
        "  • Firewall & Sharing: Remote login is disabled > [OK] No remote access services found running\n"
        "  • Firewall & Sharing: File Sharing is disabled > [OK] No file sharing services found running\n"
        "  • Access Security: SSH Server Configuration is Secure > [DISABLED] sshd or ssh service is not running\n"
        "  • Application Updates: Pareto Security is up to date > [ERROR] ErrTransport: Get \"https://api.github.com/repos/ParetoSecurity/agent/releases\": dial tcp: lookup api.github.com: no such host\n"
        "  • Access Security: Password Manager Presence > [FAIL] No password manager found\n"
        "  • [root] Firewall & Sharing: Firewall is configured > [OK] Firewall is on\n"
        "  • Checks completed.\n"
    ).split("\n"))

    assert sorted(out.split("\n")) == expected, f"{expected} did not match actual, got \n{out}"

    # Test 3: disable failing checks, run again
    agent.succeed("su - alice -c 'paretosecurity config disable 37dee029-605b-4aab-96b9-5438e5aa44d8'")  # screenlock
    agent.succeed("su - alice -c 'paretosecurity config disable c96524f2-850b-4bb9-abc7-517051b6c14e'")  # secureboot
    agent.succeed("su - alice -c 'paretosecurity config disable 21830a4e-84f1-48fe-9c5b-beab436b2cdb'")  # luks
    agent.succeed("su - alice -c 'paretosecurity config disable f962c423-fdf5-428a-a57a-827abc9b253e'")  # password manager
    out = agent.succeed("su - alice -c 'paretosecurity check'")
  '';
}
