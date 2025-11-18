Newsletter-Go
=============

A rewrite of <https://github.com/club-1/newsletter> in Go.

Usage
-----

### Initialize

    newsletter [-v] init

Create necessary `.forward` files. Add `-v` option to increase verbosity.


### Setup

    newsletter setup

Interactive from to setup From display name, newsletter title, and signature.

### Send newsletter

If your content is stored in a file:

    newsletter [-y] [-p] send SUBJECT CONTENT_FILE

Alteratively, you can pipe the content through STDIN:

    echo CONTENT | newsletter [-y] [-p] SUBJECT

This will send you a preview mail and ask for confirmation.`-y` will skip confirmation and preview mail.

If `-p` is set, action is limited to preview.

### Stop

    newsletter [-v] stop

Remove `.forward` files to deactivate newsletter. Add `-v` option to increase verbosity.



Deployment
----------

    make deploy

Deploy settings can be modified by creating a `.env` file.

- `REMOTE` remote server name (default is `club1.fr`)
- `REMOTE_PATH` instal path (default is `/var/tmp/nlgo`)

Deployment will create:

    bin/newsletter
    sbin/newsletterctl
