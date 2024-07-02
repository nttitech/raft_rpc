#!/bin/bash

# サーバーの数
NUM_SERVERS=5

# サーバープロセスを起動
for i in $(seq 0 $(($NUM_SERVERS - 1)))
do
    go run main.go $i $NUM_SERVERS &
done


