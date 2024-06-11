#!/bin/bash

# go run で起動したプロセスを特定し、killするスクリプト

# 一時ファイルの名前を特定する正規表現
REGEX='.*go-build.*'

# 'go run'プロセスを特定する
PIDS=$(ps aux | grep 'go run' | grep -v 'grep' | awk '{print $2}')

# 一時ファイル名のプロセスを特定する
TMP_PIDS=$(ps aux | grep -E "$REGEX" | grep -v 'grep' | awk '{print $2}')

# 'go run'プロセスを終了する
if [ -z "$PIDS" ]; then
  echo "No 'go run' processes are running."
else
  echo "Stopping 'go run' processes..."
  kill -9 $PIDS
  echo "'go run' processes stopped."
fi

# 一時ファイル名のプロセスを終了する
if [ -z "$TMP_PIDS" ]; then
  echo "No temporary build processes are running."
else
  echo "Stopping temporary build processes..."
  kill -9 $TMP_PIDS
  echo "Temporary build processes stopped."
fi
