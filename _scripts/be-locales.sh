#!/bin/bash
echo "# go:generate _scripts/$(basename $0)"
export GOFLAGS="-tags=all"
exec gotext \
     -srclang=en \
     -declare-var=LocalesCatalog \
     -go-build='!exclude_enjin_locales' \
     update \
     -lang=en,ja \
     -out=/dev/null \
     ./...
