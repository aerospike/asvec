.NOTPARALLEL:

# TODO: move build/pkg related makefile jobs to pkg/

## requirements: make dpkg rpmbuild upx golang zip wget jq
## macos pkg requirement: https://docker-laptop.s3.eu-west-1.amazonaws.com/Packages.pkg

## environment variable import
SIGNER := "$(asvec_signer)"
INSTALLSIGNER := "$(asvec_installsigner)"
APPLEID := "$(asvec_appleid)"
APPLEPW := "$(asvec_applepw)"
TEAMID := "$(asvec_teamid)"
ROOT_DIR = $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
BIN_DIR = ./bin
COVERAGE_DIR = $(ROOT_DIR)/coverage
COV_UNIT_DIR = $(COVERAGE_DIR)/unit
COV_INTEGRATION_DIR = $(COVERAGE_DIR)/integration

## available make commands
.PHONY: help
help:
	@printf "\nSHORTHANDS:\n\
	\tmake build && make install                 - build and install on current system\n\
	\tmake build-linux && make pkg-linux         - build and package all linux releases\n\
	\n\
	UPDATE COMMANDS:\n\
	\tdeps               - Update all dependencies in the project\n\
	\n\
	BUILD COMMANDS:\n\
	\tbuild              - A version for the current system (linux and mac only)\n\
	\tbuildall           - All versions for all supported systems\n\
	\tbuild-linux-amd64  - Linux on x86_64\n\
	\tbuild-linux-arm64  - Linux on aarch64\n\
	\tbuild-linux        - Linux on x86_64 and aarach64\n\
	\tbuild-darwin-amd64 - MacOS on x86_64\n\
	\tbuild-darwin-arm64 - MacOS on M1/M2 style aarach64\n\
	\tbuild-darwin       - MacOS on x86_64 and aarch64\n\
	\tbuild-windows-amd64- Windows on x86_64 and aarch64\n\
	\tbuild-windows-arm64- Windows on ARM\n\
	\tbuild-windows      - Windows on x86_64 and arm64\n\
	\n\
	INSTALL COMMANDS:\n\
	\tinstall            - Install a previously built asvec on the current system (linux and mac only)\n\
	\n\
	CLEAN COMMANDS:\n\
	\tclean              - Remove remainders of a build and reset source modified during build\n\
	\tcleanall           - Clean and remove all created packages\n\
	\n\
	LINUX PACKAGING COMMANDS:\n\
	\tpkg-linux          - Package all linux packages - zip, rpm and deb\n\
	\tpkg-zip            - Package linux zip\n\
	\tpkg-rpm            - Package linux rpm\n\
	\tpkg-deb            - Package linux deb\n\
	\tpkg-zip-amd64      - Package linux zip for amd64 only\n\
	\tpkg-rpm-amd64      - Package linux rpm for amd64 only\n\
	\tpkg-deb-amd64      - Package linux deb for amd64 only\n\
	\tpkg-zip-arm64      - Package linux zip for arm64 only\n\
	\tpkg-rpm-arm64      - Package linux rpm for arm64 only\n\
	\tpkg-deb-arm64      - Package linux deb for arm64 only\n\
	\n\
	WINDOWS PACKAGING COMMANDS:\n\
	\tpkg-windows-zip       - Package windows zip\n\
	\tpkg-windows-zip-amd64 - Package windows zip for amd64 only\n\
	\tpkg-windows-zip-arm64 - Package windows zip for arm64 only\n\
	\n\
	MACOS PACKAGING COMMANDS:\n\
	\tmacos-codesign     - Codesign MacOS binaries\n\
	\tmacos-pkg-build    - Create MacOS pkg installer\n\
	\tmacos-pkg-sign     - Productsign MacOS pkg installer\n\
	\tmacos-pkg-notarize - Codesign MacOS binaries\n\
	\tmacos-zip-build    - Create MacOS zip packages\n\
	\tmacos-zip-notarize - Notarize MacOS ZIP packages\n\
	\tmacos-build-all    - Build and sign pkg and zip\n\
	\tmacos-notarize-all - Notarize pkg and zip\n\
	\n\
	OUTPUTS: $(BIN_DIR)/ and $(BIN_DIR)/packages/\n\
	"

