# DLGrab (docker layer grab)

## Requirements

You need to have docker running on your local machine. Remote daemons won't
work, though it wouldn't be hard to add them.

## Caveats

DLGrab shims the api endpoints docker touches when pushing an image. If
Docker starts expecting different responses from these endpoints, DLGrab may
stop working.

That said, it's probably more robust than trying to export directly from the
filesystem.

## Build dlgrab

    $ git clone https://github.com/aidanhs/dlgrab.git
    $ cd dlgrab
    $ make
    $ ./dlgrab

## Usage

    $ ./dlgrab
    Usage for ./dlgrab [flags...] LAYER
      LAYER: layer id to export, or image name to export top layer of
      --clean=false              Remove the temporary tag after use
                                   WARNING: can trigger layer deletion if run on a layer with no children or other references
      --debug=false              Set log level to debug
      -h, --help=false           Print this help text
      -o, --outdir="."           Directory to write layer to
      -p=0                       Port to use, defaults to a random unallocated port
      --registry-format=false    Output in the format a registry would use, rather than for an image export
    
      The DOCKER_HOST environment variable overrides the default location to find the docker daemon

Exporting a single layer:

    $ docker history ubuntu:14.04
    IMAGE               CREATED             CREATED BY                                      SIZE
    5ba9dab47459        2 weeks ago         /bin/sh -c #(nop) CMD [/bin/bash]               0 B
    51a9c7c1f8bb        2 weeks ago         /bin/sh -c sed -i 's/^#\s*\(deb.*universe\)$/   1.895 kB
    5f92234dcf1e        2 weeks ago         /bin/sh -c echo '#!/bin/sh' > /usr/sbin/polic   194.5 kB
    27d47432a69b        2 weeks ago         /bin/sh -c #(nop) ADD file:62400a49cced0d7521   188.1 MB
    511136ea3c5a        20 months ago                                                       0 B
    $ ./dlgrab --clean 51a9c7c1f8bb # this layer is referenced by child images, so --clean is fine
    Full layer id found: 51a9c7c1f8bb2fa19bcd09789a34e63f35abb80044bc10196e304f6634cc582c
    Layer folder will be dumped into .
    Writing file: json
    Writing file: layer.tar
    Export complete
    $ ls 51a9c7c1f8bb2fa19bcd09789a34e63f35abb80044bc10196e304f6634cc582c
    json  layer.tar  VERSION
    $ tar tf 51a9c7c1f8bb2fa19bcd09789a34e63f35abb80044bc10196e304f6634cc582c/layer.tar 
    etc/
    etc/apt/
    etc/apt/sources.list

Loading an image built with individually exported layers:

    $ cd $(mktemp -d)
    /tmp/tmp.R433Vilvoa $ docker history centos:6.6
    IMAGE               CREATED             CREATED BY                                      SIZE
    77e369743e24        9 days ago          /bin/sh -c #(nop) ADD file:f10753ca9b3793921c   202.6 MB
    5b12ef8fd570        4 months ago        /bin/sh -c #(nop) MAINTAINER The CentOS Proje   0 B
    511136ea3c5a        20 months ago                                                       0 B
    /tmp/tmp.R433Vilvoa $ dlgrab --clean 77e369743e24
    Full layer id found: 77e369743e242d5b13f6426b128b1609f8aa1091b935e50b0de81c8bc3e1510d
    Layer folder will be dumped into .
    Writing file: json
    Writing file: layer.tar
    Export complete
    /tmp/tmp.R433Vilvoa $ dlgrab --clean 5b12ef8fd570
    Full layer id found: 5b12ef8fd57065237a6833039acc0e7f68e363c15d8abb5cacce7143a1f7de8a
    Layer folder will be dumped into .
    Writing file: json
    Writing file: layer.tar
    Export complete
    /tmp/tmp.R433Vilvoa $ dlgrab --clean 511136ea3c5a
    Full layer id found: 511136ea3c5a64f264b78b5433614aec563103b4d4702f3ba7d4d2698e22c158
    Layer folder will be dumped into .
    Writing file: json
    Writing file: layer.tar
    Export complete
    /tmp/tmp.R433Vilvoa $ docker rmi centos:6.6
    Untagged: centos:6.6
    Deleted: 77e369743e242d5b13f6426b128b1609f8aa1091b935e50b0de81c8bc3e1510d
    Deleted: 5b12ef8fd57065237a6833039acc0e7f68e363c15d8abb5cacce7143a1f7de8a
    /tmp/tmp.R433Vilvoa $ tar cf img.tar *
    /tmp/tmp.R433Vilvoa $ tar tf img.tar
    511136ea3c5a64f264b78b5433614aec563103b4d4702f3ba7d4d2698e22c158/
    511136ea3c5a64f264b78b5433614aec563103b4d4702f3ba7d4d2698e22c158/json
    511136ea3c5a64f264b78b5433614aec563103b4d4702f3ba7d4d2698e22c158/VERSION
    511136ea3c5a64f264b78b5433614aec563103b4d4702f3ba7d4d2698e22c158/layer.tar
    5b12ef8fd57065237a6833039acc0e7f68e363c15d8abb5cacce7143a1f7de8a/
    5b12ef8fd57065237a6833039acc0e7f68e363c15d8abb5cacce7143a1f7de8a/json
    5b12ef8fd57065237a6833039acc0e7f68e363c15d8abb5cacce7143a1f7de8a/VERSION
    5b12ef8fd57065237a6833039acc0e7f68e363c15d8abb5cacce7143a1f7de8a/layer.tar
    77e369743e242d5b13f6426b128b1609f8aa1091b935e50b0de81c8bc3e1510d/
    77e369743e242d5b13f6426b128b1609f8aa1091b935e50b0de81c8bc3e1510d/json
    77e369743e242d5b13f6426b128b1609f8aa1091b935e50b0de81c8bc3e1510d/VERSION
    77e369743e242d5b13f6426b128b1609f8aa1091b935e50b0de81c8bc3e1510d/layer.tar
    /tmp/tmp.R433Vilvoa $ docker load -i img.tar
    /tmp/tmp.R433Vilvoa $ docker run -it 77e369743e24 bash
    [root@4b899c017d55 /]# cat /etc/issue
    CentOS release 6.6 (Final)
    Kernel \r on an \m
    
    [root@4b899c017d55 /]#
