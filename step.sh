#!/bin/bash

set -e

RESTORE='\033[0m'
RED='\033[00;31m'
YELLOW='\033[00;33m'
BLUE='\033[00;34m'
GREEN='\033[00;32m'

function color_echo {
	color=$1
	msg=$2
	echo -e "${color}${msg}${RESTORE}"
}

function echo_fail {
	msg=$1
	echo
	color_echo "${RED}" "${msg}"
	exit 1
}

function echo_warn {
	msg=$1
	color_echo "${YELLOW}" "${msg}"
}

function echo_info {
	msg=$1
	echo
	color_echo "${BLUE}" "${msg}"
}

function echo_details {
	msg=$1
	echo "  ${msg}"
}

function echo_done {
	msg=$1
	color_echo "${GREEN}" "  ${msg}"
}

function validate_required_input {
	key=$1
	value=$2
	if [ -z "${value}" ] ; then
		echo_fail "[!] Missing required input: ${key}"
	fi
}

#=======================================
# Main
#=======================================

#
# Validate parameters
echo_info "Configs:"
echo_details "* workdir: ${workdir}"
echo_details "* command: ${command}"

validate_required_input "command" "${command}"

if [ ! -z "${workdir}" ] ; then
  echo_info "Switching to working directory: ${workdir}"

  cd "${workdir}"
  if [ $? -ne 0 ] ; then
    echo_fail "Failed to switch to working directory: ${workdir}"
  fi
fi

echo_info "Npm version"

npm --version

echo_info "Run npm command"

set -x
npm ${command}
