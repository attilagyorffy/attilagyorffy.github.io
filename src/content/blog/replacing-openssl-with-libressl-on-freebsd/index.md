---
title: "Replacing OpenSSL with LibreSSL on FreeBSD"
subtitle: "After Heartbleed proved OpenSSL could not be trusted blindly, LibreSSL became the obvious alternative. Here is how to swap it into FreeBSD base and what breaks when you do."
date: 2016-07-02
description: "After Heartbleed proved OpenSSL could not be trusted blindly, LibreSSL became the obvious alternative. Here is how to swap it into FreeBSD base cleanly."
summary: "Heartbleed happened because OpenSSL had its own custom allocator that hid buffer overflows from every tool designed to catch them. Here's how to rip it out."
topics:
  - FreeBSD
  - Security
type: Code
read_time: "7 min read"
footer: "If you care about SSL, FreeBSD, or the security of your stack, I'd love to hear from you. Find me on [Bluesky](https://bsky.app/profile/attilagyorffy.com), [Mastodon](https://fosstodon.org/@attila), [~~Twitter~~ X](https://twitter.com/attilagyorffy), or [LinkedIn](https://linkedin.com/in/attilagyorffy). You can also check out my FreeBSD fork and other projects on [GitHub](https://github.com/attilagyorffy)."
---

OpenSSL is that friend who keeps crashing your car and somehow still has your spare keys. Heartbleed in 2014 didn't just leak private keys on roughly 17% of the internet's TLS servers -- it did so because of plain old C memory mismanagement, the programming equivalent of leaving your front door open and acting shocked when someone walks in. The codebase had ballooned to the point where nobody on earth could properly review it, which is exactly how you'd design software if your goal was a security disaster. LibreSSL was the OpenBSD team basically saying "right, give it here," forking the whole thing, gutting the nonsense, and building something a human being could actually audit. On FreeBSD, you can swap it in today.

## Replacing OpenSSL with LibreSSL in FreeBSD 10.3 base

