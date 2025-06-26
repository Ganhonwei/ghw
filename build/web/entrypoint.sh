#!/bin/sh
set -e

check_config() {
    echo "Checking config status..."
    while true; do
        if curl -s -f -o /dev/null -w "%{http_code}" "http://127.0.0.1:1234" | grep -q "200"; then
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
    ln -s /config/config.git/assets assets
    ln -s /config/config.git/conf conf
    ln -s /config/config.git/views views
}

main() {    
    check_config
    setup_config
    echo "Starting web with args: $@"

    if [ "${BUILD_TYPE}" = "debug" ]; then
        echo "Starting in debug mode..."
        exec dlv --listen=:40000 --headless=true --api-version=2 --accept-multiclient exec ./web -- "$@"
    else
        echo "Starting in release mode..."
        exec ./web "$@"
    fi
}

main "$@"    