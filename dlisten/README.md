# dlisten
Listens to Docker and enrols/unenrols containers matching services registered in etcd

Configuration environment variables:

- `DOCKER_HOST` - default `unix:///var/run/docker.sock`
- `ETCD_ADDRESS` - default `http://127.0.0.1:4001`

