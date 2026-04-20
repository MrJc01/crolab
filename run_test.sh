#!/bin/bash
export CROLAB_HOME=/tmp/pool-test
mkdir -p /tmp/pool-test/.crolab
cat << 'EOC' > /tmp/pool-test/.crolab/config.yaml
cloud_token: "mock-token-123"
cloud_api: "http://127.0.0.1:8844"
EOC
./test-bin run . --plan start | tee test-output.log &
pid=$!
sleep 15
fuser -k 8844/tcp || true
fuser -k 49993/tcp || true
kill -9 $pid || true
cat test-output.log
