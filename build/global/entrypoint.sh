#!/bin/sh
set -e

check_config() {
    echo "Checking config status..."
    while true; do
        if curl -s -f -o /dev/null -w "%{http_code}" "http://config.global:1234" | grep -q "200"; then
            echo "Config is ready!"
            break
        else
            echo "Waiting for config to be ready..."
            sleep 1
        fi
    done
}

setup_config() {
    echo "Setting up config symlink..."
    ln -s /config/config.git config
}

main() {    
    check_config
    setup_config
    echo "Starting global with args: $@"

    if [ "${BUILD_TYPE}" = "debug" ]; then
        echo "Starting in debug mode..."
        exec dlv --listen=:40000 --headless=true --api-version=2 --accept-multiclient exec ./global -- "$@"
    else
        echo "Starting in release mode..."
        exec ./global "$@"
    fi
}

main "$@"    