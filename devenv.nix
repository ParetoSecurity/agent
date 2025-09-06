{
  pkgs,
  lib,
  config,
  inputs,
  ...
}:
let
  upstream = import inputs.upstream { inherit (pkgs.stdenv) system; };
in
{
  packages = [
    upstream.goreleaser
    upstream.go_1_24
  ];

  languages.nix.enable = true;

  env.GOROOT = upstream.go_1_24 + "/share/go/";
  env.GOPATH = config.env.DEVENV_STATE + "/go";
  env.GOTOOLCHAIN = "local";

  # apple.sdk = null; # use installed apple sdk, fixes broken nix ui skd packages on darwin, but breaks debugger

  scripts.help-scripts.description = "List all available scripts";
  scripts.help-scripts.exec = ''
    echo
    echo Helper scripts:
    echo
    ${upstream.gnused}/bin/sed -e 's| |••|g' -e 's|=| |' <<EOF | ${upstream.util-linuxMinimal}/bin/column -t | ${upstream.gnused}/bin/sed -e 's|••| |g'
    ${lib.generators.toKeyValue { } (
      lib.filterAttrs (name: _: name != "help-scripts") (
        lib.mapAttrs (_name: value: value.description) config.scripts
      )
    )}
    EOF
    echo
  '';

  scripts.coverage.description = "Run tests and check coverage";
  scripts.coverage.exec = ''
    set -o pipefail
    go test -coverprofile=coverage.txt ./... || exit $?
    coverage=$(go tool cover -func=coverage.txt | grep total | awk '{print $3}' | tr -d %)
    if [ -n "$coverage" ] && [ "$(echo "$coverage" | sed 's/\..*//')" -lt 45 ]; then
      echo "Error: Test coverage is below 45% at $coverage%"
      exit 1
    fi
    echo "Test coverage: $coverage%"
  '';

  scripts.update-vendor-hash.description = "Update vendorHash in flake.nix";
  scripts.update-vendor-hash.exec = ''
    output=$(nix build .# 2>&1 || true)
    specified=$(echo "$output" | grep -o "specified: sha256-[A-Za-z0-9+/=]*" | cut -d' ' -f2)
    got=$(echo "$output" | grep -o "got: *sha256-[A-Za-z0-9+/=]*" | cut -d' ' -f2)
    echo "Specified: $specified"
    echo "Got: $got"
    if [ -n "$specified" ] && [ -n "$got" ] && [ "$specified" != "$got" ]; then
      echo "Mismatch detected, updating flake.nix vendorHash from $specified to $got"
      sed -i -e "s|vendorHash = \"$specified\"|vendorHash = \"$got\"|g" ./flake.nix
    else
      if [ -z "$specified" ] && [ -z "$got" ]; then
        echo "No hash mismatch found in build output."
      else
        echo "Hashes match; no update required."
      fi
    fi
  '';

  scripts.build.description = "Build the project";
  scripts.build.exec = ''
    set -o pipefail
    goreleaser --clean --snapshot
    if [ $? -ne 0 ]; then
      echo "Build failed. Please check the logs for details."
      exit 1
    fi
    echo "Build completed successfully."
    echo "Binaries are available in the ./dist directory."
  '';

  enterShell = ''
    export PATH=$GOPATH/bin:$PATH
    help-scripts

    echo "Hint: Run 'devenv test -d' to run tests"
  '';

  # https://devenv.sh/tests/
  enterTest = ''
    go mod verify
    coverage
  '';

  # https://devenv.sh/pre-commit-hooks/
  git-hooks.hooks = {
    nixfmt-rfc-style.enable = true;
    deadnix.enable = true;
    statix.enable = true;
    gofmt.enable = true;
    # golangci-lint.enable = true;
    # revive.enable = true;
    vendor-hash = {
      name = "Update vendorHash in flake.nix";
      enable = true;
      pass_filenames = false;
      files = "go.(mod|sum)$";
      entry = "update-vendor-hash";
    };
  };

  # See full reference at https://devenv.sh/reference/options/
}
