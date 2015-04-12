#!/bin/bash
# NOTE: make sure that grid width and grid height are in line with that java says!!!

for i in {1..25}
	do
		NEW_UUID=$(cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 5 | head -n 1)
		NAME=sirAi_$NEW_UUID
		echo "Starting AI player $NAME"
		screen -S $NAME -m -d go run go/server.go "" localhost:8081 false 200 200 $NAME
	done
