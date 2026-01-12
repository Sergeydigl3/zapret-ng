#!/bin/sh
set -e

# Detect init system
detect_init_system() {
    if [ -d /run/systemd/system ]; then
        echo "systemd"
    elif [ -f /sbin/openrc-run ] || [ -f /usr/sbin/openrc-run ]; then
        echo "openrc"
    elif [ -f /sbin/init ] && /sbin/init --version 2>/dev/null | grep -q upstart; then
        echo "upstart"
    else
        echo "sysvinit"
    fi
}

# Create runtime directory if it doesn't exist
mkdir -p /run/zapret

# Set proper permissions
chmod 755 /run/zapret

# Detect and configure init system
INIT_SYSTEM=$(detect_init_system)

echo "Detected init system: $INIT_SYSTEM"

case "$INIT_SYSTEM" in
    systemd)
        # Reload systemd daemon
        if command -v systemctl >/dev/null 2>&1; then
            systemctl daemon-reload || true

            # Enable the service (but don't start it automatically)
            systemctl enable zapret-daemon.service || true

            echo "Zapret daemon service has been enabled."
            echo "To start the service, run: systemctl start zapret-daemon"
            echo "To view status, run: systemctl status zapret-daemon"
        fi
        ;;

    openrc)
        # Make sure the script is executable
        chmod +x /etc/init.d/zapret-daemon

        # Add to default runlevel
        if command -v rc-update >/dev/null 2>&1; then
            rc-update add zapret-daemon default || true
            echo "Zapret daemon service has been added to default runlevel."
            echo "To start the service, run: rc-service zapret-daemon start"
            echo "To view status, run: rc-service zapret-daemon status"
        fi
        ;;

    sysvinit)
        # Make sure the script is executable
        chmod +x /etc/init.d/zapret-daemon

        # Register with update-rc.d (Debian/Ubuntu)
        if command -v update-rc.d >/dev/null 2>&1; then
            update-rc.d zapret-daemon defaults || true
            echo "Zapret daemon service has been registered."
            echo "To start the service, run: service zapret-daemon start"
            echo "To view status, run: service zapret-daemon status"
        # Register with chkconfig (RHEL/CentOS)
        elif command -v chkconfig >/dev/null 2>&1; then
            chkconfig --add zapret-daemon || true
            chkconfig zapret-daemon on || true
            echo "Zapret daemon service has been registered."
            echo "To start the service, run: service zapret-daemon start"
            echo "To view status, run: service zapret-daemon status"
        fi
        ;;

    *)
        echo "Unknown init system. Please manually configure the service."
        ;;
esac

# Create symlink for easy access to zapret CLI if not exists
if [ ! -L /usr/local/bin/zapret ] && [ -f /usr/bin/zapret ]; then
    ln -sf /usr/bin/zapret /usr/local/bin/zapret || true
fi

echo ""
echo "Installation complete!"
echo "Configuration file: /etc/zapret/config.yaml"
echo "Please review and adjust the configuration before starting the service."

exit 0
