#!/bin/bash

set -e

ETCD_PID_FILE="{{ etcd_data_directory }}/.pid"
CURR_ETCD_PID=$(pgrep etcd)  # this works on etcd pods as well

mkdir -p "{{ etcd_data_directory }}"
touch "${ETCD_PID_FILE}"

OLD_ETCD_PID=$(cat "${ETCD_PID_FILE}")

if [ "${CURR_ETCD_PID}" = "${OLD_ETCD_PID}" ]; then
  # no change, so exit successfully
  exit 0
fi

# if we reached here, the pid is outdated

# set ionice priority to none on old pid (we make the assumption that ionice hasnt been set for other processes)
if [ ! -z "${OLD_ETCD_PID}"]; then
  ionice -c0 -p "${OLD_ETCD_PID}"
fi

# write current pid to pid file
echo "${CURR_ETCD_PID}" > "${ETCD_PID_FILE}"

# set etcd disk priority to best effort, high priority
ionice -c2 -n0 -p "${CURR_ETCD_PID}"
