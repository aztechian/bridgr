#!/bin/sh

yum clean -y -q all
yum install -y -q yum-plugin-downloadonly createrepo curl
yumdownloader --resolve --archlist=x86_64 --destdir=/packages/7/x86_64 {{Join . " "}}
cd /packages/7/x86_64
echo "Creating YUM repository..."
createrepo .
