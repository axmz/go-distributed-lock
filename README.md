# Distributed lock
The idea is that multiple worker jobs (distrilock) will increase the counter in the Redis intermitently.
Once all of them report, a quick check is made.
If the counter matches the exepcted results, then there was no race condition.

Several demos of distributed lock on different git branches.
- SETNX + single Redis node + docker-compose: Requirement: docker-compose
- Redislock + Redis cluster and k8s. Requirements: helm

Tech used: grpc, k8s, helm, docker

# Run
Make changes to values.yml and then run.
```
make run
```

