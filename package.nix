{
  # pkgs,
  paretosecurity,
  lib,
}: let
  # nixpkgsPareto = pkgs.callPackage (pkgs.path + "/pkgs/by-name/pa/paretosecurity/package.nix") {};

  # Create a fake src with rev attribute
  srcWithRev = {
    outPath = ./.;
    rev = lib.substring 0 8 (builtins.hashFile "sha256" ./go.sum);
  };
in
  paretosecurity.overrideAttrs (oldAttrs: {
    src = srcWithRev;
    version = "${builtins.hashFile "sha256" "${toString ./go.sum}"}";

    # Updated with pre-commit, don't change manually
    vendorHash = "sha256-YoztEDfflo4sFa7MjB5yoKdXwlAPVb+5rNQ8iv+y3d8=";

    # Uncomment this while developing to skip Go tests
    # doCheck = false;
  })
