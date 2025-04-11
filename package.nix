{
  pkgs,
  lib,
}:
pkgs.buildGo124Module rec {
  pname = "paretosecurity";
  version = "${builtins.hashFile "sha256" "${toString ./go.sum}"}";
  src = ./.;
  vendorHash = "sha256-847zti0TiyJ1kNDUWG2x4OIvgD8XeKtGdmd3To8C8C0=";
  doCheck = true;
}
