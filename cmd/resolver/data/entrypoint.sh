#!/bin/bash
set -e

# Initialize Unbound
if [ "$BEACON_RESOLVER_TYPE" = "unbound" ]; then
    exec /usr/bin/supervisord -c /etc/supervisord.conf
else
    exec /app/resolver
fi
