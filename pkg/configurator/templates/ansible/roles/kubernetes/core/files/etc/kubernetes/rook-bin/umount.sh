#!/bin/bash

TIMEOUT_IN_SECONDS=5

# umount the rook ceph volumes
for MNT in $(awk '$9 == "ceph" {print $5}' /proc/self/mountinfo); do
  timeout ${TIMEOUT_IN_SECONDS} umount "${MNT}" || timeout ${TIMEOUT_IN_SECONDS} umount -l "${MNT}" || timeout ${TIMEOUT_IN_SECONDS} umount -f "${MNT}"
done
