# http://github.com/f-secure-foundry/GoTEE-example
#
# Copyright (c) F-Secure Corporation
# https://foundry.f-secure.com
#
# Use of this source code is governed by the license
# that can be found in the LICENSE file.

BUILD_USER = $(shell whoami)
BUILD_HOST = $(shell hostname)
BUILD_DATE = $(shell /bin/date -u "+%Y-%m-%d %H:%M:%S")
BUILD_TAGS = linkramsize,linkramstart
BUILD = ${BUILD_USER}@${BUILD_HOST} on ${BUILD_DATE}
REV = $(shell git rev-parse --short HEAD 2> /dev/null)

SHELL = /bin/bash

CROSS_COMPILE = arm-none-eabi-
APP := ""
GOENV := GO_EXTLINK_ENABLED=0 CGO_ENABLED=0 GOOS=tamago GOARM=7 GOARCH=arm
GOFLAGS = -ldflags "-T $(TEXT_START) -E _rt0_arm_tamago -R 0x1000 -X 'main.Build=${BUILD}' -X 'main.Revision=${REV}'"
QEMU ?= qemu-system-arm -machine mcimx6ul-evk -cpu cortex-a7 -m 512M \
        -nographic -monitor none -serial null -serial stdio -net none \
        -semihosting

.PHONY: clean qemu qemu-gdb

#### primary targets ####

trusted_os: APP=trusted_os
trusted_os: DIR=$(CURDIR)/trusted_os
trusted_os: TEXT_START=0x80010000
trusted_os: imx

trusted_os_signed: APP=trusted_os
trusted_os_signed: DIR=$(CURDIR)/trusted_os
trusted_os_signed: TEXT_START=0x80010000
trusted_os_signed: imx_signed

trusted_applet_go: APP=trusted_applet
trusted_applet_go: DIR=$(CURDIR)/trusted_applet_go
trusted_applet_go: TEXT_START=0x82010000
trusted_applet_go: imx
	cp trusted_applet.elf $(CURDIR)/trusted_os/

nonsecure_os_go: APP=nonsecure_os_go
nonsecure_os_go: DIR=$(CURDIR)/nonsecure_os_go
nonsecure_os_go: TEXT_START=0x84010000
nonsecure_os_go: imx
	cp nonsecure_os_go.elf $(CURDIR)/trusted_os/

nonsecure_os_linux: APP=nonsecure_os_linux
nonsecure_os_linux: DIR=$(CURDIR)/nonsecure_os_linux
nonsecure_os_linux: TEXT_START=0x84010000
nonsecure_os_linux: todo

imx: $(APP).imx

imx_signed: $(APP)-signed.imx

elf: $(APP).elf

#### utilities ####

check_tamago:
	@if [ "${TAMAGO}" == "" ] || [ ! -f "${TAMAGO}" ]; then \
		echo 'You need to set the TAMAGO variable to a compiled version of https://github.com/f-secure-foundry/tamago-go'; \
		exit 1; \
	fi

check_hab_keys:
	@if [ "${HAB_KEYS}" == "" ]; then \
		echo 'You need to set the HAB_KEYS variable to the path of secure boot keys'; \
		echo 'See https://github.com/f-secure-foundry/usbarmory/wiki/Secure-boot-(Mk-II)'; \
		exit 1; \
	fi

dcd:
	cp -f $(GOMODCACHE)/$(TAMAGO_PKG)/board/f-secure/usbarmory/mark-two/imximage.cfg $(APP).dcd; \

clean:
	@rm -fr *.elf *.bin *.imx *-signed.imx *.csf *.dcd trusted_os/*.elf

qemu:
	$(QEMU) -kernel trusted_os.elf

qemu-gdb:
	$(QEMU) -kernel trusted_os.elf -S -s

#### dependencies ####

$(APP).elf: check_tamago
	cd $(DIR) && $(GOENV) $(TAMAGO) build -tags ${BUILD_TAGS} $(GOFLAGS) -o $(CURDIR)/$(APP).elf

$(APP).dcd: check_tamago
$(APP).dcd: GOMODCACHE=$(shell ${TAMAGO} env GOMODCACHE)
$(APP).dcd: TAMAGO_PKG=$(shell grep "github.com/f-secure-foundry/tamago v" go.mod | awk '{print $$1"@"$$2}')
$(APP).dcd: dcd

$(APP).bin: $(APP).elf
	$(CROSS_COMPILE)objcopy -j .text -j .rodata -j .shstrtab -j .typelink \
	    -j .itablink -j .gopclntab -j .go.buildinfo -j .noptrdata -j .data \
	    -j .bss --set-section-flags .bss=alloc,load,contents \
	    -j .noptrbss --set-section-flags .noptrbss=alloc,load,contents \
	    $(APP).elf -O binary $(APP).bin

$(APP).imx: $(APP).bin $(APP).dcd
	@if [ "$(APP)" == "trusted_os" ]; then \
		echo "## disabling TZASC bypass in DCD for pre-DDR use initialization ##"; \
		chmod 644 $(APP).dcd; \
		echo "DATA 4 0x020e4024 0x00000001  # TZASC_BYPASS" >> $(APP).dcd; \
	fi
	mkimage -n $(APP).dcd -T imximage -e $(TEXT_START) -d $(APP).bin $(APP).imx
	# Copy entry point from ELF file
	dd if=$(APP).elf of=$(APP).imx bs=1 count=4 skip=24 seek=4 conv=notrunc

#### secure boot ####

$(APP)-signed.imx: check_hab_keys $(APP).imx
	${TAMAGO} install github.com/f-secure-foundry/crucible/cmd/habtool
	$(shell ${TAMAGO} env GOPATH)/bin/habtool \
		-A ${HAB_KEYS}/CSF_1_key.pem \
		-a ${HAB_KEYS}/CSF_1_crt.pem \
		-B ${HAB_KEYS}/IMG_1_key.pem \
		-b ${HAB_KEYS}/IMG_1_crt.pem \
		-t ${HAB_KEYS}/SRK_1_2_3_4_table.bin \
		-x 1 \
		-s \
		-i $(APP).imx \
		-o $(APP).csf && \
	cat $(APP).imx $(APP).csf > $(APP)-signed.imx
