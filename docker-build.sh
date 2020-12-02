#!/bin/bash

docker build -t lo-harbor.yyjzt.com/shengyi/nacos-prometheus-discovery:v1 .
docker push lo-harbor.yyjzt.com/shengyi/nacos-prometheus-discovery:v1
