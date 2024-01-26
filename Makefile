#!/usr/bin/make --no-print-directory --jobs=1 --environment-overrides -f

# Copyright (c) 2022  The Go-Enjin Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

#: uncomment to echo instead of execute
#CMD=echo

.PHONY: all help compile generate build locales
.PHONY: be-update local unlocal tidy
.PHONY: deps fmt reportcard

MAKEFILE_VERSION = v0.0.3

SHELL = /bin/bash

GOPKG_KEYS := GOXT DJHT GOTS

# Go-Enjin gotext package
GOXT_GO_PACKAGE := github.com/go-enjin/golang-org-x-text
GOXT_LOCAL_PATH := ../../../github.com/go-enjin/golang-org-x-text
GOXT_LATEST_VER := v0.12.1-enjin.2

# Go-Enjin times package
DJHT_GO_PACKAGE := github.com/go-enjin/github-com-djherbis-times
DJHT_LOCAL_PATH := ../../../github.com/go-enjin/github-com-djherbis-times

## go-pkgz auth package
#GPA_GO_PACKAGE := github.com/go-enjin/github-com-go-pkgz-auth
#GPA_LOCAL_PATH := ../../../github.com/go-enjin/github-com-go-pkgz-auth

# Go-Enjin text/scanner package
GOTS_GO_PACKAGE := github.com/go-enjin/go-stdlib-text-scanner
GOTS_LOCAL_PATH := ../../../github.com/go-enjin/go-stdlib-text-scanner

GOLANG ?= 1.21.6
GO_MOD ?= 1021

GOIMPORT_LOCALS += github.com/go-corelibs
GOIMPORT_LOCALS += github.com/go-curses
GOIMPORT_LOCALS += github.com/go-enjin

_INTERNAL_BUILD_LOG_ := /dev/null
#_INTERNAL_BUILD_LOG_ := ./build.log

all: help

help:
	@echo "usage: make <compile|generate|build|locales>"
	@echo "       make <be-update|local|unlocal|tidy>"
	@echo "       make <deps|fmt|reportcard>"

export GO_BIN=$(shell which go)
export GOFMT_BIN=$(shell which gofmt)
export GOIMPORTS_BIN=$(shell which goimports)
export LOCALS_VALUE=$(shell echo "${GOIMPORT_LOCALS}" | perl -pe 's/\s+/,/msg')
export GOCONVEY_BIN=$(shell which goconvey)
export GOVULNCHECK_BIN=$(shell which govulncheck)
export GOCYCLO_BIN=$(shell which gocyclo)
export INEFFASSIGN_BIN=$(shell which ineffassign)
export MISSPELL_BIN=$(shell which misspell)

define __install_dep
if [ -z "$(1)" ]; then \
	echo "# installing $(1)"; \
	${GO_BIN} install "$(2)@latest"; \
else \
	echo "# $(1) installed already"; \
fi
endef

ifeq ($(origin ENJENV_BIN),undefined)
ENJENV_BIN:=$(shell which enjenv)
endif

ifeq ($(origin ENJENV_EXE),undefined)
ENJENV_EXE:=$(shell \
	echo "ENJENV_EXE" >> ${_INTERNAL_BUILD_LOG_}; \
	if [ "${ENJENV_BIN}" != "" -a -x "${ENJENV_BIN}" ]; then \
		echo "${ENJENV_BIN}"; \
	else \
		if [ -x "${PWD}/.enjenv.bin" ]; then \
			echo "${PWD}/.enjenv.bin"; \
		elif [ -d "${PWD}/.bin" -a -x "${PWD}/.bin/enjenv" ]; then \
			echo "${PWD}/.bin/enjenv"; \
		else \
			echo "ERROR"; \
		fi; \
	fi)
endif

ENJENV_DIR_NAME ?= .enjenv
ENJENV_DIR ?= ${ENJENV_DIR_NAME}

