{
  pkgs,
  lib,
  config,
  inputs,
  ...
}: let
  flakePackage = import ./package.nix {inherit pkgs lib;};
  upstream = import inputs.upstream {system = pkgs.stdenv.system;};
in {
  packages = [
    upstream.alejandra
    upstream.goreleaser
    upstream.go_1_24
  ];
  languages.nix.enable = true;

  env.GOROOT = upstream.go_1_24 + "/share/go/";
  env.GOPATH = config.env.DEVENV_STATE + "/go";
  env.GOTOOLCHAIN = "local";

  apple.sdk = null; # use installed apple sdk, fixes broken nix packages

  scripts.help-scripts.description = "List all available scripts";
  scripts.help-scripts.exec = ''
    echo
    echo Helper scripts:
    echo
    ${upstream.gnused}/bin/sed -e 's| |••|g' -e 's|=| |' <<EOF | ${upstream.util-linuxMinimal}/bin/column -t | ${upstream.gnused}/bin/sed -e 's|••| |g'
    ${lib.generators.toKeyValue {} (lib.filterAttrs (name: _: name != "help-scripts") (lib.mapAttrs (name: value: value.description) config.scripts))}
    EOF
    echo
  '';

  scripts.coverage.description = "Run tests and check coverage";
  scripts.coverage.exec = ''
    go test -coverprofile=coverage.txt ./...
    coverage=$(go tool cover -func=coverage.txt | grep total | awk '{print $3}' | tr -d %)
    if [ -n "$coverage" ] && [ "$(echo "$coverage" | sed 's/\..*//')" -lt 45 ]; then
      echo "Error: Test coverage is below 45% at $coverage%"
      exit 1
    fi
    echo "Test coverage: $coverage%"
  '';

  scripts.verify-package.description = "Verify package.nix hash";
  scripts.verify-package.exec = ''
    output=$(nix run .# -- --help 2>&1)
    specified=$(echo "$output" | grep "specified:" | awk '{print $2}')
    got=$(echo "$output" | grep "got:" | awk '{print $2}')
    echo "Specified: $specified"
    echo "Got: $got"
    if [ "$specified" != "$got" ]; then
      echo "Mismatch detected, updating package.nix hash from $specified to $got"
      sed -i -e "s|$specified|$got|g" ./package.nix
    else
      echo "Hashes match; no update required."
    fi
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
  pre-commit.hooks = {
    alejandra.enable = true;
    gofmt.enable = true;
    # golangci-lint.enable = true;
    # revive.enable = true;
    packaga-sha = {
      name = "Verify package.nix hash";
      enable = true;
      pass_filenames = false;
      files = "go.(mod|sum)$";
      entry = "verify-package";
    };
  };

  # See full reference at https://devenv.sh/reference/options/
}
