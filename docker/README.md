# Zoraxy Docker

[![Repo](https://img.shields.io/badge/Docker-Repo-007EC6?labelColor-555555&color-007EC6&logo=docker&logoColor=fff&style=flat-square)](https://hub.docker.com/r/zoraxydocker/zoraxy)
[![Version](https://img.shields.io/docker/v/zoraxydocker/zoraxy/latest?labelColor-555555&color-007EC6&style=flat-square)](https://hub.docker.com/r/zoraxydocker/zoraxy)
[![Size](https://img.shields.io/docker/image-size/zoraxydocker/zoraxy/latest?sort=semver&labelColor-555555&color-007EC6&style=flat-square)](https://hub.docker.com/r/zoraxydocker/zoraxy)
[![Pulls](https://img.shields.io/docker/pulls/zoraxydocker/zoraxy?labelColor-555555&color-007EC6&style=flat-square)](https://hub.docker.com/r/zoraxydocker/zoraxy)

## Usage

If you are attempting to access your service from outside your network, make sure to forward ports 80 and 443 to the Zoraxy host to allow web traffic. If you know how to do this, great! If not, find the manufacturer of your router and search on how to do that. There are too many to be listed here. Read more about it from [whatismyip](https://www.whatismyip.com/port-forwarding/).

In the examples below, make sure to update `/path/to/zoraxy/config/`. If a path is not provided, a Docker volume will be created at the location but it is recommended to store the data at a defined host location or a named Docker volume.

Once setup, access the webui at `http://<host-ip>:8000` to configure Zoraxy. Change the port in the URL if you changed the management port.

### Docker Run

```
docker run -d \
  --name zoraxy \
  --restart unless-stopped \
  -p 80:80 \
  -p 443:443 \
  -p 8000:8000 \
  -v /path/to/zoraxy/config/:/opt/zoraxy/config/ \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v /etc/localtime:/etc/localtime \
  -e FASTGEOIP="true" \
  zoraxydocker/zoraxy:latest
```

### Docker Compose

```yml
services:
  zoraxy:
    image: zoraxydocker/zoraxy:latest
    container_name: zoraxy
    restart: unless-stopped
    ports:
      - 80:80
      - 443:443
      - 8000:8000
    volumes:
      - /path/to/zoraxy/config/:/opt/zoraxy/config/
      - /var/run/docker.sock:/var/run/docker.sock
      - /etc/localtime:/etc/localtime
    environment:
      FASTGEOIP: "true"
```

### Ports

| Port | Details |
|:-|:-|
| `80` | HTTP traffic. |
| `443` | HTTPS traffic. |
| `8000` | Management interface. Can be changed with the `PORT` env. |

### Volumes

| Volume | Details |
|:-|:-|
| `/opt/zoraxy/config/` | Zoraxy configuration. |
| `/var/run/docker.sock` | Docker socket. Used for additional functionality with Zoraxy. |
| `/etc/localtime` | Localtime. Set to ensure the host and container are synchronized. |

### Environment

Variables are the same as those in [Start Parameters](https://github.com/tobychui/zoraxy?tab=readme-ov-file#start-paramters).

| Variable | Default | Details |
|:-|:-|:-|
| `AUTORENEW` | `86400` (Integer) | ACME auto TLS/SSL certificate renew check interval. |
| `CFGUPGRADE` | `true` (Boolean) | Enable auto config upgrade if breaking change is detected. |
| `DB` | `auto` (String) | Database backend to use (leveldb, boltdb, auto) Note that fsdb will be used on unsupported platforms like RISCV (default "auto"). |
| `DOCKER` | `true` (Boolean) | Run Zoraxy in docker compatibility mode. |
| `EARLYRENEW` | `30` (Integer) | Number of days to early renew a soon expiring certificate. |
| `FASTGEOIP` | `false`  (Boolean) | Enable high speed geoip lookup, require 1GB extra memory (Not recommend for low end devices). |
| `MDNS` | `true` (Boolean) | Enable mDNS scanner and transponder. |
| `MDNSNAME` | `''` (String) | mDNS name, leave empty to use default (zoraxy_{node-uuid}.local). |
| `NOAUTH` | `false` (Boolean) | Disable authentication for management interface. |
| `PORT` | `8000` (Integer) | Management web interface listening port |
| `SSHLB` | `false` (Boolean) | Allow loopback web ssh connection (DANGER). |
| `UPDATE_GEOIP` | `false` (Boolean) | Download the latest GeoIP data and exit. |
| `VERSION` | `false` (Boolean) | Show version of this server. |
| `WEBFM` | `true` (Boolean) | Enable web file manager for static web server root folder. |
| `WEBROOT` | `./www` (String) | Static web server root folder. Only allow change in start parameters. |
| `ZEROTIER` | `false` (Boolean) | Enable ZeroTier functionality for GAN. |
| `ZTAUTH` | `""` (String) | ZeroTier authtoken for the local node. |
| `ZTPORT` | `9993` (Integer) | ZeroTier controller API port. |

> [!IMPORTANT]
> Contrary to the Zoraxy README, Docker usage of the port flag should NOT include the colon. Ex: `-e PORT="8000"` for Docker run and `PORT: "8000"` for Docker compose.

### Building

To build the Docker image:
  - Check out the repository/branch.
  - Copy the Zoraxy `src/` directory into the `docker/` (here) directory.
  - Run the build command with `docker build -t zoraxy_build .`
  - You can now use the image `zoraxy_build`
    - If you wish to change the image name, then modify`zoraxy_build` in the previous step and then build again.

