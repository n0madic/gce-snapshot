#!/bin/sh

gcloud functions deploy google-compute-snapshot \
                        --runtime go111 \
                        --entry-point SnapshotPubSub \
                        --trigger-topic snapshots
