# http://github.com/usbarmory/GoTEE-example
#
# Copyright (c) WithSecure Corporation
# https://foundry.withsecure.com
#
# Use of this source code is governed by the license
# that can be found in the LICENSE file.

BUILD_USER ?= $(shell whoami)
BUILD_HOST ?= $(shell hostname)
BUILD_DATE ?= $(shell /bin/date -u "+%Y-%m-%d %H:%M:%S")
BUILD_TAGS = linkramsize,linkramstart
BUILD = ${BUILD_USER}@${BUILD_HOST} on ${BUILD_DATE}
REV = $(shell git rev-parse --short HEAD 2> /dev/null)

SHELL = /bin/bash

APP := ""
TARGET ?= "usbarmory"
TEXT_START := 0x80010000 # ramStart (defined in mem.go under relevant tamago/soc package) + 0x10000

GOENV := GO_EXTLINK_ENABLED=0 CGO_ENABLED=0 GOOS=tamago GOARM=7 GOARCH=arm
ENTRY_POINT := _rt0_arm_tamago
QEMU ?= qemu-system-arm -machine mcimx6ul-evk -cpu cortex-a7 -m 512M \
        -nographic -monitor none -serial null -serial stdio -net none \
        -semihosting

GOFLAGS = -tags ${TARGET},${BUILD_TAGS} -trimpath -ldflags "-T ${TEXT_START} -E ${ENTRY_POINT} -R 0x1000 -X 'main.Build=${BUILD}' -X 'main.Revision=${REV}'"
RUSTFLAGS = -C linker=arm-none-eabi-ld -C link-args="--Ttext=$(TEXT_START)" --target armv7a-none-eabi

.PHONY: clean qemu qemu-gdb trusted_applet_rust

#### primary targets ####

elf: $(APP).elf

trusted_os: APP=trusted_os_$(TARGET)
trusted_os: DIR=$(CURDIR)/trusted_os_$(TARGET)
trusted_os: TEXT_START=0x98010000
trusted_os: imx

trusted_os_signed: APP=trusted_os_$(TARGET)
trusted_os_signed: DIR=$(CURDIR)/trusted_os_$(TARGET)
trusted_os_signed: TEXT_START=0x98010000
trusted_os_signed: imx_signed

trusted_applet_go: APP=trusted_applet
trusted_applet_go: DIR=$(CURDIR)/trusted_applet_go
trusted_applet_go: TEXT_START=0x9c010000
trusted_applet_go: elf
	mkdir -p $(CURDIR)/trusted_os_$(TARGET)/assets
	cp $(CURDIR)/bin/trusted_applet.elf $(CURDIR)/trusted_os_$(TARGET)/assets

trusted_applet_rust: TEXT_START=0x9c010000
trusted_applet_rust:
	cd $(CURDIR)/trusted_applet_rust && rustc ${RUSTFLAGS} -o $(CURDIR)/bin/trusted_applet.elf main.rs
	mkdir -p $(CURDIR)/trusted_os_$(TARGET)/assets
	cp $(CURDIR)/bin/trusted_applet.elf $(CURDIR)/trusted_os_$(TARGET)/assets

nonsecure_os_go: APP=nonsecure_os_go
nonsecure_os_go: DIR=$(CURDIR)/nonsecure_os_go
nonsecure_os_go: TEXT_START=0x80010000
nonsecure_os_go: elf
	mkdir -p $(CURDIR)/trusted_os_$(TARGET)/assets
	cp $(CURDIR)/bin/nonsecure_os_go.elf $(CURDIR)/trusted_os_$(TARGET)/assets

nonsecure_os_linux: APP=nonsecure_os_linux
nonsecure_os_linux: DIR=$(CURDIR)/nonsecure_os_linux
nonsecure_os_linux: TEXT_START=0x80010000
nonsecure_os_linux: todo

#### ARM targets ####

imx: $(APP).imx

imx_signed: $(APP)-signed.imx

