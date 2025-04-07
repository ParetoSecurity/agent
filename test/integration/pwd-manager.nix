let
  common = import ./common.nix;
  inherit (common) pareto ssh;
in {
  name = "Password Manager";

  nodes = {
    withPwdManager = {
      pkgs,
      lib,
      ...
    }: {
      imports = [
        (pareto {inherit pkgs lib;})
      ];
      environment.systemPackages = with pkgs; [
        bitwarden
      ];
    };

    noPwdManager = {
      pkgs,
      lib,
      ...
    }: {
      imports = [
        (pareto {inherit pkgs lib;})
      ];
      # No password manager installed
    };
  };

  interactive.nodes.withPwdManager = {...}:
    ssh {port = 2221;} {};

  interactive.nodes.noPwdManager = {...}:
    ssh {port = 2222;} {};

  testScript = ''
    # Test 1: Check passes with password managers installed
    out = withPwdManager.succeed("paretosecurity check --only f962c423-fdf5-428a-a57a-827abc9b253e")
    expected = (
        "  • Starting checks...\n"
        "  • Access Security: Password Manager Presence > [OK] Password manager is present\n"
        "  • Checks completed.\n"
    )
    assert out == expected, f"Expected did not match actual, got \n{out}"

    # Test 2: Check fails without password manager
    status, out = noPwdManager.execute("paretosecurity check --only f962c423-fdf5-428a-a57a-827abc9b253e")
    expected = (
        "  • Starting checks...\n"
        "  • Access Security: Password Manager Presence > [FAIL] No password manager found\n"
        "  • Checks completed.\n"
    )
    assert out == expected, f"Expected did not match actual, got \n{out}"
  '';
}
