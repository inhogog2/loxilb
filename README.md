[![Build Status](https://app.travis-ci.com/loxilb-io/loxilb.svg?branch=main)](https://app.travis-ci.com/loxilb-io/loxilb)

## What is loxilb

loxilb is a cloud-native networking/security stack built from grounds up using eBPF at its core. loxilb aims to provide the following :

- Service type external load-balancer for kubernetes (hence the name loxilb)
- L4/NAT stateful loadbalancer 
   * High-availability support
   * K8s CCM compliance
-  Optimized SRv6 implementation in eBPF 
-  Make GTP tunnels first class citizens of the Linux world 
   * Support for QFI and other extension headers
-  eBPF based kernel forwarding (GPLv2 license)
   * Complete kernel bypass with home-grown stack for advanced features like [Conntrack](https://thermalcircle.de/doku.php?id=blog:linux:connection_tracking_1_modules_and_hooks), QoS etc
   * Highly scalable with low-latency & high througput 
   * Mainly uses TC-eBPF hooks
-  goLang based control plane components (Apache license)
-  Seamless integration with goBGP based routing stack
-  Easily cuztomizable to run in DPU environments
   * goLang based easy to use APIs/Interfaces


## How to build/run

1. Install package dependencies 

```
sudo apt install clang llvm libelf-dev gcc-multilib libpcap-dev
sudo apt install linux-tools-$(uname -r)
sudo apt install elfutils dwarves
```

*loxilb also requires a special version of [iproute2](https://github.com/shemminger/iproute2) tool for its operation. The customized repository can be found [here](https://github.com/loxilb-io/iproute2). Detailed build instructions can be found [here](https://github.com/loxilb-io/iproute2/blob/main/README.loxilb).*

2. Build libbpf locally

```
#  cd libbpf/src/
#  mkdir build
#  DESTDIR=build make install
```

3. Make loxilb

```
make
```

4. Run  loxilb

```
sudo go run .
```

  or 

```
sudo ./loxilb 
```


We encourage loxilb users to follow various guides in loxilb docs [repository](https://github.com/loxilb-io/loxilbdocs)
