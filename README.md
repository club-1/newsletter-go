Newsletter-Go
=============

A rewrite of <https://github.com/club-1/newsletter> in Go.

Usage
-----

    newsletter init

Initialize newsletter


Deployment
----------

    make deploy

Deploy settings can be modified by creating a `.env` file.

- `REMOTE` remote server name (default is `club1.fr`)
- `REMOTE_PATH` instal path (default is `/var/tmp/nlgo`)

Deployment will create:

    bin/newsletter
    sbin/newsletterctl
