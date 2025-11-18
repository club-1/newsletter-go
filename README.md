Newsletter-Go
=============

A very simple newsletter for CLUB1 server members.

The design strategy of this piece of code is to take advantage of Postfix `.forward` files combined with recipient delimiter. 

This is a rewrite of <https://github.com/club-1/newsletter> in Go.



Features
--------

- subscription
    - [x] users can subscribe using email
        - [x] subscription verify sender's authenticiy by sending a confirm email
    - [x] users can unsubscribe using email
- newsletter sending
    - [x] plain text only
    - [ ] allow markdown formating
    - [x] can be send through CLI
    - [x] can be send through email
    - [x] send a preview email to owner before sending confirmation
- configuration
    - [x] subscribeds emails are stored line by line in a plain text file
    - [x] signature is stored as a plain text file
    - [x] advanced config is stored in JSON file
    - [x] interactive setup through CLI
    - [ ] change language of mail subscribers interface (currently hardcoded in english)
- other
    - [x] logger
    - [ ] store archives of newsletters

Usage
-----

### Initialize

    newsletter [-v] init

Create necessary `.forward` files. Add `-v` option to increase verbosity.


### Setup

    newsletter setup

Interactive setup to edit display name, newsletter title, and signature.

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
