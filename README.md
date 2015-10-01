# coatl
running, conducting, directing

The parts here are:

coatlctl - Command-Line Interface to set up and enrol services
listen - simple listener that just prints out when something happens.

These programs rely on a running `etcd` listening on port 4001.
To run `etcd`:

    docker run --name etcd -d -p 4001:4001 quay.io/coreos/etcd \
      -advertise-client-urls http://0.0.0.0:4001 \
      -listen-client-urls http://0.0.0.0:4001
