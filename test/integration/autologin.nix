let
  common = import ./common.nix;
  inherit (common) pareto ssh;
in {
  name = "Automatic Login";

  nodes = {
    gnome = {
      pkgs,
      lib,
      ...
    }: {
      imports = [
        (pareto {inherit pkgs lib;})
      ];
      # Install GNOME Desktop Environment
      services.xserver.desktopManager.gnome.enable = true;
      services.xserver.displayManager.gdm.enable = true;
    };

    kde = {
      pkgs,
      lib,
      ...
    }: {
      imports = [
        (pareto {inherit pkgs lib;})
      ];
      # Install KDE Plasma 5 Desktop Environment
      services.xserver.enable = true;
      services.xserver.desktopManager.plasma5.enable = true;
      services.xserver.displayManager.sddm.enable = true;
      services.colord.enable = false;
    };
  };

  interactive.nodes.gnome = {...}:
    ssh {port = 2221;} {};

  interactive.nodes.kde = {...}:
    ssh {port = 2222;} {};

  testScript = ''
    # Test GNOME
    # Test 1: Check passes by default
    out = gnome.succeed("paretosecurity check --only f962c423-fdf5-428a-a57a-816abc9b253e")
    expected = (
        "  • Starting checks...\n"
        "  • Access Security: Password is required to unlock the screen > [OK] Password after sleep or screensaver is on\n"
        "  • Checks completed.\n"
    )
    assert out == expected, f"Expected did not match actual, got \n{out}"
  '';
}
