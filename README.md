# floc
create or write file or device/file images

work in progress

WARNING: if used as intended, this tool works on raw devices and files. This means: if used incorrectly it will result in data loss or render your device unusable (aka brick it).

## Installation / Build from source
```
go get github.com/w33zl3p00tch/floc
```

## Usage
First: Windows users beware: this currently does not work on raw devices under Windows. On UNIX-like systems there are no problems, AFAICT, as long as you have root privileges.

Example: flash a bootable Linux ISO to USB, using a buffer size of 1 megabyte and comparing sha256 hashes of input and output afterwards:
```
# floc -source ~/Downloads/some_linux.iso -target /dev/sdX
```
To skip the potentially lengthy checksum comparison, simply add the ```-nocheck``` flag.

Different buffer sizes than the default of 1 megabyte can be set by adding the ```-buffersize``` flag which currently only supports a value in kilobytes. This will become more flexible.

If you are using macOS, you'll probly run into an "Operation not permitted" error if your target device contains a readable filesystem and is currently mounted. To fix this, don't eject the device in Finder but unmount it on the commandline like so:
```
sudo diskutil unmount /dev/diskX
```
