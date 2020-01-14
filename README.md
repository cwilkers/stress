[![Docker Repository on Quay](https://quay.io/repository/cwilkers/stress-pin/status "Docker Repository on Quay")](https://quay.io/repository/cwilkers/stress-pin)

# Lightweight compute resource stress utility

Forked from https://github.com/vishh/stress

Added cpu pinning code to ensure threads spread out over available CPUs.

Added multi-stage Dockerfile setup to build Go code then build bare bones stress container.
