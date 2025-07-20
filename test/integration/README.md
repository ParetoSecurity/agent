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

For a nicer shell, you can SSH into the test VM via a testing backdoor. This is useful for debugging or interacting with the test environment. The first testing node will be accessible via `vsock/3`, the second via `vsock/4`, and so on:

```
$ ssh -o User=root vsock/3
$ ssh -o User=root vsock/4
$ ssh -o User=root vsock/5
...
```

This assumes your test file has the `interactive.sshBackdoor.enable = true;` line.
Read more about the SSH backdoor in the [NixOS manual](https://nixos.org/manual/nixos/stable/#sec-nixos-test-ssh-access).


## Seeing UI changes

A quick way to see the changes you made to the UI is to build the test runner in
`screenlock.nix` and run it on a NixOS (VM) machine, so that QEMU can display the UI.

```console
$ nix build .#checks.x86_64-linux.screenlock.driverInteractive
$ ./result/bin/nixos-test-driver
>>> start_all()
```

A NixOS VM, managed by UTM on a Mac, running the `screenlock` test VMs:

![NixOS VM with screenlock test](vms.png)

