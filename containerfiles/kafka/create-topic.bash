#!/usr/bin/env bash
#
# Created by Uğur "vigo" Özyılmazel on 2025-01-07.
# Copyright (c) 2025 VB Yazılım. All rights reserved.
set -e
set -o pipefail
set -o errexit
set -o nounset


echo "hello"

# ls -al /usr/bin

# echo "alo"
# kafka-topics.sh --bootstrap-server localhost:9092 --list
# echo "---"

# echo "list topic(s): $(kafka-topics.sh --bootstrap-server localhost:9092 --list)"

# if kafka-topics.sh --bootstrap-server localhost:9092 --list | grep -q "${KCP_TOPIC_GITHUB}"; then
#     echo "topic: ${KCP_TOPIC_GITHUB} is already exsit"
# else
#     kafka-topics.sh \
#         --bootstrap-server localhost:9092 --create \
#         --topic github \
#         --config retention.ms=5000 &&
#     echo "topic: ${KCP_TOPIC_GITHUB} has been created."
# fi
