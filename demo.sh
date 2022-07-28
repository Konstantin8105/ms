#!/bin/bash

go run ./demo.go > err.out 2>&1
cat err.out
