Newsletter-Go
=============

[![build][build-svg]][build-url] [![coverage][cover-svg]][cover-url]

A very basic newsletter program for POSIX servers.

It allow local server users to send a newsletter with their own email address to a list of subscribers.

It was designed for [CLUB1 community server](https://club1.fr/english), but may be usable by other UNIX-like servers (like *tilde* servers).

The design strategy of this piece of code is to take advantage of Sendmail-style `.forward` files combined with recipient delimiter (`+` addresses). On CLUB1's server, we use it with Postfix ([man](https://www.postfix.org/local.8.html)).

This is a rewrite of the previous Bash version <https://github.com/club-1/newsletter> in Go.

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
    - [x] change language of mail subscribe/unsubscribe interface
        - [x] english (default)
        - [x] french
- other
    - [x] fancy ascii banner
    - [x] log to the Syslog
    - [ ] store archives of newsletters


Installation
------------

- `newsletter` is the user facing program that can be used to setup and send the news.
- `newsletterctl` is the part that catch subscription and unsubscriptions.

They **have to** be inside the following subfolders :

    bin/newsletter
    sbin/newsletterctl

For example inside `/usr/local`.


Usage
-----

Users may setup, send, or stop the newsletter using the command line.

### Setup

    newsletter setup

Interactive setup to edit display name, newsletter title, language, and signature.

Create necessary `.forward` files if they do not exist.

Add `-v` option to increase verbosity.

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

### read logs

Logs are stored in `syslog` using the identifier `newsletter`.
You can read them using the following command:

    journalctl -t newsletter


Deployment
----------

    make deploy

Deploy settings can be modified by creating a `.env` file.

- `REMOTE` remote server name (default is `club1.fr`)
- `REMOTE_PATH` instal path (default is `/var/tmp/nlgo`)

Deployment will create:

    bin/newsletter
    sbin/newsletterctl


[build-svg]: https://github.com/club-1/newsletter-go/actions/workflows/build.yml/badge.svg
[build-url]: https://github.com/club-1/newsletter-go/actions/workflows/build.yml
[cover-svg]: https://github.com/club-1/newsletter-go/wiki/coverage.svg
[cover-url]: https://raw.githack.com/wiki/club-1/newsletter-go/coverage.html
