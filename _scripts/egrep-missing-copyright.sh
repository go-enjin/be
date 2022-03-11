#!/bin/bash

find * -name "*.go" | while read NAME
do
    if ! egrep -q ' Copyright ' "${NAME}"
    then
        echo "${NAME}"
    fi
done
