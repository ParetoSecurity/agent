let
  common = import ./common.nix;
  inherit (common) pareto;
in {
  name = "Help";
  interactive.sshBackdoor.enable = true;

  nodes = {
    vanilla = {
      pkgs,
      lib,
      ...
    }: {
      imports = [(pareto {inherit pkgs lib;})];
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

    # Test help output variations
    help_commands = [
        ("default output", "paretosecurity"),
        ("--help flag", "paretosecurity --help"),
        ("-h flag", "paretosecurity -h"),
        ("help command", "paretosecurity help")
    ]

    for test_name, command in help_commands:
        out = vanilla.succeed(command)
        assert out == expected, f"Test '{test_name}' failed. Expected:\n{expected}\n\nActual:\n{out}"
  '';
}
