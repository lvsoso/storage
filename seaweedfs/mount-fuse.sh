#!/usr/bin/env bash

# should have /sbin/weed
sudo mount -t fuse.weed fuse /tmp/fuse -o "filer=localhost:8888,filer.path=/"
