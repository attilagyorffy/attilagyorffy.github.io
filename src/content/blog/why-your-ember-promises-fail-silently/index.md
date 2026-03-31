---
title: "Why Your Ember Promises Fail Silently"
subtitle: "RSVP swallows unhandled rejections by default. A five-line initializer fixes that."
date: 2014-08-07
description: "Ember's RSVP promise library swallows errors by default. A five-line initializer surfaces them in the console and saves you hours of confused debugging."
summary: "Something breaks in an Ember promise and you get... nothing. A blank page, a clean console, and a slow descent into madness. Four lines fix it."
topics:
  - JavaScript
type: Code
read_time: "3 min read"
footer: "Got opinions on promises, error handling, or Ember in general? Come find me on [Bluesky](https://bsky.app/profile/attilagyorffy.com), [Mastodon](https://fosstodon.org/@attila), [~~Twitter~~ X](https://twitter.com/attilagyorffy), or [LinkedIn](https://linkedin.com/in/attilagyorffy). You can also browse my other projects on [GitHub](https://github.com/attilagyorffy)."
---

So you write a model hook. You fetch some data. It works beautifully. You feel like a genius. Then something breaks — a dodgy response, a missing property, an exception buried six layers deep in a `then` chain — and your Ember app just… does nothing. No error. No stack trace. The page sits there half-rendered like a website having an existential crisis. Brilliant.

The villain here is [RSVP](https://github.com/tildeio/rsvp.js/), the promise library Ember uses under the hood. RSVP dutifully follows the Promises/A+ spec, which — in its infinite wisdom — says that if a promise is rejected and nobody's listening, the error should just be quietly thrown in the bin. Not logged. Not surfaced. Binned. That's not a bug, that's the spec working as designed: <mark>unhandled rejections vanish into the void</mark>.

What this means in the real world is that any uncaught exception inside a promise chain — a typo, a null reference, a failed assertion — produces absolutely zero output. Nothing. You're flying blind. So you do what any rational person does: you start plastering `console.log` statements everywhere like crime scene tape, toggling breakpoints, muttering profanities, and slowly losing the will to live while trying to figure out which promise ate your error for lunch. You can easily burn an hour on this before you realize you've been debugging the wrong bloody layer the entire time.

## The fix

Here's the thing though — RSVP actually has a global event hook for exactly this situation. It's been sitting there the whole time, waiting for you to notice it, like a fire extinguisher you walk past every day. You subscribe to an `'error'` event that fires whenever a promise is rejected and nobody catches it. Stick it in an Ember initializer, and suddenly every swallowed error shows up in your console like it always should have:

```javascript
// app/initializers/rsvp-error-reporting.js

import Ember from 'ember';

export default {
  name: 'ember-rsvp-error-reporting',

  initialize: function (container, application) {
    Ember.RSVP.on('error', function(reason) {
      console.group('Ember.RSVP error:')
      console.info(reason);
      console.groupEnd();
    });
  }
};
```

That's it. That's the whole thing. The initializer runs once when your Ember app boots, calls `Ember.RSVP.on('error', ...)` to register a global listener, and then just sits there doing its job like a responsible adult. Whenever a promise rejects without a handler, RSVP fires the callback with the rejection reason — typically an `Error` object with a message and stack trace. The `console.group` wrapper keeps it all neatly collapsed in DevTools so your console doesn't look like someone vomited stack traces all over it.

Without this fix, a failing API call in a route's `model` hook gives you a blank page and a spotless console — the debugging equivalent of a crime scene that's been wiped clean. With it, you get a grouped log entry with the full error and stack trace pointing straight at the problem like a helpful witness. <mark>The difference is between spending five seconds reading an error message and spending an hour guessing.</mark>

Drop this file into any Ember CLI project, forget it exists, and get on with your life. Honestly, the fact that this isn't in every Ember app's boilerplate from day one is a minor tragedy of the JavaScript ecosystem.

<ul class="takeaway">
<li>Register an <code>Ember.RSVP.on('error', ...)</code> listener in an initializer</li>
<li>Every swallowed promise rejection becomes immediately visible in the console</li>
<li>Add this to every Ember project from day one — it costs nothing and saves hours</li>
</ul>
