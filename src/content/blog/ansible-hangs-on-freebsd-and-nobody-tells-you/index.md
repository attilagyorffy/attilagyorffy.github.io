---
title: "Ansible Hangs on FreeBSD and Nobody Tells You"
subtitle: "Because FreeBSD ships without Python and Ansible has no idea what to do about it."
date: 2016-12-09
description: "FreeBSD does not ship Python, and Ansible does not ship patience. Here is how to bootstrap pkg and raw modules without the whole thing hanging on you forever."
summary: "FreeBSD doesn't ship Python. Ansible needs Python. One environment variable stops the whole thing from hanging forever."
topics:
  - FreeBSD
  - Infrastructure
type: Code
read_time: "2 min read"
footer: "If you have also been personally victimised by a package manager that ignores its own flags, come commiserate. I'm on [Bluesky](https://bsky.app/profile/attilagyorffy.com), [Mastodon](https://fosstodon.org/@attila), [~~Twitter~~ X](https://twitter.com/attilagyorffy), and technically [LinkedIn](https://linkedin.com/in/attilagyorffy), though I cannot imagine discussing FreeBSD bootstrap quirks there with a straight face. The rest of the damage is on [GitHub](https://github.com/attilagyorffy)."
---

Right. So FreeBSD does not come with Python preinstalled. Let that sink in. You have got Ansible on your control machine, ready to automate the world, and the target system is just sitting there like a bloke at a pub who does not speak the language. Ansible needs Python. FreeBSD does not have Python. Nobody told either of them about this arrangement beforehand.

Now, you would think, "Fine, I will just use `pkg` to install Python and get on with my life." And you would be wrong. Because the first time you invoke `pkg` on a fresh FreeBSD box, it pulls this little stunt:

```bash
The package management tool is not yet installed on your system.
Do you want to fetch and install it now? [y/N]:
```

A yes-or-no prompt. On a system you are trying to manage *programmatically*. Brilliant. And here is the best part: this happens even if you pass the `-y` flag. You are standing there telling `pkg`, "Yes, mate, I want this, I explicitly said yes," and `pkg` is ignoring you like a bouncer who has already decided you are not getting in. <mark>Ansible will just hang there on your control machine, waiting forever for a prompt that no human will ever answer.</mark>

The fix is absurdly simple once you know it exists, which of course you do not until you have wasted twenty minutes staring at a frozen terminal wondering if the internet is broken. You set the `ASSUME_ALWAYS_YES` environment variable to `yes` and use the `raw` module because, remember, there is no Python yet, so none of Ansible's nice modules work:

```bash
$ ansible freebsd -m raw -a "setenv ASSUME_ALWAYS_YES yes; pkg install -y python27"
```

That is it. That is the whole trick. An environment variable. Not a secret flag, not a kernel patch, not a blood sacrifice to the BSD daemon. Just an environment variable that tells `pkg` to stop asking stupid questions and get on with installing things like a proper package manager.

<ul class="takeaway">
<li>FreeBSD has no Python out of the box, so use Ansible's <code>raw</code> module to bootstrap it</li>
<li>The <code>-y</code> flag does not suppress the initial <code>pkg</code> bootstrap prompt — you need <code>ASSUME_ALWAYS_YES=yes</code></li>
<li>Once Python is installed, normal Ansible modules work and you can stop living like an animal</li>
</ul>
