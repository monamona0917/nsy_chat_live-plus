#!/bin/bash

APP_COMMAND="./replive"
APP_NAME="replive"
LOG_FILE="/var/log/${APP_NAME}/monitor.log"

mkdir -p $(dirname $LOG_FILE)

find . -name "replive_*.log" -exec mv {} logs/ \;

start_at=$(date +%s)

while true; do
    # TODO 进程运行了超过24小时且处于半夜4点01分杀掉重启进程
    current_time=$(date +%s)
    if [ $((current_time - start_at)) -gt 86400 ] && [ $(date +%H) -eq 4 ] && [ $(date +%M) -eq 1 ]; then
        echo "$(date): $APP_NAME has been running for more than 24 hours. Restarting it..." >> $LOG_FILE
        pkill -f "$APP_COMMAND"
        $APP_COMMAND &
        echo "$(date): $APP_NAME restarted with PID $!" >> $LOG_FILE
        start_at=$(date +%s)
    fi

    if ! pgrep -f "$APP_COMMAND" > /dev/null; then
        echo "$(date): $APP_NAME is not running. Starting it..." >> $LOG_FILE
        $APP_COMMAND &
        echo "$(date): $APP_NAME started with PID $!" >> $LOG_FILE
    fi
    sleep 5
done