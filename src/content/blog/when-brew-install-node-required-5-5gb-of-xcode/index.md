---
title: "When brew install node required 5.5GB of Xcode"
subtitle: "A snapshot from the Yosemite beta era, when Homebrew's build patches hadn't caught up and installing Node meant downloading all of Xcode."
date: 2014-07-09
description: "On the Yosemite beta, Homebrew's build patches hadn't caught up yet. Installing Node.js meant downloading the entire Xcode beta — all 5.5 gigabytes of it."
summary: "Apple released Yosemite and broke the compiler. Installing a JavaScript runtime now required downloading an entire IDE. A love letter to 2014."
topics:
  - devtools
type: Code
read_time: "2 min read"
footer: "If you've made it through a Node.js install guide from 2014, you deserve a medal. Come reminisce about the good old days with me on [Bluesky](https://bsky.app/profile/attilagyorffy.com), [Mastodon](https://fosstodon.org/@attila), [~~Twitter~~ X](https://twitter.com/attilagyorffy), or even [LinkedIn](https://linkedin.com/in/attilagyorffy). You can also find what I'm building these days on [GitHub](https://github.com/attilagyorffy), where the tooling has thankfully improved since the Yosemite era."
---

*This post is from July 2014, when OS X Yosemite was still in beta and everything was on fire. The specific disaster described here was patched ages ago. I'm keeping it as a monument to the suffering — a love letter to the era when Apple's development toolchain would casually ruin your afternoon.*

So there I was, fresh Yosemite beta install, feeling like a pioneer. I typed `brew install node` and waited for the magic to happen. It did not. No missing dependency, no network hiccup — just a straight-up compilation error, like the computer was personally offended. Homebrew builds Node.js from source, and the community patches that kept things working on Mavericks hadn't been updated for Yosemite yet. The Command Line Tools — normally all Homebrew needs — just stood there and shrugged.

The fix? Oh, you'll love this: <mark>download the entire Xcode Beta 3 — all 5.5 gigabytes of it — just so that `brew install node` could compile a JavaScript runtime</mark>. Five and a half gigs. To run JavaScript. On the server. That's like buying a house because you needed a screwdriver. Xcode ships a full SDK with headers and libraries that the standalone Command Line Tools don't bother to include, and during the Yosemite beta, those missing bits were exactly what Node.js demanded like some kind of entitled houseguest.

## The workaround

The steps are almost insultingly simple once you've made peace with downloading half a DVD's worth of IDE:

- Download the Xcode beta from the Apple Developer site
- Launch Xcode once to accept the license agreement and let it install its components
- Run `brew install node`

```bash
brew install node
```

That's it. Three steps. One of which involves downloading a full-blown IDE you will literally never open again, just so a package manager can compile a dependency. This was the cutting edge, folks. The pinnacle of developer experience on a brand-new Apple operating system in the year of our lord 2014.

The underlying [compilation issue](https://github.com/Homebrew/homebrew/issues/25294) was a well-known ritual that happened with every major macOS release, like clockwork, like taxes, like your build breaking the one day you actually need it. The Homebrew community always caught up eventually, but <mark>in the gap between a new OS beta and updated build patches, you were on your own</mark> — and the answer was always the same: "just install all of Xcode, mate." Every. Single. Time.

<ul class="takeaway">
<li>On new macOS betas, Homebrew's build patches lag behind — expect compilation failures</li>
<li>The Command Line Tools alone may not be enough; sometimes you need the full Xcode SDK</li>
<li>This class of problem surfaced with every major macOS release — budget time for toolchain breakage on day one</li>
</ul>
