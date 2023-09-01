#!/bin/bash

echo "# go:generate _scripts/be-pkg-list.sh"

PKG_PREFIX="github.com/go-enjin/be"

(
    echo "// Code generated with _scripts/bg-pkg-list.sh DO NOT EDIT.

// Copyright (c) $(date +%Y)  The Go-Enjin Authors
//
// Licensed under the Apache License, Version 2.0 (the \"License\");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an \"AS IS\" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package _templates

func GoEnjinPackageList() (list []string) {
	list = append(list,
		\"${PKG_PREFIX}\","

    for PARENT in pkg features types drivers presets
    do
        find ${PARENT} -type d | sort -V | while read DIR
        do
            fileCount=$(ls -1 "${DIR}" | egrep '\.go' | wc -l)
            if [ ${fileCount} -gt 0 ]
            then
                echo -e "\t\t\"${PKG_PREFIX}/${DIR}\","
            fi
        done
    done

    echo -e "\t)"
    echo -e "\treturn"
    echo "}"
) > _templates/pkg_list.go

sha256sum _templates/pkg_list.go