ifeq ($(origin ENJENV_PATH),undefined)
ENJENV_PATH := $(shell \
	echo "_enjenv_path" >> ${_INTERNAL_BUILD_LOG_}; \
	if [ -x "${ENJENV_EXE}" ]; then \
		${ENJENV_EXE}; \
	elif [ -d "${PWD}/${ENJENV_DIR}" ]; then \
		echo "${PWD}/${ENJENV_DIR}"; \
	fi)
endif


_ALL_FEATURES_PRESENT := $(shell ${ENJENV_EXE} features list 2>/dev/null)

_ENJENV_PRESENT := $(shell \
	echo "_enjenv_present" >> ${_INTERNAL_BUILD_LOG_}; \
	if [ -n "${ENJENV_EXE}" -a -x "${ENJENV_EXE}" ]; then \
		echo "present"; \
	fi)

_GO_PRESENT := $(shell \
	echo "_go_present" >> ${_INTERNAL_BUILD_LOG_}; \
	if [ -n "$(call _has_feature,go)" ]; then \
		echo "present"; \
	fi)

define _validate_extra_pkgs
$(if ${GOPKG_KEYS},$(foreach key,${GOPKG_KEYS},$(shell \
		if [ \
			-z "$($(key)_GO_PACKAGE)" \
			-o -z "$($(key)_LOCAL_PATH)" \
			-o ! -d "$($(key)_LOCAL_PATH)" \
		]; then \
			echo "echo \"# $(key)_GO_PACKAGE and/or $(key)_LOCAL_PATH not found\"; false;"; \
		fi \
)))
endef

define _make_go_local
echo "_make_go_local $(1) $(2)" >> ${_INTERNAL_BUILD_LOG_}; \
echo "# go.mod local: $(1)"; \
${CMD} ${ENJENV_EXE} go-local "$(1)" "$(2)"
endef

define _make_go_unlocal
echo "_make_go_unlocal $(1)" >> ${_INTERNAL_BUILD_LOG_}; \
echo "# go.mod unlocal $(1)"; \
${CMD} ${ENJENV_EXE} go-unlocal "$(1)"
endef

define _make_extra_pkgs
$(if ${GOPKG_KEYS},$(foreach key,${GOPKG_KEYS},$($(key)_GO_PACKAGE)@$(if $($(key)_LATEST_VER),$($(key)_LATEST_VER),latest)))
endef

define _has_feature
$(shell \
	if [ -n "$(1)" -a "$(1)" != "yarn--" -a "$(1)" != "yarn---install" ]; then \
		for feature in ${_ALL_FEATURES_PRESENT}; do \
			if [ "$${feature}" == "$(1)" ]; then \
				echo "_has_feature $(1) (found)" >> ${_INTERNAL_BUILD_LOG_}; \
				echo "$${feature}"; \
				break; \
			else \
				echo "_has_feature $(1) (is not $${feature})" >> ${_INTERNAL_BUILD_LOG_}; \
			fi; \
		done; \
	fi)
endef

define _source_activate_run
	if [ -f "${ENJENV_PATH}/activate" ]; then \
		source "${ENJENV_PATH}/activate" 2>/dev/null \
		&& ${CMD} ${1} ${2} ${3} ${4} ${5} ${6} ${7} ${8} ${9}; \
	else \
		echo "# missing ${ENJENV_PATH}/activate"; \
	fi
endef

_enjenv:
	@if [ -z "${ENJENV_EXE}" -o ! -x "${ENJENV_EXE}" ]; then \
		echo "# critical error: enjenv not found"; \
		false; \
	fi

_golang: _enjenv
	@if [ -z "$(call _has_feature,golang--build)" ]; then \
		if [ "${GOLANG}" != "" ]; then \
			${CMD} ${ENJENV_EXE} golang init --golang "${GOLANG}"; \
		else \
			${CMD} ${ENJENV_EXE} golang init; \
		fi; \
		${CMD} ${ENJENV_EXE} write-scripts; \
		$(call _source_activate_run,${ENJENV_EXE},golang,setup-nancy); \
	elif [ ! -f "${ENJENV_PATH}/activate" ]; then \
		${CMD} ${ENJENV_EXE} write-scripts; \
	else \
		echo "# golang present"; \
	fi

