---
title: "Zero-Config Mobile Testing with Pow and Bonjour"
subtitle: "pow-mDNS uses Bonjour to make your Pow apps accessible from phones, tablets, and other computers on the local network -- no IP addresses, no extra servers."
date: 2014-12-11
description: "pow-mDNS uses Bonjour to make your Pow apps accessible from phones, tablets, and other computers on the local network -- no IP addresses, no extra servers."
summary: "I wanted to load my Rails app on an iPad without looking up my IP address every five minutes. So I made mDNS do it for me."
topics:
  - Testing
  - devtools
type: Code
read_time: "5 min read"
footer: "If you've read this far and you're as excited about zero-configuration tooling as I am, let's connect. I'm on [Bluesky](https://bsky.app/profile/attilagyorffy.com), [Mastodon](https://fosstodon.org/@attila), [~~Twitter~~ X](https://twitter.com/attilagyorffy), and occasionally [LinkedIn](https://linkedin.com/in/attilagyorffy). You can also find pow-mDNS and my other projects on [GitHub](https://github.com/attilagyorffy), where I continue to build tools that Just Work."
---

Here's how testing on real devices usually goes: you look up your IP address, forget it immediately, type it wrong twice, spin up a second server on port 3000, and then your phone still can't see anything because the universe hates you. I got tired of this ritual, so I built **pow-mDNS**. Install it with `npm install -g pow-mdns`, and every Pow app on your machine becomes visible to any phone, tablet, or computer on the same network. Zero configuration. Like magic, except it actually works.

## The problem

If you've ever touched a Rails app, you probably know Pow -- the beautifully lazy zero-config Rack server where you symlink your app into `$HOME/.pow/` and suddenly it's available at a `.dev` domain. No running `rails server`, no port numbers, no nonsense. You symlink it and it works. It's the closest thing to sorcery that Ruby developers have ever experienced.

Here's the catch, though -- and there's always a catch. Pow's magic relies on a local DNS resolver that routes everything to `127.0.0.1`. That resolver only exists on your machine. So those lovely `.dev` domains? Completely invisible to everything else on the network. Your iPhone can't see them. Your iPad can't see them. Your colleague's laptop certainly can't see them. It's like throwing a party and forgetting to send out the address.

Basecamp's [xip.io](http://xip.io/) was one workaround: a clever public wildcard DNS service that lets you hit `myapp.192.168.1.107.xip.io` from any device on the same network. And look, it works. But you need to know your current IP address, which changes every time you switch networks or your DHCP lease decides to renew itself. So now you're running `ifconfig` every twenty minutes like some kind of network archaeologist. That's not zero-config. That's just config with extra steps.

## How pow-mDNS solves it

The answer was sitting right under our noses the whole time: mDNS -- multicast DNS, the protocol behind Apple's [Bonjour](https://developer.apple.com/bonjour/index.html). mDNS resolves hostnames to IP addresses on local networks without a dedicated name server. It's the same technology that lets your Mac magically find printers, AirPlay devices, and other computers without you lifting a finger. And if it can find a printer -- which, frankly, is a miracle every time -- there's no reason it can't advertise a web app.

pow-mDNS uses Bonjour to advertise each of your Pow apps as an HTTP service on the local network. Combined with xip.io for the actual DNS resolution, any Bonjour-capable device on the network can discover and access your apps without you ever typing an IP address. The IP lookup, the xip.io domain construction -- all that tedious rubbish you were doing by hand? pow-mDNS does it for you. You're welcome.

Look, I'm not going to stand here and pretend I reinvented networking. Most of the heavy lifting belongs to Pow, xip.io, and mDNS. What pow-mDNS does is wire them together into a single command so that <mark>the developer experience actually matches the promise of zero-configuration</mark>. Because apparently someone had to do the plumbing, and that someone was me.

## Why not use existing tools?

Yeah, tools like GhostLab and Adobe Edge Inspect exist. They solve the same problem -- synchronized cross-device testing -- and they work fine for a lot of people. But they're closed-source commercial products that want your money, and some of them need a proprietary agent installed on each device, which is roughly as fun as a tax audit. I wanted something open source, stitched together from freely available parts (Pow and xip.io are both open source), that just bloody works out of the box. No license key. No phoning home. No guilt.

## What comes next

pow-mDNS is one piece of a larger toolkit, and now that the hardest problem (getting off the couch to build it) is solved, here's what's next:

- **Baking LiveReload in** — automatically refreshing browsers on all connected devices when files change
- **Selective advertising** — choosing which Pow applications to advertise on the network
- **Creating an OS X menu bar app** — a native interface for managing pow-mDNS
- **Decentralised access** — enabling access to applications across different networks

The source is on [GitHub](https://github.com/liquid/pow-mdns). Install it, point your phone at the network, and for the love of all that is holy, stop memorizing IP addresses. Life's too short.

<ul class="takeaway">
<li>Install with <code>npm install -g pow-mdns</code> to advertise Pow apps via Bonjour</li>
<li>Any device on the same network can discover and access your apps without IP addresses</li>
<li>Combine Pow, xip.io, and mDNS into a single zero-config workflow</li>
</ul>
