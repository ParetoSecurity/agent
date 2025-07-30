let
  common = import ./common.nix;
  inherit (common) pareto testHelpers;
  inherit (testHelpers) formatCheckOutput;

  passwordManagerUuid = "f962c423-fdf5-428a-a57a-827abc9b253e";

  # Check messages specific to password manager
  checkMessages = {
    ok = "Access Security: Password Manager Presence > [OK] Password manager is present";
    fail = "Access Security: Password Manager Presence > [FAIL] No password manager found";
  };
in {
  name = "Password Manager";
  interactive.sshBackdoor.enable = true;

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

  testScript = ''
    with subtest("Check passes with password manager installed"):
      out = withPwdManager.succeed("paretosecurity check --only ${passwordManagerUuid}")
      expected = ${builtins.toJSON (formatCheckOutput [checkMessages.ok])}
      assert out == expected, f"Expected:\n{expected}\n\nActual:\n{out}"

    with subtest("Check fails without password manager"):
      out = noPwdManager.fail("paretosecurity check --only ${passwordManagerUuid}")
      expected = ${builtins.toJSON (formatCheckOutput [checkMessages.fail])}
      assert out == expected, f"Expected:\n{expected}\n\nActual:\n{out}"
  '';
}
