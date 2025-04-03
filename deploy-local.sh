#!/bin/bash

make install

IMG=example.com/controller-manager:$1

make docker-build IMG=$IMG

kind load docker-image $IMG --name kind

make deploy IMG=$IMG