.PHONY: deps
deps:
	go get -u && go mod tidy && GOWORK=off go mod vendor


.PHONY: macos-build-all
macos-build-all: macos-codesign macos-zip-build macos-pkg-build macos-pkg-sign

.PHONY: macos-notarize-all
macos-notarize-all: macos-pkg-notarize macos-zip-notarize

.PHONY: build
build: run_build

.PHONY: buildall
buildall: prep compile_linux_wip_amd64 compile_linux_wip_arm64 reset1 compile_linux_amd64 compile_linux_arm64 compile_darwin compile_windows reset2

.PHONY: build-linux-amd64
build-linux-amd64: prep compile_linux_wip_arm64 reset1 compile_linux_amd64 reset2

.PHONY: build-linux-arm64
build-linux-arm64: prep compile_linux_wip_amd64 reset1 compile_linux_arm64 reset2

.PHONY: build-linux
build-linux: prep compile_linux_wip_amd64 compile_linux_wip_arm64 reset1 compile_linux_amd64 compile_linux_arm64 reset2

.PHONY: build-darwin-amd64
build-darwin-amd64: prep compile_linux_wip_amd64 compile_linux_wip_arm64 reset1 compile_darwin_amd64 reset2

.PHONY: build-darwin-arm64
build-darwin-arm64: prep compile_linux_wip_amd64 compile_linux_wip_arm64 reset1 compile_darwin_arm64 reset2

.PHONY: build-darwin
build-darwin: prep compile_linux_wip_amd64 compile_linux_wip_arm64 reset1 compile_darwin reset2

.PHONY: build-windows-amd64
build-windows-amd64: prep compile_linux_wip_amd64 compile_linux_wip_arm64 reset1 compile_windows_amd64 reset2

.PHONY: build-windows-arm64
build-windows-arm64: prep compile_linux_wip_amd64 compile_linux_wip_arm64 reset1 compile_windows_arm64 reset2

.PHONY: build-windows
build-windows: prep compile_linux_wip_amd64 compile_linux_wip_arm64 reset1 compile_windows reset2

.PHONY: install
install: run_install