check_hab_keys:
	@if [ "${HAB_KEYS}" == "" ]; then \
		echo 'You need to set the HAB_KEYS variable to the path of secure boot keys'; \
		echo 'See https://github.com/usbarmory/usbarmory/wiki/Secure-boot-(Mk-II)'; \
		exit 1; \
	fi

$(APP).bin: CROSS_COMPILE=arm-none-eabi-
$(APP).bin: $(APP).elf
	$(CROSS_COMPILE)objcopy -j .text -j .rodata -j .shstrtab -j .typelink \
	    -j .itablink -j .gopclntab -j .go.buildinfo -j .noptrdata -j .data \
	    -j .bss --set-section-flags .bss=alloc,load,contents \
	    -j .noptrbss --set-section-flags .noptrbss=alloc,load,contents \
	    $(CURDIR)/bin/$(APP).elf -O binary $(CURDIR)/bin/$(APP).bin

$(APP).imx: $(APP).bin $(APP).dcd
	@if [ "$(APP)" == "trusted_os_usbarmory" ]; then \
		echo "## disabling TZASC bypass in DCD for pre-DDR initialization ##"; \
		chmod 644 $(CURDIR)/bin/$(APP).dcd; \
		echo "DATA 4 0x020e4024 0x00000001  # TZASC_BYPASS" >> $(CURDIR)/bin/$(APP).dcd; \
	fi
	mkimage -n $(CURDIR)/bin/$(APP).dcd -T imximage -e $(TEXT_START) -d $(CURDIR)/bin/$(APP).bin $(CURDIR)/bin/$(APP).imx
	# Copy entry point from ELF file
	dd if=$(CURDIR)/bin/$(APP).elf of=$(CURDIR)/bin/$(APP).imx bs=1 count=4 skip=24 seek=4 conv=notrunc

$(APP).dcd: check_tamago
$(APP).dcd: GOMODCACHE=$(shell ${TAMAGO} env GOMODCACHE)
$(APP).dcd: TAMAGO_PKG=$(shell grep "github.com/usbarmory/tamago v" go.mod | awk '{print $$1"@"$$2}')
$(APP).dcd: dcd

#### utilities ####

check_tamago:
	@if [ "${TAMAGO}" == "" ] || [ ! -f "${TAMAGO}" ]; then \
		echo 'You need to set the TAMAGO variable to a compiled version of https://github.com/usbarmory/tamago-go'; \
		exit 1; \
	fi

dcd:
	cp -f $(GOMODCACHE)/$(TAMAGO_PKG)/board/usbarmory/mk2/imximage.cfg $(CURDIR)/bin/$(APP).dcd; \

clean:
	@rm -fr $(CURDIR)/bin/* $(CURDIR)/trusted_os_*/assets/*

qemu:
	$(QEMU) -kernel $(CURDIR)/bin/trusted_os_$(TARGET).elf

qemu-gdb:
	$(QEMU) -kernel $(CURDIR)/bin/trusted_os_$(TARGET).elf -S -s

#### application target ####

$(APP).elf: check_tamago
	cd $(DIR) && $(GOENV) $(TAMAGO) build -tags ${BUILD_TAGS} $(GOFLAGS) -o $(CURDIR)/bin/$(APP).elf

#### HAB secure boot ####

$(APP)-signed.imx: check_hab_keys $(APP).imx
	${TAMAGO} install github.com/usbarmory/crucible/cmd/habtool
	$(shell ${TAMAGO} env GOPATH)/bin/habtool \
		-A ${HAB_KEYS}/CSF_1_key.pem \
		-a ${HAB_KEYS}/CSF_1_crt.pem \
		-B ${HAB_KEYS}/IMG_1_key.pem \
		-b ${HAB_KEYS}/IMG_1_crt.pem \
		-t ${HAB_KEYS}/SRK_1_2_3_4_table.bin \
		-x 1 \
		-s \
		-i $(CURDIR)/bin/$(APP).imx \
		-o $(CURDIR)/bin/$(APP).csf && \
	cat $(CURDIR)/bin/$(APP).imx $(CURDIR)/bin/$(APP).csf > $(CURDIR)/bin/$(APP)-signed.imx
