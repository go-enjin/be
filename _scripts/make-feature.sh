#!/bin/bash
SCRIPT_NAME=$(basename $0 ".sh")
SCRIPT_PATH=$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )
TEMPLATE_PATH=$(dirname "${SCRIPT_PATH}")/_templates
FEATURE_TMPL=${TEMPLATE_PATH}/feature.go.tmpl

if [ ! -f "${FEATURE_TMPL}" ]
then
  echo "error: file not found - ${FEATURE_TMPL}"
  exit 1
fi

if [ $# -ne 2 ]
then
  echo "usage: $(basename $0) <pkg> <tag>"
  exit 1
fi

PACKAGE_NAME="$1"
FEATURE_TAG="$2"
CURRENT_YEAR=$(date +%Y)

cat "${FEATURE_TMPL}" \
  | perl -pe "\
s#\{\{\s*\.PackageName\s*\}\}#${PACKAGE_NAME}#; \
s#\{\{\s*\.FeatureTag\s*\}\}#${FEATURE_TAG}#; \
s#\{\{\s*\.CurrentYear\s*\}\}#${CURRENT_YEAR}#; \
"