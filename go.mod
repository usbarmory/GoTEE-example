module github.com/usbarmory/GoTEE-example

go 1.23.4

require (
	github.com/usbarmory/GoTEE v0.0.0-20241204121110-7ea78738897b
	github.com/usbarmory/armory-boot v0.0.0-20241007114806-656160cd9b23
	github.com/usbarmory/imx-usbnet v0.0.0-20240909221106-d242d2c2d20b
	github.com/usbarmory/tamago v0.0.0-20241204113720-e648ef3a4633
	golang.org/x/crypto v0.29.0
	golang.org/x/term v0.26.0
)

require (
	github.com/dsoprea/go-ext4 v0.0.0-20190528173430-c13b09fc0ff8 // indirect
	github.com/dsoprea/go-logging v0.0.0-20200710184922-b02d349568dd // indirect
	github.com/go-errors/errors v1.0.2 // indirect
	github.com/google/btree v1.1.2 // indirect
	github.com/pierrec/lz4/v4 v4.1.14 // indirect
	github.com/u-root/u-root v0.14.0 // indirect
	github.com/u-root/uio v0.0.0-20240209044354-b3d14b93376a // indirect
	golang.org/x/exp v0.0.0-20240222234643-814bf88cf225 // indirect
	golang.org/x/net v0.21.0 // indirect
	golang.org/x/sync v0.9.0 // indirect
	golang.org/x/sys v0.27.0 // indirect
	golang.org/x/time v0.5.0 // indirect
	gvisor.dev/gvisor v0.0.0-20240909175600-91fb8ad18db5 // indirect
)

replace github.com/usbarmory/GoTEE => /mnt/git/public/GoTEE
replace github.com/usbarmory/tamago => /mnt/git/public/tamago
