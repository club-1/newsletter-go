Newsletter-Go
=============

A rewrite of <https://github.com/club-1/newsletter> in Go.

Usage
-----

**Initialize newsletter**

    newsletter init

**Preview**

    newsletter preview SUBJECT CONTENT_FILE

**Send a newsletter to subscribed addresses**

    newsletter send SUBJECT CONTENT_FILE


Deployment
----------

    make deploy

Deploy settings can be modified by creating a `.env` file.

- `REMOTE` remote server name (default is `club1.fr`)
- `REMOTE_PATH` instal path (default is `/var/tmp/nlgo`)

Deployment will create:

    bin/newsletter
    sbin/newsletterctl
