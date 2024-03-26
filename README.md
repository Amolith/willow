<!--
SPDX-FileCopyrightText: Amolith <amolith@secluded.site>

SPDX-License-Identifier: CC0-1.0
-->

# Willow

[![Go report card status][goreportcard-badge]][goreportcard]
[![REUSE status][reuse-shield]][reuse]
[![Donate with fosspay][fosspay-shield]][fosspay]

_Forge-agnostic software release tracker_

![screenshot of willow's current web UI](.files/2024-02-24.png)

_This UI is Amolith's attempt at a balance between simple, pleasant, and
functional. Amolith is not a UX professional and would **very** much welcome
input from someone more knowledgeable!_

## What is it?

_If you'd rather watch a short video, Amolith gave a 5-minute [lightning talk on
Willow] at the 2023 Ubuntu Summit._

[lightning talk on Willow]: https://youtu.be/XIGxKyekvBQ?t=29900

Willow helps developers, sysadmins, and homelabbers keep up with software
releases across arbitrary forge platforms, including full-featured forges like
GitHub, GitLab, or [Forgejo] as well as more minimal options like [cgit] or
[stagit].

[Forgejo]: https://forgejo.org/
[cgit]: https://git.zx2c4.com/cgit/
[stagit]: https://codemadness.org/stagit.html

It exists because decentralisation, as wonderful as it is, does have some pain
points. One piece of software is on GitHub, another piece is on GitLab, one on
Bitbucket, a fourth on [SourceHut], a fifth on the developer's self-hosted
Forgejo instance.

[SourceHut]: https://sourcehut.org/

The capabilities of each platform can also differ, further complicating the
space. For example, Forgejo and GitHub have APIs and RSS release feeds,
SourceHut has an API and RSS feeds that notify you of _all_ activity in the
repo, GitLab only has an API, and there's no standard for discovering the
capabilities of arbitrary git frontends like [legit].

[legit]: https://github.com/icyphox/legit

And _then_ you have different pieces of information in different places; some
developers might publish release announcements on their personal blog and some
projects might release security advisories on an external platform prior to
publishing a release.

All this important info is scattered all over the internet. Willow brings some
order to that chaos by supporting both RSS and one of the _very_ few things all
the forges and frontends have in common: their **V**ersion **C**ontrol
**S**ystem. At the moment, [Git] is the _only_ supported VCS, but we're
definitely interested in adding support for [Pijul], [Fossil], [Mercurial], and
potentially others.

[Git]: https://git-scm.com/
[Pijul]: https://pijul.org/
[Fossil]: https://www.fossil-scm.org/
[Mercurial]: https://www.mercurial-scm.org/

Amolith (the creator) has recorded some of his other ideas, thoughts, and plans
in [his wiki].

[his wiki]: https://wiki.secluded.site/hypha/willow

## Installation and use

**Disclaimers:** 
1. Prebuilt binaries will be available with the [v0.0.1] release, greatly
   simplifying installation.
2. We consider the project _alpha-quality_. There will be bugs.
3. Amolith has tried to make the web UI accessible, but is unsure of its current
   usability.
4. The app is not localised yet and English is the only available language.
5. Help with any/all of the above is most welcome!

[v0.0.1]: https://todo.sr.ht/~amolith/willow?search=status%3Aopen%20label%3A%22v0.0.1%22
[communication platforms]: #contributing

### Installation

This assumes Willow will run on an always-on server, like a VPS.

* Clone the repo with `git clone https://git.sr.ht/~amolith/willow`
* Enter the repo's folder with `cd willow`
* Build the binary with `CGO_ENABLED=0 go build -ldflags="-s -w" -o willow
  ./cmd`
* Transfer the binary to the server however you like
* Execute the binary with `./willow`
* Edit the config with `vim config.toml`
* Daemonise Willow using systemd or OpenRC or whatever you prefer
* Reverse-proxy the web UI (defaults to `localhost:1313`) with Caddy or NGINX or
  whatever you prefer

### Use

* Create a user with `./willow -a <username>`
* Open the web UI (defaults to `localhost:1313`, but [installation] had you put
  a proxy in front)
* Click `Track new project`
* Fill out the form and press `Next`
* Indicate which version you're currently on and press `Track releases`
* You're now tracking that project's releases!

[installation]: #installation

If you no longer use that project, click the `Delete?` link to remove it, and,
if applicable, Willow's copy of its repo.

If you're no longer running the version Willow says you've selected, click the
`Modify?` link to select a different version.

