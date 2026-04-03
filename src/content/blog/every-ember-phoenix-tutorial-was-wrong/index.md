---
title: "Every Ember-Phoenix Tutorial Was Wrong"
subtitle: "Because every existing tutorial was either broken, outdated, or wilfully ignoring the conventions both frameworks ship with."
date: 2016-07-24
description: "Wiring up Ember and Phoenix with JSON:API for TodoMVC, because every existing tutorial got it wrong and nobody seemed particularly bothered about that."
summary: "Every existing tutorial was either broken or ignoring the conventions both frameworks ship with. So I built one from scratch."
topics:
  - JavaScript
type: Code
read_time: "4 min read"
footer: "If you have opinions about Ember adapters, Phoenix pipes, or the ongoing human tragedy that is CORS configuration, I am always happy to hear them. Find me on [Bluesky](https://bsky.app/profile/attilagyorffy.com), [Mastodon](https://fosstodon.org/@attila), [~~Twitter~~ X](https://twitter.com/attilagyorffy), and technically [LinkedIn](https://linkedin.com/in/attilagyorffy), though I cannot imagine why you would want to discuss TodoMVC there. The code is on [GitHub](https://github.com/attilagyorffy), where it has been sitting patiently since 2016, waiting for someone to appreciate it."
---

Right, so I wanted to learn [Phoenix](https://www.phoenixframework.org/). Fair enough. And because I am apparently incapable of learning anything without building something pointlessly elaborate, I decided the best way to do that was to wire up a Phoenix backend to the infamous [TodoMVC](http://todomvc.com/) example using ember-cli and the [JSON:API](http://jsonapi.org/) standard. A todo list. Backed by two frameworks. Using a formal serialisation specification. Because apparently I hate free time.

Now, there are tutorials out there for this. Plenty of them. And they fall into two reliable categories: ones that no longer work because the authors wrote them during a specific fourteen-minute window when that version of Ember was current, and ones that technically work but treat the conventions of both frameworks with the same casual disregard you'd give a Terms of Service checkbox.

On the Ember side, half the examples introduce custom serialiser and adapter code that you absolutely do not need. Ember ships with a perfectly stable [JSONAPI adapter](https://github.com/emberjs/data/blob/master/addon/adapters/json-api.js) right out of the box. It is sitting there, ready to go, doing exactly the thing these tutorials are reinventing from scratch with bespoke boilerplate. It is like building your own wheels because you forgot your car already had some.

On the Phoenix side, the tutorials are even more entertaining. Variable bindings everywhere. Which, fine, I get it, if you are coming from Ruby you might reach for the familiar patterns. But mate, you are writing Elixir now. Elixir has pipes. Beautiful, elegant, functional pipes. <mark>Using variable bindings instead of pipes in Elixir is like buying a Ferrari and then pushing it to the shops.</mark>

And then there is CORS. Oh, CORS. The thing that has caused more junior developers to silently weep into their keyboards than any other web specification in history. Setting up the right headers between an Ember app running on one port and a Cowboy server running on another seemed to be giving people the sort of trouble normally reserved for flat-pack furniture assembly.

So I decided to build the whole bloody thing from scratch. Not because I thought I was smarter than everyone else, but because learning by doing is the only way I actually retain anything that isn't a grudge.

## TL;DR

In a nutshell, here is everything that needed to happen to get the Ember frontend and Phoenix backend speaking to each other like civilised adults:

- Configure Ember Data to use a JSON:API serialiser (`DS.JSONAPIAdapter`)
- Point Ember Data at the Cowboy server
- Set the necessary CORS headers in Ember for Ajax requests to the Phoenix API server
- Set up CORS in Phoenix using the [Corsica plug](https://github.com/whatyouhide/corsica)
- Teach Phoenix how to handle requests with `application/vnd.api+json` content types
- Serialise controller parameters in Phoenix using the [ja_serializer](https://github.com/AgilionApps/ja_serializer) package

That is it. Six steps. Not twelve blog posts, not a weekend retreat, not a blood sacrifice to the CORS gods. Six steps, each one perfectly reasonable once you stop fighting the frameworks and start using what they actually give you.

## Get the code

Both repositories are on my GitHub. Have at them. Break them. Learn something. Or just clone them and pretend you wrote the whole thing yourself at a job interview. I will not judge.

- [Phoenix API backend](http://github.com/attilagyorffy/todos-api-phoenix)
- [Ember CLI frontend](https://github.com/attilagyorffy/todos-ui-ember-cli/tree/todos-api-phoenix)

<h2 class="conclusion">Karma Points</h2>

This thing did not happen in a vacuum. A few genuinely helpful humans pointed me in the right direction at various moments of confused desperation:

- [Andrea Leopardi](https://twitter.com/whatyouhide) --- author of the Corsica plug, without which CORS would still be ruining my evenings
- [Balint Erdi](https://twitter.com/baaz) --- pointed me to the right Ember Data adapter, saving me from writing adapter code like some kind of animal
- [Gabor Babicz](https://twitter.com/xeppelin) --- Professional Rubber Duck (TM), which is the most undervalued role in all of software engineering
- [Sander Hahn](https://twitter.com/sanderhahn) --- helped debug the Phoenix API with curl, which is the developer equivalent of a doctor saying "let's have a look"

<ul class="takeaway">
<li>Use what the frameworks give you — Ember's built-in JSONAPI adapter exists for a reason</li>
<li>Write idiomatic Elixir — use pipes, not variable bindings that look like Ruby with a funny accent</li>
<li>CORS is not magic — it is six headers and a plug, and you will survive</li>
</ul>
