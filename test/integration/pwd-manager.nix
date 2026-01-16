{
  name = "PWD";
  interactive.sshBackdoor.enable = true;

  nodes = {
    bitwarden =
      { pkgs, ... }:
      {
        services.paretosecurity.enable = true;
        environment.systemPackages = with pkgs; [
          bitwarden
        ];
      };

    rbw =
      { pkgs, ... }:
      {
        services.paretosecurity.enable = true;
        environment.systemPackages = with pkgs; [
          rbw
        ];
      };

    noPwdManager =
      { ... }:
      {
        services.paretosecurity.enable = true;
        # No password manager installed
      };
  };

  testScript = ''
    # Test 1: Check passes with bitwarden installed
    out = bitwarden.succeed("paretosecurity check --only f962c423-fdf5-428a-a57a-827abc9b253e")
    expected = (
        "  • Starting checks...\n"
        "  • Access Security: Password Manager Presence > [OK] Password manager is present\n"
        "  • Checks completed.\n"
    )
    assert out == expected, f"Expected did not match actual, got \n{out}"

    # Test 2: Check passes with rbw installed
    out = rbw.succeed("paretosecurity check --only f962c423-fdf5-428a-a57a-827abc9b253e")
    expected = (
        "  • Starting checks...\n"
        "  • Access Security: Password Manager Presence > [OK] Password manager is present\n"
        "  • Checks completed.\n"
    )
    assert out == expected, f"Expected did not match actual, got \n{out}"

    # Test 3: Check fails without password manager
    status, out = noPwdManager.execute("paretosecurity check --only f962c423-fdf5-428a-a57a-827abc9b253e")
    expected = (
        "  • Starting checks...\n"
        "  • Access Security: Password Manager Presence > [FAIL] No password manager found\n"
        "  • Checks completed.\n"
    )
    assert out == expected, f"Expected did not match actual, got \n{out}"
  '';
}
