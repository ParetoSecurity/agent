{
  # Create dummy user account
  users = _: {
    users.users.alice = {
      isNormalUser = true;
      description = "Alice Foobar";
      password = "foobar";
      uid = 1000;
    };
  };

  # Common configuration for Display Manager
  displayManager =
    { pkgs }:
    {
      services.displayManager.autoLogin = {
        enable = true;
        user = "alice";
      };

      virtualisation.resolution = {
        x = 800;
        y = 600;
      };

      environment.systemPackages = [ pkgs.xdotool ];
      environment.variables.XAUTHORITY = "/home/alice/.Xauthority";
    };
}
