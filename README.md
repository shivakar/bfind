# bfind
A multithreaded find alternative using breadth-first search

## Installation

```
go get -u -v github.com/shivakar/bfind
```

## Usage

```
Usage of bfind:
  -1    end after finding the first matching entry
  -name string
        find files with names matching regex (case-senstive)
  -num-threads int
        number of threads to process (default runtime.NumCPU())
  -type string
        filter on file type attribute
```

## Examples

### Comparison to find

```
find /etc/ -name *grub*
```
results in:

```
/etc/grub2.cfg
/etc/prelink.conf.d/grub2.conf
/etc/grub.d
/etc/sysconfig/grub
/etc/default/grub
```

```
bfind -name .*grub.* /etc/
```
results in:

```
/etc/grub.d
/etc/grub2.cfg
/etc/default/grub
/etc/prelink.conf.d/grub2.conf
/etc/sysconfig/grub
```

### Find first match

```
bfind -1 -name .*grub.* /etc/

/etc/grub.d
```

### Find only regular files

```
bfind -1 -type f -name .*grub.* /etc/

/etc/default/grub
/etc/prelink.conf.d/grub2.conf
/etc/sysconfig/grub
```

### Find only directories

```
bfind -1 -type d -name .*grub.* /etc/

/etc/grub.d
```
