{
  pkgs,
  lib,
}: let
  nixpkgsPareto = pkgs.callPackage (pkgs.path + "/pkgs/by-name/pa/paretosecurity/package.nix") {};
in
  nixpkgsPareto.overrideAttrs (oldAttrs: {
    src = ./.;
    version = "${builtins.hashFile "sha256" "${toString ./go.sum}"}";

    # Updated with pre-commit, don't change manually
    vendorHash = "sha256-Pw2WnE8LfPj4nklutjHkfwZiaNdfk+K8sX9Th0KYE/k=";

    # Uncomment this while developing to skip Go tests
    # doCheck = false;
  })