If there are projects where your selected version does _not_ match what Willow
thinks is latest, they'll show up at the top under the **Outdated projects**
heading and have a link at the bottom of the card to `View release notes`.
Clicking that link populates the right column with those release notes.

If there are projects where your selected version _does_ match what Willow
thinks is latest, they'll show up at the bottom under the **Up-to-date
projects** heading.

## Contributing

Contributions are very much welcome! Please take a look at the [ticket
tracker][todo] and see if there's anything you're interested in working on. If
there's specific functionality you'd like to see implemented and it's not
mentioned in the ticket tracker, please describe it through one of [the
communication platforms](#communication) below so we can discuss its inclusion.
If we don't feel like it fits with Willow's goals, you're encouraged to fork the
project and make whatever changes you like!

### Collaboration

Some people dislike GitHub, some people dislike SourceHut, and some people
dislike both. Collaboration happens on multiple platforms so anyone can
contribute to Willow however they like. Any of the following are suitable, but
they're listed in order of Amolith's preference:

- [SourceHut]
  - **Distributed:** contributions are either through [git send-email], which
    requires you to have SMTP access to an email address, or through SourceHut's
    web UI, which requires a SourceHut account.
  - **Open source:** SourceHut components are licenced under AGPL, BSD, and
    possibly others.
- [Radicle]
  - **Distributed:** contributions are through the [Heartwood protocol], which
    requires you to at least set up a local Radicle node.
  - **Open source:** Radicle components are licenced under Apache, MIT, GPL, and
    possibly others.
- [Codeberg]
  - **Centralised:** contributions are through Codeberg pull requests and
    require a Codeberg account.
  - **Open source:** Codeberg is powered by Forgejo, which is licensed under MIT.
- [GitHub]
  - **Centralised:** contributions are through GitHub pull requests and require
    a GitHub account.
  - **Mixed:** _components_ of GitHub are open source, such as the syntax
    highlighter, but everything that makes GitHub _useful_ is proprietary.

[SourceHut]: https://sr.ht/~amolith/willow
[git send-email]: https://git-send-email.io
[Radicle]: https://app.radicle.xyz/nodes/radicle.secluded.site/rad:z34saeE8jnN5KbGRuLSggJ3eeLtew
[Heartwood protocol]: https://radicle.xyz/guides/protocol
[Codeberg]: https://codeberg.org/Amolith/willow
[GitHub]: https://github.com/Amolith/willow

### Communication

Questions, comments, and patches can always go to the [mailing list][email], but
there's also an [IRC channel][irc] and an [XMPP MUC][xmpp] for real-time
interactions.

- Email: [~amolith/willow@lists.sr.ht][email]
- IRC: [irc.libera.chat/#willow][irc]
- XMPP: [willow@muc.secluded.site][xmpp]

[email]: mailto:~amolith/willow@lists.sr.ht
[irc]: ircs://irc.libera.chat/#willow
[xmpp]: xmpp:willow@muc.secluded.site?join
[todo]: https://todo.sr.ht/~amolith/willow

_If you haven't used mailing lists before, please take a look at [SourceHut's
documentation](https://man.sr.ht/lists.sr.ht/), especially the etiquette
section._

### Configuring git...

…for <code>git send-email</code>

``` shell
git config sendemail.to "~amolith/willow@lists.sr.ht"
git config format.subjectPrefix "PATCH willow"
git send-email [HASH]
```

…for signing the [DCO]

``` shell
git config format.signOff yes
```

[DCO]: https://developercertificate.org/

### Required tools

- [Go](https://go.dev/)
- [gofumpt](https://github.com/mvdan/gofumpt)
  - Stricter formatting rules than the default `go fmt`
- [golangci-lint](https://golangci-lint.run/)
  - Aggregates various preinstalled Go linters, runs them in parallel, and makes
    heavy use of the Go build cache
- [Staticcheck](https://staticcheck.dev/)
  - Uses static analysis to find bugs and performance issues, offer
    simplifications, and enforce style rules

### Suggested tools

- [just](https://github.com/casey/just)
  - Command runner to simplify use of the required tools
- [air](https://github.com/cosmtrek/air)
  - Watches source files and rebuilds/executes the project when sources change

[goreportcard-badge]: https://goreportcard.com/badge/git.sr.ht/~amolith/willow
[goreportcard]: https://goreportcard.com/report/git.sr.ht/~amolith/willow
[reuse]: https://api.reuse.software/info/git.sr.ht/~amolith/willow
[reuse-shield]: https://shields.io/reuse/compliance/git.sr.ht/~amolith/willow
[fosspay]: https://secluded.site/donate/
[fosspay-shield]: https://shields.io/badge/donate-fosspay-yellow
