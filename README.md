<h1 align="center">
  <img src="https://avatars.githubusercontent.com/u/87074796?s=200&v=4" width = "200" height = "200">
  <br />
  ParetoSecurity
</h1>

<p align="center">
<a href="https://raw.githack.com/wiki/ParetoSecurity/agent/coverage.html"><img src="https://github.com/ParetoSecurity/agent/wiki/coverage.svg" alt="OpenSSF Scorecard"></a>
<img src="https://img.shields.io/github/downloads/ParetoSecurity/agent/total?label=Downloads" alt="Downloads">
<img src="https://api.scorecard.dev/projects/github.com/ParetoSecurity/agent/badge" alt="OpenSSF Scorecard">
<img src="https://github.com/ParetoSecurity/agent/actions/workflows/build.yml/badge.svg" alt="Integration Tests">
<img src="https://github.com/ParetoSecurity/agent/actions/workflows/unit.yml/badge.svg" alt="Unit Tests">
<img src="https://github.com/ParetoSecurity/agent/actions/workflows/release.yml/badge.svg" alt="Release">
</p>



# Check for basic security hygiene of any Linux desktop

Pareto Desktop is a standalone open-source app that makes sure your Linux device is correctly configured for security. It checks for 13 most imporant security settings, aimed at regular users, not security geeks. It runs automagically in the background via a systray icon, or as a one-off CLI cool. 

Documentation on https://paretosecurity.com/docs/linux/install.

![linux-2000w](https://github.com/user-attachments/assets/0a5a8572-2359-48ee-971b-1bd1f4d5c384)


## Development

### Running Integration Tests on macOS

To run NixOS VM-based integration tests on macOS (including Apple Silicon), you need to set up a Linux builder:

#### Option 1: Configure Linux Builder in nix.conf
```bash
# Add to ~/.config/nix/nix.conf or /etc/nix/nix.conf:
builders = ssh-ng://builder@linux-builder aarch64-linux /etc/nix/builder_ed25519 4 - - - c3NoLWVkMjU1MTkgQUFBQUMzTnphQzFsWkRJMU5URTVBQUFBSUpCV2N4Yi9CbGFxdDFhdU90RStGOFFVV3JVb3RpQzVxQkorVXVFV2RWQ2Igcm9vdEBuaXhvcwo=
builders-use-substitutes = true

# Then restart the nix daemon
sudo launchctl kickstart -k system/org.nixos.nix-daemon
```

#### Option 2: Use darwin.linux-builder module
```nix
# In your nix-darwin configuration:
{
  nix.linux-builder = {
    enable = true;
    ephemeral = true;
    maxJobs = 4;
  };
}
```

#### Option 3: Manual setup
```bash
# First, ensure the Linux builder VM is running (keep this terminal open):
nix run nixpkgs#darwin.linux-builder

# Wait for the VM to fully start, then in a new terminal, test the connection:
ssh -p 31022 builder@localhost

# If SSH connection works, configure Nix to use the builder.
# The correct format includes the port in the SSH URL:
nix build .#checks.aarch64-darwin.xfce \
  --builders 'ssh-ng://builder@localhost:31022 aarch64-linux - - - - -'

# Or use the store option directly:
nix build .#checks.aarch64-darwin.xfce \
  --store ssh-ng://builder@localhost:31022 \
  --eval-store auto

# Alternative: Configure SSH to use port 31022 for localhost
echo "Host linux-builder
  HostName localhost
  Port 31022
  User builder" >> ~/.ssh/config

# Then use the configured host:
nix build .#checks.aarch64-darwin.xfce \
  --builders 'ssh-ng://builder@linux-builder aarch64-linux - - - - -'
```

Note: The error "a 'aarch64-linux' with features {} is required to build" occurs when trying to build Linux packages on macOS without a Linux builder configured. The flake's `allowUnsupportedSystem = true` allows tests to be queued on Darwin, but they still need a Linux builder to actually execute.

## Installation

### Using Debian/Ubuntu/Pop!_OS/RHEL/Fedora/CentOS

See [https://pkg.paretosecurity.com](https://pkg.paretosecurity.com) for install steps.


#### Quick Start

To run a one-time security audit:

```bash
paretosecurity check
```

### Using Nix

<details>
<summary>
  
### Install from nixpkgs

</summary>

#### Install CLI from nixpkgs

```ShellSession
$ nix-env -iA nixpkgs.paretosecurity
```

or

```ShellSession
$ nix profile install nixpkgs#paretosecurity
```

</details>

<details>
<summary>
  
### Install on NixOS

</summary>

#### Install NixOS module

Add this to your NixOS configuration:

```nix
{
  services.paretosecurity.enable = true;
}
```

This will install the agent and its root helper so you don't need `sudo` to run it.

#### Install CLI only in NixOS via nixpkgs

Add this to your NixOS configuration:

```nix
{ pkgs, ... }: {
  environment.systemPackages = [ pkgs.paretosecurity ];
}
```

#### Run checks

```ShellSession
$ paretosecurity check
```

This will analyze your system and provide a security report highlighting potential improvements and vulnerabilities.

If you did not install the root helper, you need to run it with `sudo`:

```ShellSession
$ sudo paretosecurity check
```

</details>

<details>
<summary>
  
### Install via nix-channel

</summary>

As root run:

```ShellSession
$ sudo nix-channel --add https://github.com/ParetoSecurity/agent/archive/main.tar.gz paretosecurity
$ sudo nix-channel --update
```

#### Install CLI via nix-channel

To install the `paretosecurity` binary:

```nix
{
  environment.systemPackages = [ (pkgs.callPackage <paretosecurity/pkgs/paretosecurity.nix> {}) ];
}
```

#### Run checks

```bash
paretosecurity check
```

This will analyze your system and provide a security report highlighting potential improvements and vulnerabilities.

</details>


<details>
<summary>

### Install via Flakes

</summary>


#### Install CLI via Flakes

Using [NixOS module](https://wiki.nixos.org/wiki/NixOS_modules)
(replace system "x86_64-linux" with your system):

```nix
{
  environment.systemPackages = [ paretosecurity.packages.x86_64-linux.default ];
}
```

e.g. inside your `flake.nix` file:

```nix
{
  inputs.paretosecurity.url = "github:paretosecurity/agent";
  # ...

  outputs = { self, nixpkgs, paretosecurity }: {
    # change `yourhostname` to your actual hostname
    nixosConfigurations.yourhostname = nixpkgs.lib.nixosSystem {
      system = "x86_64-linux";
      modules = [
        # ...
        {
          environment.systemPackages = [ paretosecurity.packages.${system}.default ];
        }
      ];
    };
  };
}
```

#### Run checks

```bash
paretosecurity check
```

This will analyze your system and provide a security report highlighting potential improvements and vulnerabilities.
</details>
