#!/usr/bin/env bash
for line in $(cat empty_list.txt)
do
    rm "${line}"
done
