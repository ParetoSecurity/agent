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
    vendorHash = "sha256-97/yaZU0LrqyFoEkDojz1Affdpi5vG+xGRpLNVTZ97Y=";

    # Uncomment this while developing to skip Go tests
    # doCheck = false;
  })
