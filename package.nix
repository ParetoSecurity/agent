{
  pkgs,
  lib,
}: let
  nixpkgsPareto = pkgs.callPackage (pkgs.path + "/pkgs/by-name/pa/paretosecurity/package.nix") {};
in
  nixpkgsPareto.overrideAttrs (oldAttrs: {
    src = ./.;
    version = "${builtins.hashFile "sha256" "${toString ./go.sum}"}";

    # Update this when Go dependencies change
    vendorHash = "sha256-7mKAFkKGpBOjXc3J/sfF3k3pJF53tFybXZgbfJInuSY=";

    # Uncomment this while developing to skip Go tests
    # doCheck = false;
  })
