#!/bin/bash
TEST_MESSAGE="Test message"
RESPONSE=$(docker run --rm --network tp0_testing_net alpine sh -c "echo '$TEST_MESSAGE' | nc server 12345")

if [ "$RESPONSE" = "$TEST_MESSAGE" ]; then
    echo "action: test_echo_server | result: success"
else
    echo "action: test_echo_server | result: fail"
fi
