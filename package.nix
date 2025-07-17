{
  pkgs,
  lib,
}: let
  nixpkgsPareto = pkgs.callPackage (pkgs.path + "/pkgs/by-name/pa/paretosecurity/package.nix") {};

  # Create a fake src with rev attribute
  srcWithRev = {
    outPath = ./.;
    rev = lib.substring 0 8 (builtins.hashFile "sha256" ./go.sum);
  };
in
  nixpkgsPareto.overrideAttrs (oldAttrs: {
    src = srcWithRev;
    version = "${builtins.hashFile "sha256" "${toString ./go.sum}"}";

    # Updated with pre-commit, don't change manually
    vendorHash = "sha256-tYpTXCx0vp+JX7keu+m5ePrPlGVJZen38HSfQHS4f7s=";

    # Uncomment this while developing to skip Go tests
    # doCheck = false;
  })
