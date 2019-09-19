#!/usr/bin/env bash

server="http://localhost:8080"
lowVoltage='false'

_randData() {
    echo "$(( $RANDOM % 1000 )).$(( $RANDOM % 99 ))"
}

_randDelta() {
    echo "$(( $RANDOM % 1 )).$(( $RANDOM % 99 ))"
}

_randVersion() {
    echo "v0.$(( $RANDOM % 1 )).$(( $RANDOM % 9 + 1 ))"
}

_randVoltage() {
    voltage=$(( $RANDOM % 5 + 2 ))
    if [[ ${voltage} -eq 3 ]]; then
        lowVoltage='true'
    fi
    echo ${voltage}.$(( $RANDOM % 99 ))
}

while true; do 
    curl --silent --insecure -X POST -H "Content-Type: application/json" -d "{\"key\":\"Kitchen\", \"delta0\":\"$(_randDelta)\", \"delta1\":\"$(_randDelta)\", \"ch0\":\"$(_randData)\", \"ch1\":\"$(_randData)\", \"voltage\":\"$(_randVoltage)\", \"voltage_low\":\"${lowVoltage}\", \"version\":\"$(_randVersion)\", \"version_esp\":\"$(_randVersion)\"}" ${server}/data
    curl --silent --insecure -X POST -H "Content-Type: application/json" -d "{\"key\":\"Bathroom\", \"delta0\":\"$(_randDelta)\", \"delta1\":\"$(_randDelta)\", \"ch0\":\"$(_randData)\", \"ch1\":\"$(_randData)\", \"voltage\":\"$(_randVoltage)\", \"voltage_low\":\"${lowVoltage}\", \"version\":\"$(_randVersion)\", \"version_esp\":\"$(_randVersion)\"}" ${server}/data
    sleep 5
done