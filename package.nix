{
  pkgs,
  lib,
}:
pkgs.buildGo124Module rec {
  pname = "paretosecurity";
  version = "${builtins.hashFile "sha256" "${toString ./go.sum}"}";
  src = ./.;
  vendorHash = "sha256-mU9nPd49lL1Glms5rSmD1SGdeFfjSPiYoUgb1xGUXtA=";
  doCheck = true;
}