compile:
	@echo "# compiling all package sources"
	@$(call _source_activate_run,go,build,-v,-tags,all,./...)

generate: _golang
	@echo "# running go generate ./..."
	@$(call _source_activate_run,go,generate,./...)

locales: _golang
	@$(call _source_activate_run,_scripts/be-locales.sh,./...)

build: generate compile

be-update: export GOPROXY=direct
be-update: PKG_LIST = $(call _make_extra_pkgs)
be-update: _golang
	@`echo "_make_be_update" >> ${_INTERNAL_BUILD_LOG_}`
	@$(call _validate_extra_pkgs)
	@echo "# go get ${PKG_LIST}"
	@$(call _source_activate_run,go,get,${_BUILD_TAGS},${PKG_LIST})

local: _golang
	@`echo "_make_extra_locals" >> ${_INTERNAL_BUILD_LOG_}`
	@$(call _validate_extra_pkgs)
	@$(if ${GOPKG_KEYS},$(foreach key,${GOPKG_KEYS},$(call _make_go_local,$($(key)_GO_PACKAGE),$($(key)_LOCAL_PATH));))

unlocal: _golang
	@`echo "_make_extra_unlocals" >> ${_INTERNAL_BUILD_LOG_}`
	@$(call _validate_extra_pkgs)
	@$(if ${GOPKG_KEYS},$(foreach key,${GOPKG_KEYS},$(call _make_go_unlocal,$($(key)_GO_PACKAGE));))

tidy: _golang
	@if [ ${GO_MOD} -le 1017 ]; then \
		echo "# go mod tidy -go=1.16 && go mod tidy -go=1.17"; \
		$(call _source_activate_run,go,mod,tidy,-go=1.16); \
		$(call _source_activate_run,go,mod,tidy,-go=1.17); \
	else \
		echo "# go mod tidy"; \
		$(call _source_activate_run,go,mod,tidy); \
	fi

deps:
	@$(call __install_dep,${GOIMPORTS_BIN},golang.org/x/tools/cmd/goimports)
	@$(call __install_dep,${GOCONVEY_BIN},github.com/smartystreets/goconvey)
	@$(call __install_dep,${GOVULNCHECK_BIN},golang.org/x/vuln/cmd/govulncheck)
	@$(call __install_dep,${GOCYCLO_BIN},github.com/fzipp/gocyclo/cmd/gocyclo)
	@$(call __install_dep,${INEFFASSIGN_BIN},github.com/gordonklaus/ineffassign)
	@$(call __install_dep,${MISSPELL_BIN},github.com/client9/misspell/cmd/misspell)

fmt:
	@echo "# running gofmt -s ..."
	@find * -name "*.go" -print0 | xargs -0 -n 1 ${GOFMT_BIN} -s -w
	@echo "# running goimports ..."
	@find * -name "*.go" -print0 | xargs -0 -n 1 ${GOIMPORTS_BIN} -w --local ${LOCALS_VALUE}

reportcard:
	@echo "# code sanity and style report"
	@echo "#: go vet"
	@${GO_BIN} vet -tags all ./... || true
	@echo "#: gocyclo"
	@${GOCYCLO_BIN} -over 15 `find * -name "*.go"` || true
	@echo "#: ineffassign"
	@${INEFFASSIGN_BIN} ./... || true
	@echo "#: misspell"
	@${MISSPELL_BIN} ./... || true
	@echo "#: gofmt -s"
	@echo -e -n `find * -name "*.go" | while read SRC; do \
	  ${GOFMT_BIN} -s "$${SRC}" > "$${SRC}.fmts"; \
	  if ! cmp "$${SRC}" "$${SRC}.fmts" 2> /dev/null; then \
	    echo "can simplify: $${SRC}\\n"; \
	  fi; \
	  rm -f "$${SRC}.fmts"; \
	done`
	@echo "#: govulncheck"
	@echo -e -n `${GOVULNCHECK_BIN} -tags all ./... \
	  | egrep '^Vulnerability #' \
	  | sort -u -V \
	  | while read LINE; do \
	    echo "$${LINE}\n"; \
	  done`
