#!/usr/bin/env bash
# Regression tests on Debian for ParetoSecurity/agent#655 and #656.
#
# Runs as root inside a Debian system container (with systemd) where
# ParetoSecurity is installed and /usr/bin/paretosecurity is the build under
# test, so the root helper runs the same code as the CLI.
set -euo pipefail

UPDATES_UUID=7436553a-ae52-479b-937b-2ae14d15a520
TEST_USER=alice

fail() {
	echo "FAIL: $*" >&2
	exit 1
}

as_user() {
	runuser -u "$TEST_USER" -- "$@"
}

# run_check prints the output of `paretosecurity check "$@"` run as the test
# user. The exit code is ignored: failing checks exit non-zero.
run_check() {
	as_user paretosecurity check "$@" 2>&1 || true
}

if ! id "$TEST_USER" >/dev/null 2>&1; then
	useradd -m -s /bin/bash "$TEST_USER"
fi
as_user mkdir -p "/home/$TEST_USER/.config"

# Keep package scripts from starting services, so snapd is installed but
# never runs, as on the machine in #655.
printf '#!/bin/sh\nexit 101\n' >/usr/sbin/policy-rc.d
chmod +x /usr/sbin/policy-rc.d
apt-get update
apt-get install -y --no-install-recommends snapd dpkg-dev
rm -f /usr/sbin/policy-rc.d

echo "== #655: snapd is installed but not running"
systemctl disable --now snapd.service snapd.socket >/dev/null 2>&1 || true
command -v snap >/dev/null || fail "snap is not installed"
if [ -S /run/snapd.socket ]; then
	fail "snapd socket is still there, the test needs snapd stopped"
fi

start=$(date +%s)
output=$(run_check)
elapsed=$(($(date +%s) - start))
echo "$output"
echo "check took ${elapsed}s"
if grep -q "timed out" <<<"$output"; then
	fail "#655: check run timed out with snapd stopped"
fi
grep -q "Checks completed." <<<"$output" || fail "#655: check run did not complete"
[ "$elapsed" -lt 120 ] || fail "#655: check run took ${elapsed}s with snapd stopped"

echo "== #655: a hanging command must not hang the check run"
# This flatpak never returns. Without exec, sleep keeps the output pipe open.
cat >/usr/local/bin/flatpak <<'FLATPAK'
#!/bin/sh
sleep 600
FLATPAK
chmod +x /usr/local/bin/flatpak

start=$(date +%s)
output=$(run_check)
elapsed=$(($(date +%s) - start))
echo "$output"
echo "check took ${elapsed}s"
rm -f /usr/local/bin/flatpak
pkill -x sleep || true
if grep -q "timed out" <<<"$output"; then
	fail "#655: check run timed out on a hanging command"
fi
grep -q "Checks completed." <<<"$output" || fail "#655: check run did not complete"
grep -qF "Could not check for updates: Flatpak" <<<"$output" ||
	fail "#655: the hanging Flatpak is not reported"
[ "$elapsed" -lt 120 ] || fail "#655: check run took ${elapsed}s on a hanging command"

echo "== #656: a failing Flatpak must not hide APT updates"
# An installed package with a newer version in a local repository, so
# `apt list --upgradable` has a pending update.
repo=/srv/pareto-regression-repo
build_dir=$(mktemp -d)
mkdir -p "$repo"
build_package() {
	local version=$1 out=$2 dir
	dir="$build_dir/$version"
	mkdir -p "$dir/DEBIAN"
	chmod 0755 "$dir" "$dir/DEBIAN"
	cat >"$dir/DEBIAN/control" <<CONTROL
Package: pareto-regression-dummy
Version: $version
Architecture: all
Maintainer: Pareto Security CI <ci@example.invalid>
Description: Package used to test the application updates check
CONTROL
	dpkg-deb --build "$dir" "$out" >/dev/null
}
build_package 1.0 "$build_dir/pareto-regression-dummy_1.0_all.deb"
dpkg -i "$build_dir/pareto-regression-dummy_1.0_all.deb"
build_package 2.0 "$repo/pareto-regression-dummy_2.0_all.deb"
(cd "$repo" && dpkg-scanpackages --multiversion . >Packages)
echo "deb [trusted=yes] file:$repo ./" >/etc/apt/sources.list.d/pareto-regression.list
apt-get update
apt list --upgradable 2>/dev/null | grep -q pareto-regression-dummy ||
	fail "the dummy package is not upgradable, the test setup is broken"

# A flatpak that always fails, as when a remote cannot be reached. It sits
# first in the root helper's PATH.
cat >/usr/local/bin/flatpak <<'FLATPAK'
#!/bin/sh
echo "error: Unable to load summary from remote flathub" >&2
exit 1
FLATPAK
chmod +x /usr/local/bin/flatpak

output=$(run_check --only "$UPDATES_UUID")
echo "$output"
grep -qF "Updates available for: APT" <<<"$output" || fail "#656: APT updates are hidden when Flatpak fails"
grep -qF "Could not check for updates: Flatpak" <<<"$output" || fail "#656: Flatpak failure is not reported"
if grep -F "Apps are up to date" <<<"$output" | grep -qF "[OK]"; then
	fail "#656: application updates check passed while Flatpak failed"
fi

echo "== #656 control: without the failing Flatpak only APT is reported"
rm -f /usr/local/bin/flatpak
output=$(run_check --only "$UPDATES_UUID")
echo "$output"
grep -qF "Updates available for: APT" <<<"$output" || fail "control: APT updates are not reported"
if grep -qF "Could not check for updates" <<<"$output"; then
	fail "control: a package manager is reported as failing"
fi

echo "All Debian regression tests passed."
