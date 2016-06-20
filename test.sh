#!/usr/bin/env bash


export ROOT_URL_PREFIX=lmac
export ROOT_URL_DOMAIN=mobibeam.com
docker stop skydock
docker rm skydock
docker run --restart=always -d -v /var/run/docker.sock:/docker.sock --name skydock -l com.mobibeam.docker.name=skydock docker.mobibeam.com/skydock -ttl 600 -environment "$ROOT_URL_PREFIX" -s "/docker.sock" -domain "$ROOT_URL_DOMAIN" -skydns "http://skydns.mobibeam.com:8080" -secret alamakaca