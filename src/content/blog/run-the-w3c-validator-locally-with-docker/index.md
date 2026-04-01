---
title: "Run the W3C Validator Locally with Docker"
subtitle: "The W3C markup validator is indispensable, but depending on a remote service for your test suite is fragile. Run it locally in a Docker container instead."
date: 2016-01-06
description: "The W3C markup validator is indispensable, but depending on a remote service for your test suite is fragile. Run it locally in a Docker container instead."
summary: "The W3C validator requires the internet, which is exactly the thing that isn't working when you need it. So we put a validator in a container."
topics:
  - Testing
  - devtools
type: Code
read_time: "3 min read"
footer: "If valid markup matters to you as much as it does to me, let's talk about it. Find me on [Bluesky](https://bsky.app/profile/attilagyorffy.com), [Mastodon](https://fosstodon.org/@attila), [~~Twitter~~ X](https://twitter.com/attilagyorffy), or [LinkedIn](https://linkedin.com/in/attilagyorffy). You can also browse my other projects on [GitHub](https://github.com/attilagyorffy), where I occasionally validate things before pushing them."
---

Browsers are enablers. They look at your mangled nightmare of a document, quietly fix it behind your back, and render something that looks close enough. Not a word of complaint. Which is exactly why <mark>invalid HTML tends to hide real bugs</mark>: forms that silently bin your users' data, nesting so broken it looks fine in Chrome but implodes in Safari, unclosed tags that swallow entire sections on mobile. If your markup doesn't validate, something is wrong, and you really want to find that out in your test suite, not when a customer emails you a screenshot at 2am.

In the Rails world, the [`be_valid_asset`](https://github.com/unboxed/be_valid_asset) gem does the right thing here: it ships your rendered markup off to the official W3C validator and fails your spec if it comes back dodgy. Brilliant idea, one tiny problem: <mark>it depends on a live internet connection to a remote service you do not control</mark>. So you're on a train, or a coffee shop with Wi-Fi held together by prayers, or a plane where the internet costs twelve dollars and doesn't actually work, and suddenly your entire test suite is red for reasons that have absolutely nothing to do with your code. Lovely.

Now, the W3C Validator Service [can be installed locally](https://validator.w3.org/docs/install.html), but getting it running natively on macOS is the kind of afternoon you don't get back. Docker, on the other hand, makes it stupidly easy.

## Running the validator in Docker

Peter Mescalchin, absolute legend, has already done the hard work and packaged the whole thing into a ready-made [html5validator Docker container](https://hub.docker.com/r/magnetikonline/html5validator/). Just fire it up with a port forward so your test suite can talk to it:

```bash
$ docker run -d -p 8888:80 magnetikonline/html5validator
```

That pulls the image (first time only, relax), starts the container in the background, and maps port 8888 on your machine to the Apache service inside. Once it's up, open `http://localhost:8888` in your browser and bask in the glory of a W3C validator that lives on your own bloody laptop. The API endpoint you actually care about is at the `/check` path:

`http://localhost:8888/check`

## Configure BeValidAsset to use the local container

Here's the beautiful part: `be_valid_asset` already accepts a custom validator host, so pointing it at your shiny local container is literally one line of configuration. One. Line.

```ruby
# in spec/support/be_valid_asset.rb
BeValidAsset::Configuration.markup_validator_host = 'http://192.168.99.100:32770'
```

If you want to get fancy with more configuration, knock yourself out and check the [README](https://github.com/unboxed/be_valid_asset).

<ul class="takeaway">
<li>Run the W3C validator locally with <code>docker run -d -p 8888:80 magnetikonline/html5validator</code></li>
<li>Point <code>be_valid_asset</code> at <code>localhost:8888</code> to remove the internet dependency from your test suite</li>
<li>Your tests should never break because of someone else's server</li>
</ul>

*Full disclosure, because I'm not a monster: `be_valid_asset` was built by yours truly and my colleagues at [Unboxed](http://unboxed.co/). The Docker container was built by Peter Mescalchin ([@magnetikonline](http://magnetikonline.com/)), who saved us all from a very boring afternoon.*
