# Backend  Branch

## Set up development environment
1. Make the setup script exec: chmod +x setup.sh
2. run setup: `make setup`
3. install reqs: `make install-deps`

## Build and Run
1. generate libbpf go code (from c ebpf code): `make generate`
2. build the target program: `make build`
3. run the program `make run IFACE=<your network interface>`. Run `ip a` to see your interfaces. Will be something like enp1s0 or eth0