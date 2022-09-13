
```shell
dd if=/dev/zero of=zeroFile bs=1K count=1
dd if=/dev/urandom of=randomFile bs=1M count=1024  

dd if=/dev/urandom of=/tmp/20M.file bs=1M count=20
```