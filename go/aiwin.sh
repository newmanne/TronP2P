#!/bin/bash
# NOTE: make sure that grid width and grid height are in line with that java says!!!

for i in {1..5}
	do
		go run go/server.go "" localhost:8081 false 200 200 applekid &
	done
