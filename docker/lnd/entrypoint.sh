#!/bin/sh
set -e

# Create lnd.conf if it doesn't exist
cp -n /tmp/lnd.conf /home/lnd/.lnd/lnd.conf

# Store test password
echo ${AUTO_UNLOCK_PASSWORD} > /tmp/pwd

# Fix ownership
chown -R lnd /home/lnd

# Start lnd
exec sudo -u lnd "$@"