.PHONY: cleanall
cleanall: clean
	rm -f $(BIN_DIR)/packages/*
	rm -f notarize_result_pkg notarize_result_amd64 notarize_result_arm64
	rm -f $(BIN_DIR)/AsVec.pkg

.PHONY: clean
clean:
	rm -f asvec-linux-amd64-wip
	rm -f asvec-linux-arm64-wip
	rm -f *.upx
	rm -f asvec-linux-amd64
	rm -f asvec-linux-arm64
	rm -f asvec-macos-amd64
	rm -f asvec-macos-arm64
	rm -f asvec-windows-amd64.exe
	rm -f asvec-windows-arm64.exe
	rm -f embed_*.txt
	rm -f myFunction.zip
	rm -f gcpMod.txt
	rm -f gcpFunction.txt
	rm -f $(BIN_DIR)/asvec-*
	rm -f $(BIN_DIR)/deb
	rm -f $(BIN_DIR)/deb.deb
	rm -f $(BIN_DIR)/asvec
	printf "package main\n\nvar nLinuxBinaryX64 []byte\n\nvar nLinuxBinaryArm64 []byte\n" > embed_linux.go
	cp embed_linux.go embed_darwin.go
	cp embed_linux.go embed_windows.go

## actual code

OS := $(shell uname -o)
CPU := $(shell uname -m)
ver:=$(shell V=$$(git describe --tags --always); printf $${V} > ./VERSION.md; cat ./VERSION.md)
rpm_ver := $(shell echo $(ver) | sed 's/-/_/g')
$(info ver is $(ver) and rpm_ver is $(rpm_ver))
GO_LDFLAGS="-X 'asvec/cmd.Version=$(ver)' -s -w"
define _amddebscript
ver=$(cat ./VERSION.md)
cat <<EOF > ./bin/deb/DEBIAN/control
Website: www.aerospike.com
Maintainer: Aerospike <support@aerospike.com>
Name: Aerospike Vector CLI
Package: asvec
Section: aerospike
Version: ${ver}
Architecture: amd64
Description: Tool for managing Aerospike Vector Search clusters.
EOF
endef
export amddebscript = $(value _amddebscript)
define _armdebscript
ver=$(cat ./VERSION.md)
cat <<EOF > ./bin/deb/DEBIAN/control
Website: www.aerospike.com
Maintainer: Aerospike <support@aerospike.com>
Name: Aerospike Vector CLI
Package: asvec
Section: aerospike
Version: ${ver}
Architecture: arm64
Description: Tool for managing Aerospike Vector Search clusters.
EOF
endef
export armdebscript = $(value _armdebscript)

.PHONY: run_build
run_build:
ifeq ($(OS), Darwin)
ifeq ($(CPU), x86_64)
	$(MAKE) build-darwin-amd64
else
	$(MAKE) build-darwin-arm64
endif
else
ifeq ($(CPU), x86_64)
	$(MAKE) build-linux-amd64
else
	$(MAKE) build-linux-arm64
endif
endif

.PHONY: run_install
run_install:
ifeq ($(OS), Darwin)
ifeq ($(CPU), x86_64)
	sudo cp $(BIN_DIR)/asvec-macos-amd64 /usr/local/bin/asvec
else
	sudo cp $(BIN_DIR)/asvec-macos-arm64 /usr/local/bin/asvec
endif
else
ifeq ($(CPU), x86_64)
	sudo cp $(BIN_DIR)/asvec-linux-amd64 /usr/local/bin/asvec
else
	sudo cp $(BIN_DIR)/asvec-linux-arm64 /usr/local/bin/asvec
endif
endif

.PHONY: reset1
reset1:
	printf "package main\n\nvar nLinuxBinaryX64 []byte\n\nvar nLinuxBinaryArm64 []byte\n" > embed_linux.go
	cp embed_linux.go embed_darwin.go
	cp embed_linux.go embed_windows.go

.PHONY: reset2
reset2:
	rm -f asvec-linux-amd64-wip
	rm -f asvec-linux-arm64-wip
	rm -f *.upx
	rm -f embed_*.txt
	printf "package main\n\nvar nLinuxBinaryX64 []byte\n\nvar nLinuxBinaryArm64 []byte\n" > embed_linux.go
	cp embed_linux.go embed_darwin.go
	cp embed_linux.go embed_windows.go

.PHONY: prep
prep:
	go generate

.PHONY: compile_linux_wip_amd64
compile_linux_wip_amd64:
	env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags=$(GO_LDFLAGS) -o asvec-linux-amd64-wip
ifneq (, $(shell which upx))
	upx asvec-linux-amd64-wip
endif

.PHONY: compile_linux_wip_arm64
compile_linux_wip_arm64:
	env CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath -ldflags=$(GO_LDFLAGS) -o asvec-linux-arm64-wip
ifneq (, $(shell which upx))
	upx asvec-linux-arm64-wip
endif

.PHONY: compile_linux_amd64
compile_linux_amd64:
	printf "package main\n\nimport _ \"embed\"\n\nvar nLinuxBinaryX64 []byte\n\n//go:embed asvec-linux-arm64-wip\nvar nLinuxBinaryArm64 []byte\n" > embed_linux.go
	env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags=$(GO_LDFLAGS) -o asvec-linux-amd64
	mv asvec-linux-amd64 $(BIN_DIR)/

.PHONY: compile_linux_arm64
compile_linux_arm64:
	printf "package main\n\nimport _ \"embed\"\n\n//go:embed asvec-linux-amd64-wip\nvar nLinuxBinaryX64 []byte\n\nvar nLinuxBinaryArm64 []byte\n" > embed_linux.go
	env CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath -ldflags=$(GO_LDFLAGS) -o asvec-linux-arm64
	mv asvec-linux-arm64 $(BIN_DIR)/

.PHONY: compile_darwin
compile_darwin:
	printf "package main\n\nimport _ \"embed\"\n\n//go:embed asvec-linux-amd64-wip\nvar nLinuxBinaryX64 []byte\n\n//go:embed asvec-linux-arm64-wip\nvar nLinuxBinaryArm64 []byte" > embed_darwin.go
	env CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -trimpath -ldflags=$(GO_LDFLAGS) -o asvec-macos-amd64
	env CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -trimpath -ldflags=$(GO_LDFLAGS) -o asvec-macos-arm64
	mv asvec-macos-amd64 $(BIN_DIR)/
	mv asvec-macos-arm64 $(BIN_DIR)/

.PHONY: compile_darwin_amd64
compile_darwin_amd64:
	printf "package main\n\nimport _ \"embed\"\n\n//go:embed asvec-linux-amd64-wip\nvar nLinuxBinaryX64 []byte\n\n//go:embed asvec-linux-arm64-wip\nvar nLinuxBinaryArm64 []byte" > embed_darwin.go
	env CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -trimpath -ldflags=$(GO_LDFLAGS) -o asvec-macos-amd64
	mv asvec-macos-amd64 $(BIN_DIR)/

.PHONY: compile_darwin_arm64
compile_darwin_arm64:
	printf "package main\n\nimport _ \"embed\"\n\n//go:embed asvec-linux-amd64-wip\nvar nLinuxBinaryX64 []byte\n\n//go:embed asvec-linux-arm64-wip\nvar nLinuxBinaryArm64 []byte" > embed_darwin.go
	env CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -trimpath -ldflags=$(GO_LDFLAGS) -o asvec-macos-arm64
	mv asvec-macos-arm64 $(BIN_DIR)/

.PHONY: compile_windows
compile_windows:
	printf "package main\n\nimport _ \"embed\"\n\n//go:embed asvec-linux-amd64-wip\nvar nLinuxBinaryX64 []byte\n\n//go:embed asvec-linux-arm64-wip\nvar nLinuxBinaryArm64 []byte" > embed_windows.go
	env CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -trimpath -ldflags=$(GO_LDFLAGS) -o asvec-windows-amd64.exe
	env CGO_ENABLED=0 GOOS=windows GOARCH=arm64 go build -trimpath -ldflags=$(GO_LDFLAGS) -o asvec-windows-arm64.exe
	mv asvec-windows-amd64.exe $(BIN_DIR)/
	mv asvec-windows-arm64.exe $(BIN_DIR)/

.PHONY: compile_windows_amd64
compile_windows_amd64:
	printf "package main\n\nimport _ \"embed\"\n\n//go:embed asvec-linux-amd64-wip\nvar nLinuxBinaryX64 []byte\n\n//go:embed asvec-linux-arm64-wip\nvar nLinuxBinaryArm64 []byte" > embed_windows.go
	env CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -trimpath -ldflags=$(GO_LDFLAGS) -o asvec-windows-amd64.exe
	mv asvec-windows-amd64.exe $(BIN_DIR)/

.PHONY: compile_windows_arm64
compile_windows_arm64:
	printf "package main\n\nimport _ \"embed\"\n\n//go:embed asvec-linux-amd64-wip\nvar nLinuxBinaryX64 []byte\n\n//go:embed asvec-linux-arm64-wip\nvar nLinuxBinaryArm64 []byte" > embed_windows.go
	env CGO_ENABLED=0 GOOS=windows GOARCH=arm64 go build -trimpath -ldflags=$(GO_LDFLAGS) -o asvec-windows-arm64.exe
	mv asvec-windows-arm64.exe $(BIN_DIR)/

.PHONY: official
official: prep
	printf "" > embed_tail.txt

.PHONY: prerelease
prerelease: prep
	printf -- "-prerelease" > embed_tail.txt

.PHONY: build-official
build-official: official compile_linux_wip_amd64 compile_linux_wip_arm64 reset1 compile_linux_amd64 compile_linux_arm64 compile_darwin compile_windows reset2

.PHONY: build-prerelease
build-prerelease: prerelease compile_linux_wip_amd64 compile_linux_wip_arm64 reset1 compile_linux_amd64 compile_linux_arm64 compile_darwin compile_windows reset2

RET := $(shell echo)

.PHONY: pkg-deb-amd64
pkg-deb-amd64:
	cp $(BIN_DIR)/asvec-linux-amd64 $(BIN_DIR)/asvec
	rm -rf $(BIN_DIR)/deb
	mkdir -p $(BIN_DIR)/deb/DEBIAN
	mkdir -p $(BIN_DIR)/deb/usr/local/aerospike/bin
	@ eval "$$amddebscript"
	mv $(BIN_DIR)/asvec $(BIN_DIR)/deb/usr/local/aerospike/bin/
	sudo dpkg-deb -Zxz -b $(BIN_DIR)/deb
	rm -f $(BIN_DIR)/packages/asvec-linux-amd64-${ver}.deb
	mv $(BIN_DIR)/deb.deb $(BIN_DIR)/packages/asvec-linux-amd64-${ver}.deb
	rm -rf $(BIN_DIR)/deb

.PHONY: pkg-deb-arm64
pkg-deb-arm64:
	cp $(BIN_DIR)/asvec-linux-arm64 $(BIN_DIR)/asvec
	rm -rf $(BIN_DIR)/deb
	mkdir -p $(BIN_DIR)/deb/DEBIAN
	mkdir -p $(BIN_DIR)/deb/usr/local/aerospike/bin
	@ eval "$$armdebscript"
	mv $(BIN_DIR)/asvec $(BIN_DIR)/deb/usr/local/aerospike/bin/
	sudo dpkg-deb -Zxz -b $(BIN_DIR)/deb
	rm -f $(BIN_DIR)/packages/asvec-linux-arm64-${ver}.deb
	mv $(BIN_DIR)/deb.deb $(BIN_DIR)/packages/asvec-linux-arm64-${ver}.deb
	rm -rf $(BIN_DIR)/deb

.PHONY: pkg-deb
pkg-deb: pkg-deb-amd64 pkg-deb-arm64

.PHONY: pkg-zip-amd64
pkg-zip-amd64:
	cp $(BIN_DIR)/asvec-linux-amd64 $(BIN_DIR)/asvec
	rm -f $(BIN_DIR)/packages/asvec-linux-amd64-${ver}.zip
	bash -ce "cd $(BIN_DIR) && zip packages/asvec-linux-amd64-${ver}.zip asvec"
	rm -f $(BIN_DIR)/asvec

.PHONY: pkg-zip-arm64
pkg-zip-arm64:
	cp $(BIN_DIR)/asvec-linux-arm64 $(BIN_DIR)/asvec
	rm -f $(BIN_DIR)/packages/asvec-linux-arm64-${ver}.zip
	bash -ce "cd $(BIN_DIR) && zip packages/asvec-linux-arm64-${ver}.zip asvec"
	rm -f $(BIN_DIR)/asvec

.PHONY: pkg-windows-zip
pkg-windows-zip: pkg-windows-zip-amd64 pkg-windows-zip-arm64

.PHONY: pkg-windows-zip-amd64
pkg-windows-zip-amd64:
	cp $(BIN_DIR)/asvec-windows-amd64.exe $(BIN_DIR)/asvec.exe
	rm -f $(BIN_DIR)/packages/asvec-windows-amd64-${ver}.zip
	bash -ce "cd $(BIN_DIR) && zip packages/asvec-windows-amd64-${ver}.zip asvec.exe"
	rm -f $(BIN_DIR)/asvec.exe

.PHONY: pkg-windows-zip-arm64
pkg-windows-zip-arm64:
	cp $(BIN_DIR)/asvec-windows-arm64.exe $(BIN_DIR)/asvec.exe
	rm -f $(BIN_DIR)/packages/asvec-windows-arm64-${ver}.zip
	bash -ce "cd $(BIN_DIR) && zip packages/asvec-windows-arm64-${ver}.zip asvec.exe"
	rm -f $(BIN_DIR)/asvec.exe

.PHONY: pkg-zip
pkg-zip: pkg-zip-amd64 pkg-zip-arm64

.PHONY: pkg-rpm-amd64
pkg-rpm-amd64:
	rm -rf $(BIN_DIR)/asvec-rpm-centos
	cp -a $(BIN_DIR)/asvecrpm $(BIN_DIR)/asvec-rpm-centos
	sed -i.bak "s/VERSIONHERE/${rpm_ver}/g" $(BIN_DIR)/asvec-rpm-centos/asvec.spec
	cp $(BIN_DIR)/asvec-linux-amd64 $(BIN_DIR)/asvec-rpm-centos/usr/local/aerospike/bin/asvec
	rm -f $(BIN_DIR)/asvec-linux-x86_64.rpm
	bash -ce "cd $(BIN_DIR) && rpmbuild --target=x86_64-redhat-linux --buildroot \$$(pwd)/asvec-rpm-centos -bb asvec-rpm-centos/asvec.spec"
	rm -f $(BIN_DIR)/packages/asvec-linux-amd64-${rpm_ver}.rpm
	mv $(BIN_DIR)/asvec-linux-x86_64.rpm $(BIN_DIR)/packages/asvec-linux-amd64-${rpm_ver}.rpm

.PHONY: pkg-rpm-arm64
pkg-rpm-arm64:
	rm -rf $(BIN_DIR)/asvec-rpm-centos
	cp -a $(BIN_DIR)/asvecrpm $(BIN_DIR)/asvec-rpm-centos
	sed -i.bak "s/VERSIONHERE/${rpm_ver}/g" $(BIN_DIR)/asvec-rpm-centos/asvec.spec
	cp $(BIN_DIR)/asvec-linux-arm64 $(BIN_DIR)/asvec-rpm-centos/usr/local/aerospike/bin/asvec
	rm -f $(BIN_DIR)/asvec-linux-arm64.rpm
	bash -ce "cd $(BIN_DIR) && rpmbuild --target=arm64-redhat-linux --buildroot \$$(pwd)/asvec-rpm-centos -bb asvec-rpm-centos/asvec.spec"
	rm -f $(BIN_DIR)/packages/asvec-linux-arm64-${rpm_ver}.rpm
	mv $(BIN_DIR)/asvec-linux-arm64.rpm $(BIN_DIR)/packages/asvec-linux-arm64-${rpm_ver}.rpm

.PHONY: pkg-rpm
pkg-rpm: pkg-rpm-amd64 pkg-rpm-arm64

.PHONY: pkg-linux
pkg-linux: pkg-zip pkg-deb pkg-rpm

### note - static linking
###go build -ldflags="-extldflags=-static"

.PHONY: macos-codesign
macos-codesign:
ifeq (exists, $(shell [ -f $(BIN_DIR)/asvec-macos-amd64 ] && echo "exists" || echo "not found"))
	codesign --verbose --deep --timestamp --force --options runtime --sign ${SIGNER} $(BIN_DIR)/asvec-macos-amd64
	codesign --verbose --verify $(BIN_DIR)/asvec-macos-amd64
endif
ifeq (exists, $(shell [ -f $(BIN_DIR)/asvec-macos-arm64 ] && echo "exists" || echo "not found"))
	codesign --verbose --deep --timestamp --force --options runtime --sign ${SIGNER} $(BIN_DIR)/asvec-macos-arm64
	codesign --verbose --verify $(BIN_DIR)/asvec-macos-arm64
endif

.PHONY: macos-zip-build
macos-zip-build:
ifeq (exists, $(shell [ -f $(BIN_DIR)/asvec-macos-amd64 ] && echo "exists" || echo "not found"))
	cp $(BIN_DIR)/asvec-macos-amd64 $(BIN_DIR)/asvec
	rm -f $(BIN_DIR)/packages/asvec-macos-amd64-${ver}.zip
	bash -ce "cd $(BIN_DIR) && zip packages/asvec-macos-amd64-${ver}.zip asvec"
	rm -f $(BIN_DIR)/asvec
endif
ifeq (exists, $(shell [ -f $(BIN_DIR)/asvec-macos-arm64 ] && echo "exists" || echo "not found"))
	cp $(BIN_DIR)/asvec-macos-arm64 $(BIN_DIR)/asvec
	rm -f $(BIN_DIR)/packages/asvec-macos-arm64-${ver}.zip
	bash -ce "cd $(BIN_DIR) && zip packages/asvec-macos-arm64-${ver}.zip asvec"
	rm -f $(BIN_DIR)/asvec
endif

.PHONY: macos-zip-notarize
macos-zip-notarize:
ifeq (exists, $(shell [ -f $(BIN_DIR)/packages/asvec-macos-amd64-${ver}.zip ] && echo "exists" || echo "not found"))
	rm -f notarize_result_amd64
	xcrun notarytool submit --apple-id ${APPLEID} --password ${APPLEPW} --team-id ${TEAMID} -f json --wait --timeout 10m $(BIN_DIR)/packages/asvec-macos-amd64-${ver}.zip > notarize_result_amd64
	if [ "$$(cat notarize_result_amd64 |jq -r .status)" != "Accepted" ] ;\
	then \
		echo "ZIP-AMD FAILED TO NOTARIZE" ;\
		cat notarize_result_amd64 ;\
		exit 1 ;\
	else \
		echo "ZIP-AMD NOTARIZE SUCCESS" ;\
	fi
endif
ifeq (exists, $(shell [ -f $(BIN_DIR)/packages/asvec-macos-arm64-${ver}.zip ] && echo "exists" || echo "not found"))
	rm -f notarize_result_arm64
	xcrun notarytool submit --apple-id ${APPLEID} --password ${APPLEPW} --team-id ${TEAMID} -f json --wait --timeout 10m $(BIN_DIR)/packages/asvec-macos-arm64-${ver}.zip > notarize_result_arm64
	if [ "$$(cat notarize_result_arm64 |jq -r .status)" != "Accepted" ] ;\
	then \
		echo "ZIP-ARM FAILED TO NOTARIZE" ;\
		cat notarize_result_arm64 ;\
		exit 1 ;\
	else \
		echo "ZIP-ARM NOTARIZE SUCCESS" ;\
	fi
endif

.PHONY: macos-pkg-build
macos-pkg-build:
	cp -a $(BIN_DIR)/asvec-macos-amd64 $(BIN_DIR)/macos-pkg/asvec/
	cp -a $(BIN_DIR)/asvec-macos-arm64 $(BIN_DIR)/macos-pkg/asvec/
	sed "s/ASVECVERSIONHERE/${ver}/g" $(BIN_DIR)/macos-pkg/AsVec-template.pkgproj > $(BIN_DIR)/macos-pkg/AsVec.pkgproj
	bash -ce "cd $(BIN_DIR)/macos-pkg && /usr/local/bin/packagesbuild --project AsVec.pkgproj"
	mv $(BIN_DIR)/macos-pkg/build/AsVec.pkg $(BIN_DIR)/asvec-macos-${ver}-unsigned.pkg

.PHONY: macos-pkg-sign
macos-pkg-sign:
	productsign --timestamp --sign ${INSTALLSIGNER} $(BIN_DIR)/asvec-macos-${ver}-unsigned.pkg $(BIN_DIR)/packages/asvec-macos-${ver}.pkg

.PHONY: macos-pkg-notarize
macos-pkg-notarize:
	rm -f notarize_result_pkg
	xcrun notarytool submit --apple-id ${APPLEID} --password ${APPLEPW} --team-id ${TEAMID} -f json --wait --timeout 10m $(BIN_DIR)/packages/asvec-macos-${ver}.pkg > notarize_result_pkg
	if [ "$$(cat notarize_result_pkg |jq -r .status)" != "Accepted" ] ;\
	then \
		echo "PKG FAILED TO NOTARIZE" ;\
		cat notarize_result_pkg ;\
		exit 1 ;\
	else \
		echo "PKG NOTARIZE SUCCESS" ;\
	fi

### make cleanall && make build-prerelease && make pkg-linux && make pkg-windows-zip && make macos-build-all && make macos-notarize-all
### make cleanall && make build-official && make pkg-linux && make pkg-windows-zip && make macos-build-all && make macos-notarize-all

.PHONY: test
test: integration unit

.PHONY: test-large
test-large: integration-large unit

.PHONY: integration
integration:
	mkdir -p $(COV_INTEGRATION_DIR) || true
	COVERAGE_DIR=$(COV_INTEGRATION_DIR) go test -tags=integration -timeout 30m 

.PHONY: integration-large
integration-large:
	mkdir -p $(COV_INTEGRATION_DIR) || true
	COVERAGE_DIR=$(COV_INTEGRATION_DIR) go test -tags=integration_large -timeout 30m 

.PHONY: unit
unit:
	mkdir -p $(COV_UNIT_DIR) || true
	go test -tags=unit -cover ./... -args -test.gocoverdir=$(COV_UNIT_DIR)

.PHONY: coverage
coverage: test-large
	go tool covdata textfmt -i="$(COV_INTEGRATION_DIR),$(COV_UNIT_DIR)" -o=$(COVERAGE_DIR)/total.cov
	go tool cover -func=$(COVERAGE_DIR)/total.cov

PHONY: view-coverage
view-coverage: $(COVERAGE_DIR)/total.cov
	go tool cover -html=$(COVERAGE_DIR)/total.cov
