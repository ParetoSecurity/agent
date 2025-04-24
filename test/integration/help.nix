let
  common = import ./common.nix;
  inherit (common) pareto ssh;
in {
  name = "Help";

  nodes = {
    vanilla = {
      pkgs,
      lib,
      ...
    }: {
      imports = [(pareto {inherit pkgs lib;})];
      environment.systemPackages = with pkgs; [pkg-config];
    };
  };

  interactive.nodes.vanilla = {...}:
    ssh {port = 2221;} {};

  testScript = ''
    expected = (
        "Pareto Security CLI is a tool for running and reporting audits \n"
        "to paretosecurity.com.\n"
        "\n"
        "Usage:\n"
        "  paretosecurity [command]\n"
        "\n"
        "Available Commands:\n"
        "  check       Run checks on your system\n"
        "  completion  Generate the autocompletion script for the specified shell\n"
        "  config      Configure application settings\n"
        "  help        Help about any command\n"
        "  helper      A root helper\n"
        "  info        Print the system information\n"
        "  link        Link team with this device\n"
        "  schema      Output schema for all checks\n"
        "  status      Print the status of the checks\n"
        "  trayicon    Display the status of the checks in the system tray\n"
        "  unlink      Unlink this device from the team\n"
        "\n"
        "Flags:\n"
        "  -h, --help      help for paretosecurity\n"
        "      --verbose   output verbose logs\n"
        "  -v, --version   version for paretosecurity\n"
        "\n"
        'Use "paretosecurity [command] --help" for more information about a command.\n'
    )

    # Test 1: assert default output
    assert vanilla.succeed("paretosecurity") == expected

    # Test 2: assert `--help` output
    assert vanilla.succeed("paretosecurity -h") == expected

    # Test 3: assert `-h` output
    assert vanilla.succeed("paretosecurity --help") == expected

    # Test 4: assert `help` output
    assert vanilla.succeed("paretosecurity help") == expected
  '';
}
