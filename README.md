# DLGrab (docker layer grab)

## Requirements

You need to have docker running on your local machine. Remote daemons won't
work, though it wouldn't be hard to add them.

## Caveats

DLGrab shims the api endpoints docker touches when pushing an image. If
Docker starts expecting different responses from these endpoints, DLGrab may
stop working.

That said, it's probably more robust that trying to export directly from the
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