Thanks to [Bernard Spil](https://twitter.com/sp1l) and the [HardenedBSD](http://hardenedbsd.org/) team -- people who apparently enjoy doing thankless, unglamorous security work so the rest of us can sleep at night -- jamming LibreSSL into the FreeBSD base system is now surprisingly not-terrible.

What I did was steal -- sorry, "build upon" -- Bernard's existing patches and make them less painful to use. I branched off from the latest release tree (releng/10.3), dropped the current portable LibreSSL version (2.4.1) into `src/crypto`, and applied Bernard's patches as proper git commits, like a civilised person.

Doing it this way -- via a git fork instead of lobbing raw patch files at people -- has a few advantages that should be obvious but apparently need spelling out:

- You can check out the source that already contains the patches and that is guaranteed to compile
- I expect this method to be easier to maintain as we are moving towards the release of FreeBSD 11
- Git has a mechanism to test whether patches would apply successfully before having to actually apply them

```bash
# git clone https://github.com/attilagyorffy/freebsd.git /usr/src
# cd /usr/src
# git checkout releng/10.3-libressl
# make buildworld && make buildkernel && make installkernel && make installworld
# reboot
```

## What does this mean for web developers?

Look, swapping OpenSSL out of the base system feels great. Real dopamine hit. But <mark>your stack is only as secure as its weakest link</mark>, and if you stop here you're basically putting a deadbolt on your front door while leaving every window wide open.

The FreeBSD `pkg` system installs precompiled binaries, which means every SSL dependency hiding in those packages was lovingly compiled against the very OpenSSL you just swore off. To actually get LibreSSL all the way through your stack, you need to compile from ports. Yes, from source. In 2016. I know.

I went on a little compilation spree to see what actually builds against LibreSSL without throwing a tantrum:

**Software that compiled without errors:**

- nginx-devel-1.11.1
- ruby23-2.3.1
- postgresql95-client-9.5.3
- postgresql95-server-9.5.3
- go-1.6.2
- elixir-1.2.6
- erlang-18.3.4.1
- git-2.9.0
- fish-2.2.0 (because it's my favourite shell)

**Software that needs more work: node.js**

Node is the one kid in class who refuses to use the shared textbook, and naturally it's the one who needs it most. A quick scroll through the [node vulnerability archive](https://nodejs.org/en/blog/vulnerability/) reveals a comically large number of security advisories that trace right back to its bundled copy of OpenSSL. Because node drags along its own personal OpenSSL like a security blanket instead of linking against the system library, replacing OpenSSL in base does absolutely nothing for it. If you've got a node process facing the internet (Express, Ghost, whatever) and you aren't religiously updating your node port, that bundled OpenSSL is basically an engraved invitation to get owned.

There have been conversations about shoving LibreSSL into node, but the idea got binned mostly because it didn't build on MSVC. Because apparently Windows compatibility is more important than not leaking your users' private keys. As of LibreSSL 2.2.2, you can actually produce Visual Studio native builds using CMake, so that excuse is evaporating. There's hope, but don't hold your breath.

## How about alternative SSL implementations?

The Core Infrastructure Initiative has been trying to modernise OpenSSL, which is a bit like renovating a house while it's on fire -- technically possible, but the pace is glacial and everyone's still coughing. LibreSSL took the far more satisfying approach: <mark>delete tens of thousands of lines of dead code, rip out the custom allocator that hid Heartbleed from address sanitizers, and enforce modern memory practices</mark>. That's not refactoring, that's an exorcism. And it's exactly why I'd pick LibreSSL over a "reformed" OpenSSL any day of the week.

Google's BoringSSL deserves a mention too, but Google being Google, they built it for themselves and then basically put up a sign that says "don't touch this":

> "Although BoringSSL is an open source project, it is not intended for general use, as OpenSSL is. We don't recommend that third parties depend upon it. Doing so is likely to be frustrating because there are no guarantees of API or ABI stability."

So yeah, that rules out BoringSSL unless you fancy chasing an API that changes whenever a Googler has a creative afternoon. LibreSSL is the only real drop-in alternative. It came out of OpenBSD, shares enough BSD DNA with FreeBSD that the compatibility issues are minimal, and nobody's actively telling you not to use it. Low bar, but here we are.

## Will LibreSSL land in FreeBSD base?

Honestly? Who knows. Some ports still refuse to build against LibreSSL like stubborn toddlers, and FreeBSD 11 is already past code freeze, so it'll ship with good old OpenSSL because that's how these things always go. On the bright side, OpenBSD already ships LibreSSL in base -- because of course they do, they wrote the thing -- and the HardenedBSD team has successfully built FreeBSD 11 with LibreSSL. The groundwork is there. It's a question of when, not whether, the rest of the ecosystem stops dragging its feet.

## Why a git fork instead of patches?

Bernard's repository distributes the changes as a single patch set, and if you've ever tried applying patches with `patch(1)` you know it's about as reliable as a weather forecast:

> "If no original file is specified on the command line, patch will try to figure out from the contents of the patch file which file to edit. Stripping leading pathnames is not always reliable."

A git branch, on the other hand, guarantees reproducible builds. You check out the exact source tree that compiles, git's own diffing and merge machinery handles the rest, and you don't spend your evening screaming at offset errors. Fancy that.

<ul class="takeaway">
<li>Replace OpenSSL in FreeBSD base by building from the <code>releng/10.3-libressl</code> branch</li>
<li>Recompile ports from source to get LibreSSL through your entire stack — <code>pkg</code> binaries still link against OpenSSL</li>
<li>Node.js bundles its own OpenSSL and ignores base — keep it updated separately</li>
</ul>

## Special thanks

- Thanks to [Bernard Spil](http://brnrd.eu/) and the [HardenedBSD team](http://hardenedbsd.org/) for their great work on the LibreSSL patches.
- Thanks to [Bradley T. Hughes](http://bradleythughes.github.io/) for his insights on the node ports.
