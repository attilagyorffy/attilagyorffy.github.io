---
title: "npm's Progress Bar Halves Your Install Speed"
subtitle: "Because out of the box, npm behaves like a toddler with a megaphone in a library."
date: 2016-06-05
description: "npm ships with a progress bar that actively halves your install speed. A few lines in ~/.npmrc will shut it up, and your CI pipeline will thank you for it."
summary: "Out of the box, npm behaves like a toddler with a megaphone. A few lines in .npmrc make it twice as fast and half as annoying."
topics:
  - JavaScript
  - Performance
type: Code
read_time: 3 min read
footer: |
  If you've made it this far and you're still running npm without an `.npmrc`, I admire your commitment to chaos. Come find me on [Bluesky](https://bsky.app/profile/attilagyorffy.com), [Mastodon](https://fosstodon.org/@attila), [~~Twitter~~ X](https://twitter.com/attilagyorffy), and technically [LinkedIn](https://linkedin.com/in/attilagyorffy), though discussing npm configuration on LinkedIn feels like bringing a keg to a board meeting. The code is on [GitHub](https://github.com/attilagyorffy), where my progress bars are mercifully silent.
---

It has come to my attention that some of you are just… running `npm` as-is. Raw. Out of the box. Like buying a car and never adjusting the mirrors. I don't know who needs to hear this, but your package manager can be told to behave like a proper UNIX citizen instead of a hyperactive carnival barker narrating every file it touches.

What do I mean by that? Well, let me show you my `~/.npmrc` configuration, the small file that separates civilised developers from the animals:

```
progress=false
spin=false
loglevel=error
init.author.name=Attila Gyorffy
init.author.email=attila@example.com
init.author.url=https://attilagyorffy.com
//registry.npmjs.org/:_password=thisissecret
//registry.npmjs.org/:username=attilagyorffy
//registry.npmjs.org/:email=attila@example.com
//registry.npmjs.org/:always-auth=true
```

`~/.npmrc` is loaded and interpreted every single time you run an `npm` command. Every. Single. Time. So you might as well make it do something useful instead of just sitting there collecting dust like that gym membership you bought in January.

## The progress bar: a masterclass in self-sabotage

The first two parameters, `spin` and `progress`, turn off npm's beloved progress bar. "But I like the progress bar!" you say. Yeah, well, I like beer, but I don't drink it while operating heavy machinery. <mark>Turning off the progress bar results in installs that are roughly twice as fast.</mark> Let that sink in. The tool that is supposed to tell you how long things are taking is literally making things take longer. It's like hiring a bloke to time your 100-metre sprint and he tackles you at the 50.

Don't take my word for it. Go read [the famous GitHub issue](https://github.com/npm/npm/issues/11283) about it. But first, get yourself some popcorn and a cold drink, because the comment thread is the sort of beautiful disaster you usually have to pay Netflix for.

## The Rule of Silence, or: shut up, npm

The `loglevel` parameter tells npm to be less noisy. If you believe command-line tools should follow the Unix Rule of Silence — when a program has nothing surprising to say, it should say nothing — then this one's for you. Programs should be unobtrusive. They should do their job and keep their mouths shut, like a good butler or a competent plumber.

Now, setting `loglevel=error` won't stop npm from being a little too chatty. It's still going to mutter things at you like a drunk uncle at Christmas dinner. But it's an improvement. And here's the truly beautiful part: even if you set the log level to `silent` — the highest possible level of "please, for the love of God, stop talking" — npm will *still* report some unnecessary information. Because of course it will. It's npm. Asking it to be quiet is like asking a seagull not to steal your chips.

## The rest: module author business

The remaining lines in that config are for module authors. The `init.author.*` fields save you from typing your own name every time you run `npm init`, which, if you think about it, is absurd that you'd have to do in the first place. "Who are you?" asks the tool. Mate, we've been over this. I'm the same person I was three minutes ago when I ran it for the last package.

The registry authentication lines (`_password`, `username`, `email`, `always-auth`) handle authenticating with the npm registry. Nothing glamorous, but necessary if you want to publish packages without npm asking you to prove you exist every single time.

<h2 class="conclusion">Go forth and configure</h2>

Look, the above isn't exactly a PhD thesis in systems administration. It's a handful of lines in a dotfile. But it's the difference between a CLI that respects your time and one that's actively wasting it while showing you a pretty animation of itself wasting it.

Also, have a look at the [full npm config documentation](https://docs.npmjs.com/misc/config). There are more knobs to turn than you'd expect, and you might find something else worth tweaking. It's a bit like rummaging through the settings menu on your telly — mostly pointless, but every now and then you find a gem that makes you wonder why it wasn't on by default.

<ul class="takeaway">
    <li>Turn off <code>progress</code> and <code>spin</code> for dramatically faster installs</li>
    <li>Set <code>loglevel=error</code> to make npm behave like a respectable UNIX tool</li>
    <li>Pre-fill <code>init.author.*</code> so you never have to introduce yourself to your own computer again</li>
</ul>

Your `~/.npmrc` is right there, waiting. It takes thirty seconds to configure and it'll save you from years of watching a progress bar that's actively sabotaging you. If that's not a strong enough sales pitch, I don't know what to tell you. Maybe you enjoy suffering. Maybe you think the progress bar is pretty. Either way, don't come crying to me when your `npm install` takes twice as long as it should because a little animated bar is chewing through your CPU cycles for absolutely no reason.
