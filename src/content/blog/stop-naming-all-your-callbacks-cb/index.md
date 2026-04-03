---
title: "Stop Naming All Your Callbacks cb"
subtitle: "A naming convention for async.waterfall callbacks that prevents shadow-scoping bugs and makes nested control flow readable at a glance."
date: 2015-03-20
description: "Why naming every callback cb in async.waterfall chains is a shadow-scoping bug waiting to happen, and what to call them instead. Node.js footgun explained."
summary: "Everyone names their callbacks \"cb\" and then spends forty minutes debugging why the wrong one fires. Two words fix it. You're looking for \"next.\""
topics:
  - JavaScript
type: Code
read_time: "2 min read"
footer: "If you've read this far and have thoughts on callback naming conventions, I'd love to hear from you. I'm on [Bluesky](https://bsky.app/profile/attilagyorffy.com), [Mastodon](https://fosstodon.org/@attila), [~~Twitter~~ X](https://twitter.com/attilagyorffy), and even [LinkedIn](https://linkedin.com/in/attilagyorffy), though naming callbacks is probably not the kind of content that thrives there. You can also find my projects on [GitHub](https://github.com/attilagyorffy), where I promise to always call my callbacks `next`."
---

Look, if every callback in your `async.waterfall` chain is named `cb`, you're basically playing Russian roulette with closures. JavaScript doesn't give a shit. It'll let the inner `cb` shadow the outer `cb` without so much as a warning. Everything compiles, the tests pass, you high-five yourself, and then at 2am on a Saturday, production does something inexplicable.

Here's what this trainwreck looks like in the wild:

```javascript
function doSomething(foo, bar, cb) {
  async.waterfall([
    function(cb) {
      // ...
    },
    function(returnVal, cb) {
      // ...
    },
    function(returnVal, cb) {
      // ...
    },
    function(returnVal, cb) {
      // ...
    },
    function(returnVal, cb) {
      // ...
    },
  ], function(err, returnValue) {
    // ...
  });
}
```

The outer function takes a `cb`. Every step inside the waterfall also takes a `cb`. They're completely different functions in completely different scopes, wearing the same bloody name like it's a uniform. Sure, the runtime sorts it out through lexical scoping --- it's a computer, that's its whole job. But you're not a computer. You're a person scrolling through code at 4pm on a Friday, and <mark>identical names across nested scopes are exactly the kind of thing that makes you reach for the wrong reference when editing under pressure</mark>.

The fix is so stupidly simple you'll be annoyed nobody told you sooner: just give each scope a different name. That's it. That's the whole trick.

- The parent function's callback is called `callback` --- those extra six characters cost nothing and immediately signal "this is the outer continuation"
- Each waterfall step's callback is called `next`, which conveys exactly what it does: advance to the next step in a linear sequence

Same example, now with a shred of self-respect:

```javascript
function doSomething(foo, bar, callback) {
  async.waterfall([
    function(next) {
      // ...
    },
    function(returnVal, next) {
      // ...
    },
    function(returnVal, next) {
      // ...
    },
    function(returnVal, next) {
      // ...
    },
    function(returnVal, next) {
      // ...
    },
  ], function(err, returnValue) {
    // ...
  });
}
```

Suddenly, the names actually mean something. `callback` is obviously the outer function's exit door --- no ambiguity, no squinting required. `next` reads like what it does: "move along to the next step, mate." There's no shadowing, no staring at the screen wondering which `cb` is which, and zero chance of accidentally firing the parent's callback from inside a waterfall step and ruining everyone's evening.

These tiny naming decisions add up fast. In callback-heavy Node code --- where five levels of nesting is depressingly normal --- the gap between lazy names and intentional ones is <mark>the difference between code you can maintain and code you can only rewrite</mark>.

<ul class="takeaway">
<li>Name the outer function's callback <code>callback</code></li>
<li>Name each waterfall step's callback <code>next</code></li>
<li>Distinct names per scope prevent shadow-scoping bugs and make nested control flow readable</li>
</ul>
