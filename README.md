# VKPM

CLI tool that allows time report and get some stats without getting into
the VKPM UI.

## Install

You can either download a binary from the latest release, or get it with Go:
```shell
go install github.com/kudrykv/go-vkpm/cmd/vkpm@latest
```

## Usage

First you need to config the domain for communication and log in:
```shell
$ vkpm config --domain vkpm-domain
$ vkpm login
```

Now, reporting time is easy:
```shell
# report 1h, 2h30m and 4h30m
# these would go to 09:00-10:00, 10:00-12:30, and 12:30-17:00
vkpm report -p egginc -s 1h -a analysis -m 'dealing with email'
vkpm report -p k4s -s 2h30m -S 80 -m 'coding net req. Im close with completing!'
vkpm report -p egginc -s 4h30m -a management -m 'working with client, feedback'

# it is also possible to specify time range manually
vkpm report -p egginc -f 12:00 -t 13:00 -m 'developing stuff'

# also possible to report previous time
vkpm report -F 05-13 -p egginc -s 8h -m 'did stuff yesterday, forgot to report'
```
