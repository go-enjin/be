#!/bin/bash
exec egrep -h 'func \(.* MakeFeature {' "$@" \
    | perl -pe 's!^.*Feature\) ([A-Z].+?)\s*MakeFeature\s*\{!$1 MakeFeature!'
