{
  name = "Help";
  interactive.sshBackdoor.enable = true;

  nodes = {
    vanilla = {pkgs, ...}: {
      services.paretosecurity.enable = true;
    };
  };

  testScript = ''
    from textwrap import dedent
    expected = dedent("""\
    Pareto Security CLI is a tool for running and reporting audits to paretosecurity.com.

    Usage:
      paretosecurity [command]

    Available Commands:
      check       Run checks on your system
      completion  Generate the autocompletion script for the specified shell
      config      Configure application settings
      help        Help about any command
      helper      A root helper
      info        Print the system information
      link        Link this device to a team
      schema      Output schema for all checks
      status      Print the status of the checks
      trayicon    Display the status of the checks in the system tray
      unlink      Unlink this device from the team

    Flags:
      -h, --help      help for paretosecurity
          --verbose   output verbose logs
      -v, --version   version for paretosecurity

    Use "paretosecurity [command] --help" for more information about a command.
    """)

    # Test 1: assert default output
    out = vanilla.succeed("paretosecurity")
    assert out == expected, f"Expected did not match actual, got \n{out}"

    # Test 2: assert `--help` output
    out = vanilla.succeed("paretosecurity --help")
    assert out == expected, f"Expected did not match actual, got \n{out}"

    # Test 3: assert `-h` output
    out = vanilla.succeed("paretosecurity -h")
    assert out == expected, f"Expected did not match actual, got \n{out}"

    # Test 4: assert `help` output
    out = vanilla.succeed("paretosecurity help")
    assert out == expected, f"Expected did not match actual, got \n{out}"
  '';
}
