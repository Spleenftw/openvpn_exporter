#!/bin/bash

# /usr/local/bin/check_pivpn.sh

# Set the initial status
status=""

# Function to restart OpenVPN service
restart_openvpn() {
    systemctl restart openvpn_exporter
}

# Check if OpenVPN service is active
if systemctl is-active --quiet openvpn; then
    status="ACTIVE"
else
    status="NON-ACTIVE"
fi

# Display the initial status
echo "Service is $status"

# Continuously monitor the status and restart if it changes
while true; do
    # Check if OpenVPN service is active
    if systemctl is-active --quiet openvpn; then
        new_status="ACTIVE"
    else
        new_status="NON-ACTIVE"
    fi

    # Check if the status has changed
    if [ "$new_status" != "$status" ]; then
        status="$new_status"
        echo "Service is now $status"

        # Restart OpenVPN service
        restart_openvpn
    fi

    # Sleep for a while before checking again
    sleep 5
done
