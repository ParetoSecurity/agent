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
    vendorHash = "sha256-3plpvwLe32AsGuVzdM2fSmTPkKwRFmhi651NEIRdOxw=";

    # Uncomment this while developing to skip Go tests
    # doCheck = false;
  })
