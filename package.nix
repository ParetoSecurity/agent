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
    vendorHash = "sha256-oAIRBSo8HdTLsjmeQz5I22RMbxfdtiIe9w71QGt5JiQ=";

    # Uncomment this while developing to skip Go tests
    # doCheck = false;
  })
