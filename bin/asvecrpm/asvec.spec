Buildroot: ./
Name: asvec
Version: VERSIONHERE
Release: 2
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
"/usr/local/aerospike/bin/asvec"
