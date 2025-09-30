let

  # Create a user for testing
  testUser = _: {
    users.users.testuser = {
      isNormalUser = true;
      description = "Test User";
      password = "test123";
      uid = 1001;
    };
  };
in
{
  name = "Autologin";
  interactive.sshBackdoor.enable = true;

  nodes = {
    # No autologin configured - should pass
    noautologin =
      { ... }:
      {
        imports = [
          (testUser { })
        ];
        services.paretosecurity.enable = true;
      };

    # Getty autologin configured - should fail
    gettyautologin =
      { ... }:
      {
        imports = [
          (testUser { })
        ];
        services.paretosecurity.enable = true;
        services.getty.autologinUser = "root";
      };

    # Getty autologin with regular user - should fail
    gettyuser =
      { ... }:
      {
        imports = [
          (testUser { })
        ];
        services.paretosecurity.enable = true;
        services.getty.autologinUser = "testuser";
      };

    # Getty autologin once - should fail (creates marker file)
    gettyonce =
      { ... }:
      {
        imports = [
          (testUser { })
        ];
        services.paretosecurity.enable = true;
        services.getty.autologinUser = "testuser";
        services.getty.autologinOnce = true;
      };

    # Display manager automatic login - should fail
    gdmautologin =
      { ... }:
      {
        imports = [
          (testUser { })
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
    gdmtimedlogin =
      { ... }:
      {
        imports = [
          (testUser { })
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
    gdmzerodelay =
      { ... }:
      {
        imports = [
          (testUser { })
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
    sddmautologin =
      { ... }:
      {
        imports = [
          (testUser { })
        ];
        services.paretosecurity.enable = true;
        services.xserver.enable = true;
        services.desktopManager.plasma6.enable = true;
        services.displayManager.sddm.enable = true;
        services.displayManager.defaultSession = "plasmax11";
        services.displayManager.autoLogin = {
          enable = true;
          user = "testuser";
        };
      };

    # Greetd with tuigreet (no autologin) - should pass
    greetdnoautologin =
      { pkgs, ... }:
      {
        imports = [
          (testUser { })
        ];
        services.paretosecurity.enable = true;
        services.greetd = {
          enable = true;
          settings = {
            default_session = {
              command = "${pkgs.greetd.tuigreet}/bin/tuigreet --cmd sway";
              user = "greeter";
            };
          };
        };
      };

    # Greetd with autologin - should fail
    greetdautologin =
      { pkgs, ... }:
      {
        imports = [
          (testUser { })
        ];
        services.paretosecurity.enable = true;
        services.greetd = {
          enable = true;
          settings = {
            initial_session = {
              command = "sway";
              user = "testuser";
            };
            default_session = {
              command = "${pkgs.greetd.tuigreet}/bin/tuigreet --cmd sway";
              user = "greeter";
            };
          };
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

        out = gettyautologin.fail("paretosecurity check --only f962c423-fdf5-428a-a57a-816abc9b253e")
        expected = (
            "  • Starting checks...\n"
            "  • Access Security: Automatic login is disabled > [FAIL] Getty autologin detected in systemd service override\n"
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

        out = gettyuser.fail("paretosecurity check --only f962c423-fdf5-428a-a57a-816abc9b253e")
        expected = (
            "  • Starting checks...\n"
            "  • Access Security: Automatic login is disabled > [FAIL] Getty autologin detected in systemd service override\n"
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

        # The check should fail because autologin is configured
        out = gettyonce.fail("paretosecurity check --only f962c423-fdf5-428a-a57a-816abc9b253e")
        expected = (
            "  • Starting checks...\n"
            "  • Access Security: Automatic login is disabled > [FAIL] Getty autologin detected (NixOS /run/agetty.autologged exists)\n"
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

        out = gdmautologin.fail("paretosecurity check --only f962c423-fdf5-428a-a57a-816abc9b253e")
        expected = (
            "  • Starting checks...\n"
            "  • Access Security: Automatic login is disabled > [FAIL] AutomaticLoginEnable=true in GDM is enabled\n"
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

        out = gdmtimedlogin.fail("paretosecurity check --only f962c423-fdf5-428a-a57a-816abc9b253e")
        expected = (
            "  • Starting checks...\n"
            "  • Access Security: Automatic login is disabled > [FAIL] TimedLoginEnable=true in GDM is enabled\n"
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

        out = gdmzerodelay.fail("paretosecurity check --only f962c423-fdf5-428a-a57a-816abc9b253e")
        expected = (
            "  • Starting checks...\n"
            "  • Access Security: Automatic login is disabled > [FAIL] AutomaticLoginEnable=true in GDM is enabled\n"
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

        out = sddmautologin.fail("paretosecurity check --only f962c423-fdf5-428a-a57a-816abc9b253e")
        expected = (
            "  • Starting checks...\n"
            "  • Access Security: Automatic login is disabled > [FAIL] SDDM autologin user is configured\n"
            "  • Checks completed.\n"
        )
        assert out == expected, f"Test 8 failed: {expected} did not match actual, got \n{out}"
        sddmautologin.shutdown()

    # Test 9: Greetd without autologin - should pass
    with subtest("Greetd without autologin"):
        greetdnoautologin.start()
        greetdnoautologin.systemctl("start network-online.target")
        greetdnoautologin.wait_for_unit("network-online.target")
        greetdnoautologin.wait_for_unit("multi-user.target")

        out = greetdnoautologin.succeed("paretosecurity check --only f962c423-fdf5-428a-a57a-816abc9b253e")
        expected = (
            "  • Starting checks...\n"
            "  • Access Security: Automatic login is disabled > [OK] Automatic login is off\n"
            "  • Checks completed.\n"
        )
        assert out == expected, f"Test 9 failed: {expected} did not match actual, got \n{out}"
        greetdnoautologin.shutdown()

    # Test 10: Greetd with autologin - should fail
    with subtest("Greetd with autologin"):
        greetdautologin.start()
        greetdautologin.systemctl("start network-online.target")
        greetdautologin.wait_for_unit("network-online.target")
        greetdautologin.wait_for_unit("multi-user.target")

        out = greetdautologin.fail("paretosecurity check --only f962c423-fdf5-428a-a57a-816abc9b253e")
        expected = (
            "  • Starting checks...\n"
            "  • Access Security: Automatic login is disabled > [FAIL] greetd initial_session autologin is configured\n"
            "  • Checks completed.\n"
        )
        assert out == expected, f"Test 10 failed: {expected} did not match actual, got \n{out}"
        greetdautologin.shutdown()
  '';
}
