Newsletter-Go
=============

A rewrite of <https://github.com/club-1/newsletter> in Go.

Usage
-----

### Initialize newsletter

    newsletter [-v] init

Create necessary `.forward` files. Add `-v` option to increase verbosity.

### Stop newsletter

    newsletter [-v] stop

Remove `.forward` files to deactivate newsletter. Add `-v` option to increase verbosity.

### Send a newsletter to subscribed addresses

If your content is stored in a file:

    newsletter [-y] [-p] send SUBJECT CONTENT_FILE

Alteratively, you can pipe the content through STDIN:

    echo CONTENT | newsletter [-y] [-p] SUBJECT

This will send you a preview mail and ask for confirmation.`-y` will skip confirmation and preview mail.

If `-p` is set, action is limited to preview.


Deployment
----------

    make deploy

Deploy settings can be modified by creating a `.env` file.

- `REMOTE` remote server name (default is `club1.fr`)
- `REMOTE_PATH` instal path (default is `/var/tmp/nlgo`)

Deployment will create:

    bin/newsletter
    sbin/newsletterctl
