let
  common = import ./common.nix;
  inherit (common) users;

  sudoUUID = "f358454b-5d9b-4c6b-b47f-13b55b17ad0e";

  wheelUser = _: {
    users.users.alice.extraGroups = [ "wheel" ];
  };
in
{
  name = "Sudo";
  interactive.sshBackdoor.enable = true;

  nodes = {
    # alice is in wheel and sudo asks for her password
    password =
      { ... }:
      {
        imports = [
          (users { })
          wheelUser
        ];
        services.paretosecurity.enable = true;
      };

    # alice is in wheel and wheel runs everything with NOPASSWD
    nopassword =
      { ... }:
      {
        imports = [
          (users { })
          wheelUser
        ];
        services.paretosecurity.enable = true;
        security.sudo.wheelNeedsPassword = false;
      };

    # alice is in wheel and a Defaults line turns authentication off
    noauth =
      { ... }:
      {
        imports = [
          (users { })
          wheelUser
        ];
        services.paretosecurity.enable = true;
        security.sudo.extraConfig = "Defaults !authenticate";
      };

    # alice may run one command without a password, nothing else
    command =
      { ... }:
      {
        imports = [ (users { }) ];
        services.paretosecurity.enable = true;
        security.sudo.extraRules = [
          {
            users = [ "alice" ];
            commands = [
              {
                command = "/run/current-system/sw/bin/systemctl";
                options = [ "NOPASSWD" ];
              }
            ];
          }
        ];
      };
  };

  testScript = ''
    def check(node):
        return node.execute("su - alice -c 'paretosecurity check --only ${sudoUUID}'")[1]

    def expect(out, line):
        expected = "  • Starting checks...\n" + line + "  • Checks completed.\n"
        assert out == expected, f"Expected did not match actual, got \n{out}"

    ok = "  • [root] Access Security: sudo requires a password > [OK] sudo requires a password\n"
    fail = "  • [root] Access Security: sudo requires a password > [FAIL] alice can run any command with sudo without a password\n"

    # Test 1: wheel with a password passes
    expect(check(password), ok)

    # Test 2: wheel with NOPASSWD fails
    expect(check(nopassword), fail)

    # Test 3: Defaults !authenticate fails
    expect(check(noauth), fail)

    # Test 4: NOPASSWD on a single command passes
    expect(check(command), ok)
  '';
}
