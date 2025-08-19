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
    # Test 1: No autologin configured - should pass
    with subtest("No autologin configured"):
        noautologin.start()
        noautologin.systemctl("start network-online.target")
        noautologin.wait_for_unit("network-online.target")
        noautologin.wait_for_unit("multi-user.target")

        out = noautologin.succeed("paretosecurity check --only f962c423-fdf5-428a-a57a-816abc9b253e")
        expected = (
            "  • Starting checks...\n"
            "  • Security Policy: Automatic login is disabled > [OK] Automatic login is off\n"
            "  • Checks completed.\n"
        )
        assert out == expected, f"Test 1 failed: {expected} did not match actual, got \n{out}"
        noautologin.shutdown()

    # Test 2: Getty autologin as root - should fail
    with subtest("Getty autologin as root"):
        gettyautologin.start()
        gettyautologin.systemctl("start network-online.target")
        gettyautologin.wait_for_unit("network-online.target")
        gettyautologin.wait_for_unit("multi-user.target")

        out = gettyautologin.fail("paretosecurity check --only f962c423-fdf5-428a-a57a-816abc9b253e")
        expected = (
            "  • Starting checks...\n"
            "  • Security Policy: Automatic login is disabled > [FAIL] Automatic login is on\n"
            "  • Checks completed.\n"
        )
        assert out == expected, f"Test 2 failed: {expected} did not match actual, got \n{out}"

        # Additional debug check
        gettyautologin.succeed("systemctl cat getty@tty1.service | grep -q autologin || echo 'Autologin flag found in getty service'")
        gettyautologin.shutdown()

    # Test 3: Getty autologin as regular user - should fail
    with subtest("Getty autologin as regular user"):
        gettyuser.start()
        gettyuser.systemctl("start network-online.target")
        gettyuser.wait_for_unit("network-online.target")
        gettyuser.wait_for_unit("multi-user.target")

        out = gettyuser.fail("paretosecurity check --only f962c423-fdf5-428a-a57a-816abc9b253e")
        expected = (
            "  • Starting checks...\n"
            "  • Security Policy: Automatic login is disabled > [FAIL] Automatic login is on\n"
            "  • Checks completed.\n"
        )
        assert out == expected, f"Test 3 failed: {expected} did not match actual, got \n{out}"

        # Check for autologin configuration
        gettyuser.succeed("systemctl cat getty@tty1.service | grep -q autologin || echo 'Autologin configured for regular user'")
        gettyuser.shutdown()

    # Test 4: GDM automatic login - should fail
    with subtest("GDM automatic login"):
        gdmautologin.start()
        gdmautologin.systemctl("start network-online.target")
        gdmautologin.wait_for_unit("network-online.target")
        gdmautologin.wait_for_unit("multi-user.target")

        # Check if GDM config file was created
        gdmautologin.succeed("test -f /etc/gdm/custom.conf || test -f /etc/gdm3/custom.conf")
        out = gdmautologin.fail("paretosecurity check --only f962c423-fdf5-428a-a57a-816abc9b253e")
        expected = (
            "  • Starting checks...\n"
            "  • Security Policy: Automatic login is disabled > [FAIL] Automatic login is on\n"
            "  • Checks completed.\n"
        )
        assert out == expected, f"Test 4 failed: {expected} did not match actual, got \n{out}"
        gdmautologin.shutdown()

    # Test 5: GDM timed login with delay - should fail
    with subtest("GDM timed login with delay"):
        gdmtimedlogin.start()
        gdmtimedlogin.systemctl("start network-online.target")
        gdmtimedlogin.wait_for_unit("network-online.target")
        gdmtimedlogin.wait_for_unit("multi-user.target")

        gdmtimedlogin.succeed("test -f /etc/gdm/custom.conf || test -f /etc/gdm3/custom.conf")
        out = gdmtimedlogin.fail("paretosecurity check --only f962c423-fdf5-428a-a57a-816abc9b253e")
        expected = (
            "  • Starting checks...\n"
            "  • Security Policy: Automatic login is disabled > [FAIL] Automatic login is on\n"
            "  • Checks completed.\n"
        )
        assert out == expected, f"Test 5 failed: {expected} did not match actual, got \n{out}"
        gdmtimedlogin.shutdown()

    # Test 6: GDM with zero delay (immediate login) - should fail
    with subtest("GDM with zero delay"):
        gdmzerodelay.start()
        gdmzerodelay.systemctl("start network-online.target")
        gdmzerodelay.wait_for_unit("network-online.target")
        gdmzerodelay.wait_for_unit("multi-user.target")

        gdmzerodelay.succeed("test -f /etc/gdm/custom.conf || test -f /etc/gdm3/custom.conf")
        out = gdmzerodelay.fail("paretosecurity check --only f962c423-fdf5-428a-a57a-816abc9b253e")
        expected = (
            "  • Starting checks...\n"
            "  • Security Policy: Automatic login is disabled > [FAIL] Automatic login is on\n"
            "  • Checks completed.\n"
        )
        assert out == expected, f"Test 6 failed: {expected} did not match actual, got \n{out}"
        gdmzerodelay.shutdown()

    # Test 7: LightDM autologin - should fail
    with subtest("LightDM autologin"):
        lightdmautologin.start()
        lightdmautologin.systemctl("start network-online.target")
        lightdmautologin.wait_for_unit("network-online.target")
        lightdmautologin.wait_for_unit("multi-user.target")

        lightdmautologin.succeed("test -f /etc/lightdm/lightdm.conf")
        out = lightdmautologin.fail("paretosecurity check --only f962c423-fdf5-428a-a57a-816abc9b253e")
        expected = (
            "  • Starting checks...\n"
            "  • Security Policy: Automatic login is disabled > [FAIL] Automatic login is on\n"
            "  • Checks completed.\n"
        )
        assert out == expected, f"Test 7 failed: {expected} did not match actual, got \n{out}"
        lightdmautologin.shutdown()

    # Test 8: SDDM autologin - should fail
    with subtest("SDDM autologin"):
        sddmautologin.start()
        sddmautologin.systemctl("start network-online.target")
        sddmautologin.wait_for_unit("network-online.target")
        sddmautologin.wait_for_unit("multi-user.target")

        sddmautologin.succeed("test -f /etc/sddm.conf || test -d /etc/sddm.conf.d")
        out = sddmautologin.fail("paretosecurity check --only f962c423-fdf5-428a-a57a-816abc9b253e")
        expected = (
            "  • Starting checks...\n"
            "  • Security Policy: Automatic login is disabled > [FAIL] Automatic login is on\n"
            "  • Checks completed.\n"
        )
        assert out == expected, f"Test 8 failed: {expected} did not match actual, got \n{out}"
        sddmautologin.shutdown()
  '';
}
