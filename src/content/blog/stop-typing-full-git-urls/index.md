---
title: "Stop Typing Full Git URLs"
subtitle: "Because life is too short to type `git@github.com:` like a caveman."
date: 2016-12-12
description: "How to set up Git URL shortcuts so you can clone from GitHub, Heroku, and Gist without typing full URLs like some kind of animal. One file, permanent fix."
summary: "Life is too short to type git@github.com: like a caveman. A few lines in .gitconfig and you're cloning with gh:user/repo."
topics:
  - devtools
type: Code
read_time: "2 min read"
footer: "If you've been typing full Git URLs this whole time and you're now feeling personally attacked, good. Come commiserate. I'm on [Bluesky](https://bsky.app/profile/attilagyorffy.com), [Mastodon](https://fosstodon.org/@attila), [~~Twitter~~ X](https://twitter.com/attilagyorffy), and technically [LinkedIn](https://linkedin.com/in/attilagyorffy), though I mostly use that to decline recruiter messages. You can also find my dotfiles and other questionable life choices on [GitHub](https://github.com/attilagyorffy), where I promise every clone starts with `gh:`."
---

Right, so here's a thing that Git can do that almost nobody talks about, presumably because they're all too busy typing out full SSH URLs like it's 2004 and they're being paid by the keystroke. Git has a built-in mechanism to rewrite URLs during commands. Most people who stumble upon this use it to force HTTPS instead of the `git://` protocol. Which is fine. Sensible, even. But also deeply boring.

What's far more interesting is that with a bit of creativity and about thirty seconds of configuration, you can turn this feature into proper shortcuts. I'm talking about cloning a GitHub repo by typing `git clone gh:user/repo` and watching the magic happen. That's it. That's the whole command. No `git@github.com:`, no fumbling around trying to remember whether it's HTTPS or SSH, no existential crisis in the terminal. Just `gh:` and the repo name, like a civilised human being.

<mark>The fact that more people don't do this is, frankly, baffling, and I can only assume it's because they enjoy suffering.</mark>

## The config

Throw this into your `~/.gitconfig` and immediately start feeling superior to everyone you work with:

```ini
# Allow cloning repositories using shortcuts.
# For example:
#   git clone gh:attilagyorffy/foo

[url "git@github.com:"]
  insteadOf = gh://
  pushInsteadOf = "github:"
  pushInsteadOf = "git://github.com/"
[url "git://github.com/"]
  insteadOf = "github:"
[url "git@gist.github.com:"]
  insteadOf = "gst:"
  pushInsteadOf = "gist:"
  pushInsteadOf = "git://gist.github.com/"
[url "git://gist.github.com/"]
  insteadOf = "gist:"
[url "git@heroku.com:"]
  insteadOf = "heroku:"
```

That's the whole thing. No gems. No plugins. No seventeen-step installation process involving a package manager nobody's heard of. Just plain old Git configuration that's been sitting there, patiently waiting for you to notice it, like a well-behaved dog at a shelter.

## What it actually does

The `insteadOf` directive tells Git: "Hey, whenever you see this prefix, quietly replace it with the real URL before doing anything." So when you type `git clone gh://attilagyorffy/foo`, Git silently rewrites that to `git clone git@github.com:attilagyorffy/foo` behind the scenes. You get SSH access to GitHub without ever typing the full URL. It's like having a butler for your terminal.

The `pushInsteadOf` variant is even sneakier. It only kicks in for push operations, so you can pull over a read-only protocol and push over SSH. Because apparently Git thought of everything except making this feature discoverable.

And yes, there's one for Gists too, because sometimes you need to clone a Gist and you shouldn't have to feel bad about it. The `gst:` prefix sorts that right out. Heroku gets one as well, because deploying to Heroku via `heroku:` is the sort of small luxury that makes you wonder why you ever did it the long way.

<h2 class="conclusion">Go forth and be lazy</h2>

Look, I'm not going to pretend this is some groundbreaking revelation. It's a Git config trick. But it's one of those tiny quality-of-life improvements that, once you've set it up, makes you irrationally annoyed every time you see someone else type out a full GitHub URL. And that low-grade, simmering superiority? That's the real gift.

<ul class="takeaway">
<li>Use <code>insteadOf</code> in your <code>~/.gitconfig</code> to create URL shortcuts for any Git host</li>
<li>Use <code>pushInsteadOf</code> to rewrite URLs only for push operations, keeping pulls on a different protocol</li>
<li>Stop typing full URLs like you're being punished for something</li>
</ul>
