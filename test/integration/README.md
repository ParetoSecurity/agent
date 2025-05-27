# Integration Tests

This directory contains integration tests for the Pareto Security Agent. These tests verify that the agent works correctly on various Linux distributions and system configurations.

## Running Tests

On NixOS, you can run the tests with the following command:

```console
$ nix build .#checks.x86_64-linux.firewall
$ nix build .#checks.aarch64-linux.firewall
```

On macOS with nix-darwin and linux-builder enabled, you can run the tests with the following command:

```console
$ nix build .#checks.aarch64-darwin.firewall
```


## Debugging Tests

Appending `.driverInteractive` to the test name will build the test runner with interactive mode enabled. This allows you to debug the test by SSHing into the test VM.

```console
$ nix build .#checks.aarch64-darwin.firewall.driverInteractive
./result/bin/nixos-test-driver
>>> start_all()
>>> machine.shell_interact()
```

For a nicer shell, add the following to `firewall.nix`, rebuild the test and run
`start_all()`. Now you can SSH into the test VM with `ssh root@localhost -p2222`.

```nix
  interactive.nodes.machine = { ... }: {
    services.openssh = {
      enable = true;
      settings = {
        PermitRootLogin = "yes";
        PermitEmptyPasswords = "yes";
      };
    };
    security.pam.services.sshd.allowNullPassword = true;
    virtualisation.forwardPorts = [
      { from = "host"; host.port = 2222; guest.port = 22; }
    ];
  };
```

