#!/usr/bin/env sh
set -eu

if command -v systemctl >/dev/null 2>&1; then
  systemctl stop uvoominicms.service >/dev/null 2>&1 || true
  systemctl disable uvoominicms.service >/dev/null 2>&1 || true
fi
