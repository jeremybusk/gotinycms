#!/usr/bin/env sh
set -eu

mkdir -p /etc/uvoominicms /var/lib/uvoominicms/uploads
chown -R uvoominicms:uvoominicms /var/lib/uvoominicms
chmod 0750 /var/lib/uvoominicms /var/lib/uvoominicms/uploads

if command -v systemctl >/dev/null 2>&1; then
  systemctl daemon-reload || true
  systemctl enable uvoominicms.service || true
fi

cat <<'MSG'

UvooMiniCMS installed.

Before starting the service, edit:
  /etc/uvoominicms/uvoominicms.env

At minimum, set a strong CMS_ADMIN_PASS, then run:
  systemctl start uvoominicms

MSG
