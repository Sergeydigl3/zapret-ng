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

# Detect init system
INIT_SYSTEM=$(detect_init_system)

echo "Stopping zapret-daemon service..."

case "$INIT_SYSTEM" in
    systemd)
        if command -v systemctl >/dev/null 2>&1; then
            # Stop the service
            systemctl stop zapret-daemon.service || true

            # Disable the service
            systemctl disable zapret-daemon.service || true

            # Reload systemd daemon
            systemctl daemon-reload || true

            echo "Zapret daemon service has been stopped and disabled."
        fi
        ;;

    openrc)
        if command -v rc-service >/dev/null 2>&1; then
            # Stop the service
            rc-service zapret-daemon stop || true

            # Remove from default runlevel
            rc-update del zapret-daemon default || true

            echo "Zapret daemon service has been stopped and removed from default runlevel."
        fi
        ;;

    sysvinit)
        # Stop the service
        if [ -x /etc/init.d/zapret-daemon ]; then
            /etc/init.d/zapret-daemon stop || true
        fi

        # Unregister with update-rc.d (Debian/Ubuntu)
        if command -v update-rc.d >/dev/null 2>&1; then
            update-rc.d -f zapret-daemon remove || true
        # Unregister with chkconfig (RHEL/CentOS)
        elif command -v chkconfig >/dev/null 2>&1; then
            chkconfig --del zapret-daemon || true
        fi

        echo "Zapret daemon service has been stopped and unregistered."
        ;;

    *)
        echo "Unknown init system. Please manually stop the service."
        ;;
esac

# Remove symlink
if [ -L /usr/local/bin/zapret ]; then
    rm -f /usr/local/bin/zapret || true
fi

# Clean up PID file
rm -f /run/zapret-daemon.pid || true

echo "Pre-removal cleanup complete."

exit 0
