#!/bin/bash

# Fix for Docker container credential storage issues
# This sets HOME to /tmp so boba can create its config directory
# regardless of whether you're running as root or non-root

export HOME=/tmp
exec ./boba "$@"