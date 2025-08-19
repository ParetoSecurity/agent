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

    # Getty autologin once - should fail (creates marker file)
    gettyonce = {pkgs, ...}: {
      imports = [
        (testUser {})
      ];
      services.paretosecurity.enable = true;
      services.getty.autologinUser = "testuser";
      services.getty.autologinOnce = true;
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

    # SDDM autologin - should fail
    sddmautologin = {pkgs, ...}: {
      imports = [
        (testUser {})
      ];
      services.paretosecurity.enable = true;
      services.xserver.enable = true;
      services.xserver.desktopManager.plasma5.enable = true;
      services.displayManager.sddm.enable = true;
      services.displayManager.defaultSession = "plasma";
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

        # Verify no autologin files exist
        noautologin.fail("test -f /run/agetty.autologged")
        noautologin.fail("test -f /etc/gdm/custom.conf")
        noautologin.fail("test -f /etc/sddm.conf")

        out = noautologin.succeed("paretosecurity check --only f962c423-fdf5-428a-a57a-816abc9b253e")
        expected = (
            "  • Starting checks...\n"
            "  • Access Security: Automatic login is disabled > [OK] Automatic login is off\n"
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

        # Debug: Check what systemd service files are created
        print("Checking getty service configuration...")
        gettyautologin.succeed("ls -la /etc/systemd/system/ | grep getty || true")

        # Check if autologin is configured in the service
        autologin_check = gettyautologin.succeed("systemctl cat getty@tty1.service | grep -E 'autologin|ExecStart' || true")
        print(f"Getty service config: {autologin_check}")

        # Verify autologin is actually configured
        gettyautologin.succeed("systemctl cat getty@tty1.service | grep -q 'autologin.*root'")

        out = gettyautologin.fail("paretosecurity check --only f962c423-fdf5-428a-a57a-816abc9b253e")
        expected = (
            "  • Starting checks...\n"
            "  • Access Security: Automatic login is disabled > [FAIL] Automatic login is on\n"
            "  • Checks completed.\n"
        )
        assert out == expected, f"Test 2 failed: {expected} did not match actual, got \n{out}"
        gettyautologin.shutdown()

    # Test 3: Getty autologin as regular user - should fail
    with subtest("Getty autologin as regular user"):
        gettyuser.start()
        gettyuser.systemctl("start network-online.target")
        gettyuser.wait_for_unit("network-online.target")
        gettyuser.wait_for_unit("multi-user.target")

        # Check the getty service configuration
        autologin_check = gettyuser.succeed("systemctl cat getty@tty1.service | grep -E 'autologin.*testuser' || true")
        print(f"Getty service config for testuser: {autologin_check}")

        out = gettyuser.fail("paretosecurity check --only f962c423-fdf5-428a-a57a-816abc9b253e")
        expected = (
            "  • Starting checks...\n"
            "  • Access Security: Automatic login is disabled > [FAIL] Automatic login is on\n"
            "  • Checks completed.\n"
        )
        assert out == expected, f"Test 3 failed: {expected} did not match actual, got \n{out}"
        gettyuser.shutdown()

    # Test 3b: Getty autologin once - should fail due to configuration
    with subtest("Getty autologin once"):
        gettyonce.start()
        gettyonce.systemctl("start network-online.target")
        gettyonce.wait_for_unit("network-online.target")
        gettyonce.wait_for_unit("multi-user.target")

        # Check if the autologin configuration exists
        gettyonce.succeed("systemctl cat getty@tty1.service | grep -q 'autologin.*testuser'")

        # The check should fail because autologin is configured
        out = gettyonce.fail("paretosecurity check --only f962c423-fdf5-428a-a57a-816abc9b253e")
        expected = (
            "  • Starting checks...\n"
            "  • Access Security: Automatic login is disabled > [FAIL] Automatic login is on\n"
            "  • Checks completed.\n"
        )
        assert out == expected, f"Test 3b failed: {expected} did not match actual, got \n{out}"
        gettyonce.shutdown()

    # Test 4: GDM automatic login - should fail
    with subtest("GDM automatic login"):
        gdmautologin.start()
        gdmautologin.systemctl("start network-online.target")
        gdmautologin.wait_for_unit("network-online.target")
        gdmautologin.wait_for_unit("multi-user.target")

        # Check if GDM config file was created and contains autologin settings
        gdm_config = gdmautologin.succeed("cat /etc/gdm/custom.conf 2>/dev/null || cat /etc/gdm3/custom.conf 2>/dev/null || echo 'No GDM config found'")
        print(f"GDM config: {gdm_config}")

        # Verify autologin is configured
        gdmautologin.succeed("grep -q 'AutomaticLogin' /etc/gdm/custom.conf 2>/dev/null || grep -q 'AutomaticLogin' /etc/gdm3/custom.conf 2>/dev/null")

        out = gdmautologin.fail("paretosecurity check --only f962c423-fdf5-428a-a57a-816abc9b253e")
        expected = (
            "  • Starting checks...\n"
            "  • Access Security: Automatic login is disabled > [FAIL] Automatic login is on\n"
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

        # Check GDM config for timed login settings
        gdm_config = gdmtimedlogin.succeed("cat /etc/gdm/custom.conf 2>/dev/null || cat /etc/gdm3/custom.conf 2>/dev/null || echo 'No GDM config'")
        print(f"GDM timed config: {gdm_config}")

        # Verify timed login is configured
        gdmtimedlogin.succeed("grep -E 'TimedLogin|AutomaticLogin' /etc/gdm/custom.conf 2>/dev/null || grep -E 'TimedLogin|AutomaticLogin' /etc/gdm3/custom.conf 2>/dev/null")

        out = gdmtimedlogin.fail("paretosecurity check --only f962c423-fdf5-428a-a57a-816abc9b253e")
        expected = (
            "  • Starting checks...\n"
            "  • Access Security: Automatic login is disabled > [FAIL] Automatic login is on\n"
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

        # Check GDM config for zero delay setting
        gdm_config = gdmzerodelay.succeed("cat /etc/gdm/custom.conf 2>/dev/null || cat /etc/gdm3/custom.conf 2>/dev/null || echo 'No GDM config'")
        print(f"GDM zero delay config: {gdm_config}")

        out = gdmzerodelay.fail("paretosecurity check --only f962c423-fdf5-428a-a57a-816abc9b253e")
        expected = (
            "  • Starting checks...\n"
            "  • Access Security: Automatic login is disabled > [FAIL] Automatic login is on\n"
            "  • Checks completed.\n"
        )
        assert out == expected, f"Test 6 failed: {expected} did not match actual, got \n{out}"
        gdmzerodelay.shutdown()

    # Test 8: SDDM autologin - should fail
    with subtest("SDDM autologin"):
        sddmautologin.start()
        sddmautologin.systemctl("start network-online.target")
        sddmautologin.wait_for_unit("network-online.target")
        sddmautologin.wait_for_unit("multi-user.target")

        # Check SDDM config
        sddm_config = sddmautologin.succeed("cat /etc/sddm.conf 2>/dev/null || echo 'No SDDM config'")
        print(f"SDDM config: {sddm_config}")

        # Also check for conf.d files
        sddmautologin.succeed("ls -la /etc/sddm.conf.d/ 2>/dev/null || true")

        # Verify autologin is configured (check both main config and conf.d)
        sddmautologin.succeed("grep -r 'User=' /etc/sddm.conf /etc/sddm.conf.d/ 2>/dev/null || true")

        out = sddmautologin.fail("paretosecurity check --only f962c423-fdf5-428a-a57a-816abc9b253e")
        expected = (
            "  • Starting checks...\n"
            "  • Access Security: Automatic login is disabled > [FAIL] Automatic login is on\n"
            "  • Checks completed.\n"
        )
        assert out == expected, f"Test 8 failed: {expected} did not match actual, got \n{out}"
        sddmautologin.shutdown()
  '';
}
