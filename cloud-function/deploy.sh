#!/bin/sh

gcloud functions deploy google-compute-snapshot \
                        --region europe-west1 \
                        --runtime go111 \
                        --entry-point SnapshotPubSub \
                        --trigger-topic snapshots
