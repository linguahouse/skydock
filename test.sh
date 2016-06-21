#!/usr/bin/env bash


export ROOT_URL_PREFIX=lmac
export ROOT_URL_DOMAIN=linguahouse.com
docker stop skydock
docker rm skydock
#docker run --restart=always -d -v /var/run/docker.sock:/docker.sock --name skydock -l com.mobibeam.docker.name=skydock docker.mobibeam.com/skydock -ttl 600 -environment "$ROOT_URL_PREFIX" -s "/docker.sock" -domain "$ROOT_URL_DOMAIN" -skydns "http://skydns.mobibeam.com:8080" -secret alamakaca -plugins /plugins/rancher.js

docker run --restart=always -d -v /var/run/docker.sock:/docker.sock --name skydock  -l  "io.rancher.container.ip=10.42.246.232/16" docker.mobibeam.com/skydock -ttl 60 -beat 10 -environment "$ROOT_URL_PREFIX" -s "/docker.sock" -domain "$ROOT_URL_DOMAIN" -skydns "http://skydns00.linguahouse.com:8080"  -secret alamakaca -plugins /plugins/rancher.js