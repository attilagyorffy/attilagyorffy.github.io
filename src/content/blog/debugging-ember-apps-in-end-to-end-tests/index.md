---
title: "Debugging Ember Apps in End-to-End Tests"
subtitle: "When your Capybara tests drive a full Ember app, failures are opaque. Here is how to load Ember Inspector into the Selenium-driven browser."
date: 2014-11-19
description: "When Capybara drives a full Ember app, test failures are opaque. Loading Ember Inspector into the Selenium browser gives you something useful to work with."
summary: "Your Capybara test fails, the browser closes, and all you've got is \"expected true, got false.\" Loading Ember Inspector into Selenium changes everything."
topics:
  - JavaScript
  - Testing
type: Code
read_time: "3 min read"
footer: "Still wrestling with Ember testing or have a better approach? I'd love to hear about it. Find me on [Bluesky](https://bsky.app/profile/attilagyorffy.com), [Mastodon](https://fosstodon.org/@attila), [~~Twitter~~ X](https://twitter.com/attilagyorffy), or [LinkedIn](https://linkedin.com/in/attilagyorffy). More code and open source bits on [GitHub](https://github.com/attilagyorffy)."
---

End-to-end testing an Ember app through Capybara sounds dead simple on paper: Selenium drives a real browser, the browser boots the Ember app, Capybara clicks buttons and checks things. Beautiful. Except the moment anything goes sideways, you're left staring at a blank browser window like a dog watching a magic trick. You have absolutely no idea what just happened inside the Ember runtime.

Now, the "proper" Ember testing approach dodges this whole mess by mocking the backend and running integration tests entirely in JavaScript-land. Lovely. Very tidy. Also completely useless if you want to know whether your Rails API, your Ember frontend, and the actual HTTP requests flying between them work together in the real world. If you need to prove that a critical business flow doesn't explode across that boundary, you need Capybara driving a real browser against a real server. No shortcuts. (For a proper deep dive into why this is such a pain, [Jo Liss' presentation](http://www.slideshare.net/jo_liss/testing-ember-apps) is excellent.)

Here's where it gets fun. When these end-to-end tests fail, the error messages are about as helpful as a fortune cookie. Capybara tells you "the expected text never appeared on the page." Cheers, mate. But why? Was the route wrong? Did a promise reject silently into the void? Is the model sitting there in some cursed half-state? You can't just pop open DevTools in a Selenium-driven browser like you would normally. <mark>The Ember app is a black box.</mark>

## Loading Ember Inspector into Selenium

So here's the stupid-simple fix that took me way too long to figure out. Chrome has a `--load-extension` flag that lets you shove unpacked extensions into the browser Selenium launches. Clone and build [Ember Inspector](https://github.com/emberjs/ember-inspector) locally (the [README](https://github.com/emberjs/ember-inspector#building-and-testing) tells you how), point the Chrome driver at it, and suddenly <mark>every test run gives you the full Inspector panel: routes, data, components, promises</mark> — all the stuff that was completely invisible five minutes ago. It's almost offensive how easy it is.

Chuck the built extension into `spec/support/ember_inspector` and wire it up:

```ruby
# Ensure we load the Selenium Driver and the Chrome Driver helper.

group :test do
  gem 'selenium-webdriver'
  gem 'chromedriver-helper'
end
```

```ruby
# place this in rails_helper.rb if you are using RSpec

Capybara.register_driver :selenium do |app|
  ember_inspector = File.expand_path('../support/ember-inspector', __FILE__)
  Capybara::Selenium::Driver.new(app,
    browser: :chrome,
    switches: [
      "--load-extension=#{ember_inspector}"
    ]
  )
end
```

That's it. That's the whole bloody thing. The custom Capybara driver tells Chrome to load the extension on launch, so every test session starts with the Inspector right there waiting for you. Slap a `binding.pry` in your test, switch over to the browser, open DevTools, and poke around your Ember app exactly the way you would during development. You know, like a normal human being.

*Because I'm a generous soul, I've since wrapped this into the [capybara-ember-inspector](https://rubygems.org/gems/capybara-ember-inspector) gem, so you don't even have to do the wiring yourself. You're welcome.*

## Notes

- The Inspector also exists as a Firefox addon, if you're one of those people. Fair warning though, the Capybara setup is different, because of course it is.
- You'll need `chromedriver` installed on your machine to drive Chrome via Selenium. On a Mac it's just `brew install chromedriver`. One whole command. You'll survive.

<ul class="takeaway">
<li>Use Chrome's <code>--load-extension</code> flag to inject Ember Inspector into Selenium-driven test runs</li>
<li>Pause with <code>binding.pry</code>, switch to the browser, and inspect Ember state during failures</li>
<li>The <code>capybara-ember-inspector</code> gem packages this setup into a single dependency</li>
</ul>
