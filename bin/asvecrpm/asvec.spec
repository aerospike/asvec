Buildroot: ./
Name: asvec
Version: VERSIONHERE
Release: 1
Summary: Tool for deploying non-prod Aerospike server clusters on docker or in AWS
License: see github.com/aerospike/asvec
Group: aerospike

%define _rpmdir ./
%define _rpmfilename %%{NAME}-linux-%%{ARCH}.rpm
%define _unpackaged_files_terminate_build 0
%define _binaries_in_noarch_packages_terminate_build 0

%description
Tool for deploying non-prod Aerospike server clusters on docker or in AWS

%files
/opt/aerospike/bin/asvec
/etc/aerospike/asvec.yml
/usr/bin/asvec

%install
# Ensure the buildroot directories exist
mkdir -p %{buildroot}/opt/aerospike/bin
mkdir -p %{buildroot}/usr/bin

%prep
ln -sf /opt/aerospike/bin/asvec %{buildroot}/usr/bin/asvec
