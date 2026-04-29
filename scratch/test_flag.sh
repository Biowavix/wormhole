#!/bin/bash
go build -o wh_test .
echo "Default help:"
./wh_test ecs conn --help
echo "---"
echo "Trying to use --profile:"
./wh_test ecs conn --profile test-profile 2>&1 | grep "Error loading AWS config"
