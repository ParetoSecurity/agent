{
  paretosecurity,
  lib,
}: let
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
    vendorHash = "sha256-DlCGCheJHa4HPM7kfX/UbOfLukAiaoP7QZnabkZVASM=";

    # Uncomment this while developing to skip Go tests
    # doCheck = false;
  })
