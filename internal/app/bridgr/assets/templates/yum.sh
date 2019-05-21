set -e
yum clean all
yum install -y yum-plugin-downloadonly createrepo curl
yumdownloader --resolve --archlist=x86_64 --destdir=/packages/7/x86_64 {{Join . " "}}
cd /packages/7/x86_64
echo "Creating YUM repository..."
createrepo .
