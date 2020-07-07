### Echo64
In one terminal, start the 9p server:
```
% go run examples/echo64/main.go
```
A docker image is also available
```
% docker pull danielmorandini/echo64
% docker run -p 564:564 danielmorandini/echo64
```

In another one, use it:
```
% 9 mount localhost:564 mnt # This is how you mount a 9p fs using 9port
% echo "hello" > mnt/ctl
% cat mnt/retv
{"original":"hello\n","base64":"aGVsbG8K"}
```
