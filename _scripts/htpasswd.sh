#!/bin/bash

if [ $# -ne 1 ]
then
    echo "usage: $(basename $0) <password>" 1>&2
    exit 1
fi

echo -n "$1" \
    | htpasswd -ni user \
    | perl -pe 's!^\s*$!!msg;s!^\s*user\:(.+?)\s*$!$1\n!;'
