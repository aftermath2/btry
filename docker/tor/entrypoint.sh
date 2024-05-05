#!/bin/sh
set -e

# Create torrc if it doesn't exist
cp -n /tmp/torrc /etc/tor/torrc

# Change local user ID to match LND's and Bitcoin's
usermod -u "${LOCAL_USER_ID:?}" tor

# Fix ownership
chown -R tor /var/lib/tor
chown -R tor /etc/tor

exec sudo -u tor /usr/bin/tor
