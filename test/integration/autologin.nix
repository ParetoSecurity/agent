let
  common = import ./common.nix;

  # Create a user for testing
  testUser = {}: {
    users.users.testuser = {
      isNormalUser = true;
      description = "Test User";
      password = "test123";
      uid = 1001;
    };
  };
in {
  name = "Autologin";
  interactive.sshBackdoor.enable = true;

  nodes = {
    # No autologin configured - should pass
    noautologin = {pkgs, ...}: {
      imports = [
        (testUser {})
      ];
      services.paretosecurity.enable = true;
    };

    # Getty autologin configured - should fail
    gettyautologin = {pkgs, ...}: {
      imports = [
        (testUser {})
      ];
      services.paretosecurity.enable = true;
      services.getty.autologinUser = "root";
    };

    # Getty autologin with regular user - should fail
    gettyuser = {pkgs, ...}: {
      imports = [
        (testUser {})
      ];
      services.paretosecurity.enable = true;
      services.getty.autologinUser = "testuser";
    };

    # Display manager automatic login - should fail
    gdmautologin = {pkgs, ...}: {
      imports = [
        (testUser {})
      ];
      services.paretosecurity.enable = true;
      services.xserver.enable = true;
      services.displayManager.gdm.enable = true;
      services.displayManager.autoLogin = {
        enable = true;
        user = "testuser";
      };
    };

    # Display manager timed login with delay - should fail
    gdmtimedlogin = {pkgs, ...}: {
      imports = [
        (testUser {})
      ];
      services.paretosecurity.enable = true;
      services.xserver.enable = true;
      services.displayManager.gdm.enable = true;
      services.displayManager.gdm.autoLogin.delay = 30;
      services.displayManager.autoLogin = {
        enable = true;
        user = "testuser";
      };
    };

    # Display manager with zero delay (immediate login) - should fail
    gdmzerodelay = {pkgs, ...}: {
      imports = [
        (testUser {})
      ];
      services.paretosecurity.enable = true;
      services.xserver.enable = true;
      services.displayManager.gdm.enable = true;
      services.displayManager.gdm.autoLogin.delay = 0;
      services.displayManager.autoLogin = {
        enable = true;
        user = "testuser";
      };
    };

    # LightDM autologin - should fail
    lightdmautologin = {pkgs, ...}: {
      imports = [
        (testUser {})
      ];
      services.paretosecurity.enable = true;
      services.xserver.enable = true;
      services.xserver.displayManager.lightdm.enable = true;
      services.displayManager.autoLogin = {
        enable = true;
        user = "testuser";
      };
    };

    # SDDM autologin - should fail
    sddmautologin = {pkgs, ...}: {
      imports = [
        (testUser {})
      ];
      services.paretosecurity.enable = true;
      services.xserver.enable = true;
      services.displayManager.sddm.enable = true;
      services.displayManager.autoLogin = {
        enable = true;
        user = "testuser";
      };
    };
  };

  testScript = ''
    # Start all nodes
    for m in [noautologin, gettyautologin, gettyuser, gdmautologin, gdmtimedlogin, gdmzerodelay, lightdmautologin, sddmautologin]:
      m.systemctl("start network-online.target")
      m.wait_for_unit("network-online.target")
      m.wait_for_unit("multi-user.target")

    # Test 1: No autologin configured - should pass
    out = noautologin.succeed("paretosecurity check --only f962c423-fdf5-428a-a57a-816abc9b253e")
    expected = (
        "  • Starting checks...\n"
        "  • Security Policy: Automatic login is disabled > [OK] Automatic login is off\n"
        "  • Checks completed.\n"
    )
    assert out == expected, f"Test 1 failed: {expected} did not match actual, got \n{out}"

    # Test 2: Getty autologin as root - should fail
    out = gettyautologin.fail("paretosecurity check --only f962c423-fdf5-428a-a57a-816abc9b253e")
    expected = (
        "  • Starting checks...\n"
        "  • Security Policy: Automatic login is disabled > [FAIL] Automatic login is on\n"
        "  • Checks completed.\n"
    )
    assert out == expected, f"Test 2 failed: {expected} did not match actual, got \n{out}"

    # Test 3: Getty autologin as regular user - should fail
    out = gettyuser.fail("paretosecurity check --only f962c423-fdf5-428a-a57a-816abc9b253e")
    expected = (
        "  • Starting checks...\n"
        "  • Security Policy: Automatic login is disabled > [FAIL] Automatic login is on\n"
        "  • Checks completed.\n"
    )
    assert out == expected, f"Test 3 failed: {expected} did not match actual, got \n{out}"

    # Test 4: GDM automatic login - should fail
    # Check if GDM config file was created
    gdmautologin.succeed("test -f /etc/gdm/custom.conf || test -f /etc/gdm3/custom.conf")
    out = gdmautologin.fail("paretosecurity check --only f962c423-fdf5-428a-a57a-816abc9b253e")
    expected = (
        "  • Starting checks...\n"
        "  • Security Policy: Automatic login is disabled > [FAIL] Automatic login is on\n"
        "  • Checks completed.\n"
    )
    assert out == expected, f"Test 4 failed: {expected} did not match actual, got \n{out}"

    # Test 5: GDM timed login with delay - should fail
    gdmtimedlogin.succeed("test -f /etc/gdm/custom.conf || test -f /etc/gdm3/custom.conf")
    out = gdmtimedlogin.fail("paretosecurity check --only f962c423-fdf5-428a-a57a-816abc9b253e")
    expected = (
        "  • Starting checks...\n"
        "  • Security Policy: Automatic login is disabled > [FAIL] Automatic login is on\n"
        "  • Checks completed.\n"
    )
    assert out == expected, f"Test 5 failed: {expected} did not match actual, got \n{out}"

    # Test 6: GDM with zero delay (immediate login) - should fail
    gdmzerodelay.succeed("test -f /etc/gdm/custom.conf || test -f /etc/gdm3/custom.conf")
    out = gdmzerodelay.fail("paretosecurity check --only f962c423-fdf5-428a-a57a-816abc9b253e")
    expected = (
        "  • Starting checks...\n"
        "  • Security Policy: Automatic login is disabled > [FAIL] Automatic login is on\n"
        "  • Checks completed.\n"
    )
    assert out == expected, f"Test 6 failed: {expected} did not match actual, got \n{out}"

    # Test 7: LightDM autologin - should fail
    lightdmautologin.succeed("test -f /etc/lightdm/lightdm.conf")
    out = lightdmautologin.fail("paretosecurity check --only f962c423-fdf5-428a-a57a-816abc9b253e")
    expected = (
        "  • Starting checks...\n"
        "  • Security Policy: Automatic login is disabled > [FAIL] Automatic login is on\n"
        "  • Checks completed.\n"
    )
    assert out == expected, f"Test 7 failed: {expected} did not match actual, got \n{out}"

    # Test 8: SDDM autologin - should fail
    sddmautologin.succeed("test -f /etc/sddm.conf || test -d /etc/sddm.conf.d")
    out = sddmautologin.fail("paretosecurity check --only f962c423-fdf5-428a-a57a-816abc9b253e")
    expected = (
        "  • Starting checks...\n"
        "  • Security Policy: Automatic login is disabled > [FAIL] Automatic login is on\n"
        "  • Checks completed.\n"
    )
    assert out == expected, f"Test 8 failed: {expected} did not match actual, got \n{out}"

    # Additional debug checks for getty configuration
    gettyautologin.succeed("systemctl cat getty@tty1.service | grep -q autologin || echo 'Autologin flag found in getty service'")

    # Check for the autologin marker file (if autologinOnce is configured)
    # This would be created after first login, but we're checking the configuration exists
    gettyuser.succeed("systemctl cat getty@tty1.service | grep -q autologin || echo 'Autologin configured for regular user'")
  '';
}
