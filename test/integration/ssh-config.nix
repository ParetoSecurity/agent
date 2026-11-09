let
  common = import ./common.nix;
  inherit (common) users;

  forwardingUUID = "9ff5e3d6-8d56-4c86-9767-b792a856569d";
  hostKeysUUID = "79c330c4-5365-4070-b0bd-ffdbd32cca49";
in
{
  name = "SSH Config";
  interactive.sshBackdoor.enable = true;

  nodes = {
    # Stock OpenSSH client defaults - both checks pass
    defaults =
      { ... }:
      {
        imports = [ (users { }) ];
        services.paretosecurity.enable = true;
      };

    # /etc/ssh/ssh_config forwards the agent to every host
    forwarding =
      { ... }:
      {
        imports = [ (users { }) ];
        services.paretosecurity.enable = true;
        programs.ssh.extraConfig = "ForwardAgent yes";
      };

    # /etc/ssh/ssh_config disables host key verification
    nohostkeys =
      { ... }:
      {
        imports = [ (users { }) ];
        services.paretosecurity.enable = true;
        programs.ssh.extraConfig = "StrictHostKeyChecking no";
      };
  };

  testScript = ''
    def check(node, uuid):
        return node.execute(f"su - alice -c 'paretosecurity check --only {uuid}'")[1]

    def expect(out, line):
        expected = "  • Starting checks...\n" + line + "  • Checks completed.\n"
        assert out == expected, f"Expected did not match actual, got \n{out}"

    forwarding_ok = "  • Access Security: SSH agent is not forwarded to all hosts > [OK] SSH agent is not forwarded to all hosts\n"
    forwarding_fail = "  • Access Security: SSH agent is not forwarded to all hosts > [FAIL] ForwardAgent is set to yes for all hosts\n"
    hostkeys_ok = "  • Access Security: SSH verifies host keys > [OK] SSH verifies host keys\n"
    hostkeys_fail = "  • Access Security: SSH verifies host keys > [FAIL] StrictHostKeyChecking is disabled for all hosts\n"

    # Test 1: stock client passes both checks
    expect(check(defaults, "${forwardingUUID}"), forwarding_ok)
    expect(check(defaults, "${hostKeysUUID}"), hostkeys_ok)

    # Test 2: a per-host rule in ~/.ssh/config is fine
    defaults.succeed("su - alice -c 'mkdir -m 700 -p ~/.ssh'")
    defaults.succeed("su - alice -c \"printf 'Host github.com\\n  ForwardAgent yes\\n  StrictHostKeyChecking no\\n' > ~/.ssh/config\"")
    defaults.succeed("su - alice -c 'chmod 600 ~/.ssh/config'")
    expect(check(defaults, "${forwardingUUID}"), forwarding_ok)
    expect(check(defaults, "${hostKeysUUID}"), hostkeys_ok)

    # Test 3: a global rule in ~/.ssh/config fails both checks
    defaults.succeed("su - alice -c \"printf 'ForwardAgent yes\\nStrictHostKeyChecking no\\n' > ~/.ssh/config\"")
    expect(check(defaults, "${forwardingUUID}"), forwarding_fail)
    expect(check(defaults, "${hostKeysUUID}"), hostkeys_fail)

    # Test 4: system wide ForwardAgent yes fails only the forwarding check
    expect(check(forwarding, "${forwardingUUID}"), forwarding_fail)
    expect(check(forwarding, "${hostKeysUUID}"), hostkeys_ok)

    # Test 5: system wide StrictHostKeyChecking no fails only the host key check
    expect(check(nohostkeys, "${forwardingUUID}"), forwarding_ok)
    expect(check(nohostkeys, "${hostKeysUUID}"), hostkeys_fail)
  '';
}
