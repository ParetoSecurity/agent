# Test for KDE Plasma 5 using nixpkgs-24.11
# This is separate because Plasma 5 was removed from newer nixpkgs
{ pkgsOverlayed }:
{
  name = "Screen Lock - Plasma 5";
  interactive.sshBackdoor.enable = true;

  nodes = {
    kde5 =
      { ... }:
      {
        services.paretosecurity.enable = true;
        services.paretosecurity.package = pkgsOverlayed.paretosecurity;

        # Install KDE Plasma 5 Desktop Environment
        services.xserver.enable = true;
        services.xserver.desktopManager.plasma5.enable = true;
        services.xserver.displayManager.sddm.enable = true;
        services.colord.enable = false;
      };
  };

  testScript = ''
    # Test KDE Plasma 5
    # Test 1: Check passes with lock enabled
    kde5.succeed("kwriteconfig5 --file kscreenlockerrc --group Daemon --key LockOnResume true")
    out = kde5.succeed("paretosecurity check --only 37dee029-605b-4aab-96b9-5438e5aa44d8")
    assert "[OK] Password after sleep or screensaver is on" in out, f"Expected check to pass, got \n{out}"

    # Test 2: Check fails when lock is disabled
    kde5.succeed("kwriteconfig5 --file kscreenlockerrc --group Daemon --key LockOnResume false")
    status, out = kde5.execute("paretosecurity check --only 37dee029-605b-4aab-96b9-5438e5aa44d8")
    assert "[FAIL] Password after sleep or screensaver is off" in out, f"Expected check to fail, got \n{out}"
  '';
}
