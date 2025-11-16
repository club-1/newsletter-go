Newsletter-Go
=============

A rewrite of <https://github.com/club-1/newsletter> in Go.

Usage
-----

**Initialize newsletter**

    newsletter [-v] init

create necessary `.forward` files. Add `-y` option to increase verbosity.

**Preview**

    newsletter preview SUBJECT CONTENT_FILE

**Send a newsletter to subscribed addresses**

    newsletter [-y] send SUBJECT CONTENT_FILE

This will send you a preview mail and ask for confirmation. `-y` will skip confirmation and preview mail.


Deployment
----------

    make deploy

Deploy settings can be modified by creating a `.env` file.

- `REMOTE` remote server name (default is `club1.fr`)
- `REMOTE_PATH` instal path (default is `/var/tmp/nlgo`)

Deployment will create:

    bin/newsletter
    sbin/newsletterctl
