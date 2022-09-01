This folder contains resources which are used as part of the acceptance tests.

```shell
ssh-keygen -q -t ed25519 -N '' -f 'ed25519-primary' -C 'ed25519-primary'
ssh-keygen -q -t ed25519 -N '' -f 'ed25519-secondary' -C 'ed25519-secondary'

ssh-keygen -q -t ecdsa -N '' -f 'ecdsa-primary' -C 'ecdsa-primary'
ssh-keygen -q -t ecdsa -N '' -f 'ecdsa-secondary' -C 'ecdsa-secondary'
```
