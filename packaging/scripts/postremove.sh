#!/usr/bin/env sh
set -eu

if command -v systemctl >/dev/null 2>&1; then
  systemctl daemon-reload || true
fi

cat <<'MSG'

UvooMiniCMS package removed.

The service user, configuration, and data were left in place:
  /etc/uvoominicms
  /var/lib/uvoominicms

MSG
