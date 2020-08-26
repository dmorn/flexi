<!--
SPDX-FileCopyrightText: 2020 jecoz

SPDX-License-Identifier: BSD-3-Clause
-->

# flexi
![Docker Cloud Build Status](https://img.shields.io/docker/cloud/build/jecoz/flexi?label=docker%20build%20-%20flexi)
![Docker Cloud Build Status](https://img.shields.io/docker/cloud/build/jecoz/echo64?label=docker%20build%20-%20echo64)
[![GoDoc](https://godoc.org/github.com/jecoz/flexi?status.svg)](https://godoc.org/github.com/jecoz/flexi)
[![asciicast](https://asciinema.org/a/345847.svg)](https://asciinema.org/a/345847)

### Warning: work in progress, do not use this library.

### docker session
on one terminal:
```
% docker run -p 564:564 --env-file docker.env --privileged jecoz/flexi
```
or, if you want to persist mounts across sessions (preferred way), create a Docker volume and issue
```
% docker run -p 564:564 --env-file docker.env --privileged --mount src=<desired volume name>,dst=/mnt jecoz/flexi
```

on another terminal:
```
% 9 mount localhost:9pfs mnt
% cat mnt/clone
0
% cat testdata/input.1.json > mnt/0/spawn
% cat mnt/0/state
0.14285714285714285,starting mnt/0 mount process
0.2857142857142857,spawning remote process
0.42857142857142855,remote process spawned @ 3.249.96.176:564
0.5714285714285714,remote process mounted @ mnt/0
0.7142857142857143,storing spawn information at mnt/0
0.8571428571428571,remote process info encoded & saved
1,done!
% echo brother is your turn > mnt/0/mirror/ctl
% cat mnt/0/mirror/state
0.3333333333333333,buffering input payload
0.6666666666666666,base64 encoding 21 bytes
1,done!
 % cat mnt/0/mirror/retv
{"original":"brother is your turn\n","base64":"YnJvdGhlciBpcyB5b3VyIHR1cm4K"}
```

### Notes about deploying to AWS
- flexi needs to be hosted in an environment that allows it to "mount", hence **not** Fargate but rather ECS with priviledged flag enabled (hang on we're working on it. See [issue #10](https://github.com/jecoz/flexi/issues/10))
