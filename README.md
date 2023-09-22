<!--
SPDX-FileCopyrightText: Amolith <amolith@secluded.site>

SPDX-License-Identifier: CC0-1.0
-->

# Willow

[![Go report card status][goreportcard-badge]][goreportcard]
[![REUSE status][reuse-shield]][reuse]
[![Donate with fosspay][fosspay-shield]][fosspay]

_Software release tracker supporting arbitrary forges_

## What is it?

Willow tracks software releases across arbitrary forge platforms by trying to
support one of the very few things they all have in common: the VCS. At the
moment, git is the _only_ supported VCS, but I would be interested in adding
Pijul, Fossil, Mercurial, etc. You can also track releases using RSS feeds.

Willow exists because decentralisation can be annoying. One piece of software
can be found on GitHub, another piece on GitLab, one on Bitbucket, a fourth on
SourceHut, and a fifth on the developer's self-hosted Forgejo instance. Forgejo
and GitHub have RSS feeds that only notify you of releases. GitLab doesn't
support RSS feeds for anything, just an API you can poke. Some software updates
might be on the developers' personal blog. Sometimes there are CVEs for specific
software and they get published somewhere completely different before they're
fixed in a release.

I want to bring all that scattered information under one roof so a developer or
sysadmin can pop open willow's web UI and immediately see what needs updating
where. I've recorded some of my other ideas and plans in [my wiki].

[my wiki]: https://wiki.secluded.site/hypha/willow

## Installation and use

* Clone the repo
* Build the binary with `CGO_ENABLED=0 go build .`
* Upload it to a remote server
* Execute the binary
* Reverse proxy `localhost:1337`
* Open the web UI
* Click `Track new project`
* Fill out the form
* Indicate which version you're currently on
* That's it!

Note that there's currently no authentication, so consider putting your instance
behind HTTP Basic Auth, keeping it private, or helping me implement
authentication.

## Questions & Contributions

Questions, comments, and patches can always be sent to my public inbox, but I'm
also in my IRC channel/XMPP room pretty much 24/7. However, I might not see
messages right away because I'm working on something else (or sleeping) so
please stick around!

If you're wanting to introduce a new feature and I don't feel like it fits with
this project's goal, I encourage you to fork the repo and make whatever changes
you like!

- Email: [~amolith/public-inbox@lists.sr.ht][email]
- IRC: [irc.nixnet.services/#secluded][irc]
- XMPP: [secluded@muc.secluded.site][xmpp]

_If you haven't used mailing lists before, please take a look at [SourceHut's
documentation](https://man.sr.ht/lists.sr.ht/), especially the etiquette
section._

[email]: mailto:~amolith/public-inbox@lists.sr.ht
[irc]: irc://irc.nixnet.services/#secluded
[xmpp]: xmpp:secluded@muc.secluded.site?join

[goreportcard-badge]: https://goreportcard.com/badge/git.sr.ht/~amolith/willow
[goreportcard]: https://goreportcard.com/report/git.sr.ht/~amolith/willow
[reuse]: https://api.reuse.software/info/git.sr.ht/~amolith/willow
[reuse-shield]: https://shields.io/reuse/compliance/git.sr.ht/~amolith/willow
[fosspay]: https://secluded.site/donate/
[fosspay-shield]: https://shields.io/badge/donate-fosspay-yellow
