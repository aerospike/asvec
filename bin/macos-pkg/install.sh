#!/bin/bash
mkdir -p /usr/local/bin/
BIN=asvec-macos-amd64
uname -p |grep arm && BIN=asvec-macos-arm64 || echo "amd"
uname -m |grep arm && BIN=asvec-macos-arm64 || echo "amd"
chmod 755 /Library/asvec/*
rm -f /usr/local/bin/asvec || echo "first_install"
ln -s /Library/asvec/${BIN} /usr/local/bin/asvec
ln -s /Library/asvec/${BIN} /usr/local/aerospike/bin/asvec
mkdir -p /etc/paths.d || echo "path install will fail"
echo "/usr/local/bin" |tee /etc/paths.d/asvec || echo "path install failed"
