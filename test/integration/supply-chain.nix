{
  name = "Supply Chain";
  interactive.sshBackdoor.enable = true;

  nodes = {
    protected =
      { pkgs, ... }:
      {
        services.paretosecurity.enable = true;
        environment.systemPackages = [ pkgs.nodejs ];
      };

    unprotected =
      { pkgs, ... }:
      {
        services.paretosecurity.enable = true;
        environment.systemPackages = [ pkgs.nodejs ];
      };
  };

  testScript = ''
    protected.succeed("printf 'min-release-age=7\\nminimum-release-age=10080\\nsave-exact=true\\n' > /root/.npmrc")

    out = protected.succeed("paretosecurity check --only 61bf7ef7-a3ee-4d66-a859-49c3ebeb1e7f")
    expected = (
        "  • Starting checks...\n"
        "  • System Integrity: Package managers delay new releases > [OK] ~/.npmrc delays npm-compatible package releases and pins exact versions\n"
        "  • Checks completed.\n"
    )
    assert out == expected, f"Expected did not match actual, got \n{out}"

    status, out = unprotected.execute("paretosecurity check --only 61bf7ef7-a3ee-4d66-a859-49c3ebeb1e7f")
    expected = (
        "  • Starting checks...\n"
        "  • System Integrity: Package managers delay new releases > [FAIL] /root/.npmrc is missing\n"
        "  • Checks completed.\n"
    )
    assert out == expected, f"Expected did not match actual, got \n{out}"
  '';
}
