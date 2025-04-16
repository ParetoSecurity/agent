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



Automatically audit your device for basic security hygiene.

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
