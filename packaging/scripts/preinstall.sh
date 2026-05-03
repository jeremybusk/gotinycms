#!/usr/bin/env sh
set -eu

if ! getent group uvoominicms >/dev/null 2>&1; then
  groupadd --system uvoominicms
fi

if ! id -u uvoominicms >/dev/null 2>&1; then
  nologin=/usr/sbin/nologin
  if [ ! -x "$nologin" ]; then
    nologin=/sbin/nologin
  fi

  useradd --system \
    --gid uvoominicms \
    --home-dir /var/lib/uvoominicms \
    --shell "$nologin" \
    uvoominicms
fi
