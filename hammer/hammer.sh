#!/bin/bash

# script to run the dummy scrapers and connect clients to each sse endpoint

sources=`./hammer -l`

./hammer -interval 0 &
PID=$!
sleep 0.1
for s in $sources; do
    # repeat this line to run multiple clients for each source
    curl http://localhost:9998/$s/ -H "Last-Event-ID: 0" -s >/dev/null &
done

# run for 30 sec then kill everything off (killing the server should
# cause the curl processes to disconnect and shut down too)
sleep 30
kill $PID

