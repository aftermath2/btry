#!/bin/sh
set -e

# Create bitcoin.conf if it doesn't exist
cp -n /tmp/bitcoin.conf /home/bitcoin/.bitcoin/bitcoin.conf

# Fix ownership
chown -R bitcoin /home/bitcoin

# Run original entrypoint
exec /entrypoint.sh "$@"
