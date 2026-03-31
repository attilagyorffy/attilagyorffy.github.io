---
title: "Stop Shipping Desktop Images to Phones"
subtitle: "Because shipping a 4000-pixel hero image to a phone on 3G is not a personality trait."
date: 2017-06-11
description: "How to build a picture element helper in Rails so browsers stop downloading comically oversized desktop images on phones. Your users on mobile will thank you."
summary: "Shipping a 4000-pixel hero image to a phone on 3G is not a personality trait. Here's a Rails helper that uses the picture element like a grown-up."
topics:
  - Ruby on Rails
  - Performance
type: Code
read_time: "3 min read"
footer: "If you're still serving full-resolution desktop images to phones, I'm not angry, just disappointed. Come find me on [Bluesky](https://bsky.app/profile/attilagyorffy.com), [Mastodon](https://fosstodon.org/@attila), [~~Twitter~~ X](https://twitter.com/attilagyorffy), or even [LinkedIn](https://linkedin.com/in/attilagyorffy) if you fancy discussing image optimisation in a place where people list \"synergy\" as a skill. Code lives on [GitHub](https://github.com/attilagyorffy), as always."
---

Right, quick tip. You know how every website you've ever built probably loads the same massive image regardless of whether the user is on a 27-inch iMac or a phone being held together with optimism and a cracked screen protector? Yeah, let's fix that. Here's how to generate a `<picture>` element in Rails that tells the browser to stop being an idiot and load the right image for the right screen.

The `<picture>` element is one of those HTML features that's been around long enough that you have absolutely no excuse for not using it. It lets you specify multiple image sources with media queries, and the browser picks the first one that matches. It's like a buffet, but the browser only takes what it can actually eat. Revolutionary concept, I know.

So here's a helper you can drop into your Rails app. Takes about thirty seconds to write, saves your users from downloading photographs the size of small countries:

```ruby
# application_helper.rb
module ApplicationHelper
  def responsive_image_tag(image, options = {})
    content_tag(:picture) do
      concat content_tag(:source, nil, media: '(max-width: 768px)', srcset: image.url(:thumbnail_mobile))
      concat content_tag(:source, nil, media: '(max-width: 960px)', srcset: image.url(:thumbnail_tablet))
      concat image_tag(image.url(:thumbnail_desktop), options)
    end
  end
end
```

That's it. That's the whole thing. You wrap your `<source>` elements inside a `<picture>` tag, slap some media queries on them, and the browser does the rest. The media queries here are standard breakpoints --- 768px for mobile, 960px for tablets --- but obviously you can change those to whatever arbitrary numbers your designer has decided are gospel this week.

<mark>A `<picture>` element can take any number of `<source>` children and will load the first one that matches the current screen.</mark> That's the beauty of it. You're not doing any clever JavaScript viewport detection nonsense. You're not loading all three images and hiding two of them with CSS like some sort of bandwidth arsonist. The browser genuinely only fetches the one it needs.

Now, the `image` argument I'm passing into this helper is a [CarrierWave](https://github.com/carrierwaveuploader/carrierwave) uploader object. If you're not familiar, CarrierWave lets you define different *versions* of an uploaded image --- so one upload gives you a mobile thumbnail, a tablet thumbnail, and a desktop version, each at the appropriate resolution. The `.url(:thumbnail_mobile)` calls are just fetching the URL for each version. If you're using ActiveStorage or Shrine or whatever the cool kids have moved onto this month, the principle is exactly the same --- you just swap in the equivalent method calls.

The fallback `image_tag` at the bottom is your standard `<img>` element. It's what loads when the browser doesn't support `<picture>`, which at this point basically means Internet Explorer, and if your users are still on IE, responsive images are the least of their problems. Or yours.

If you want to go deeper on the `<picture>` element --- and you should, because it's genuinely one of the better things to happen to HTML since they stopped trying to make `<marquee>` a thing --- have a look at the [MDN documentation](https://developer.mozilla.org/en/docs/Web/HTML/Element/picture) and the [Can I Use](http://caniuse.com/picture/embed/) page for current browser support. Spoiler: it's basically everything.

<ul class="takeaway">
<li>Use the <code>&lt;picture&gt;</code> element with <code>&lt;source&gt;</code> children and media queries to serve the right image for the right screen</li>
<li>Wrap it in a Rails helper so you write it once instead of copy-pasting HTML like it's 2004</li>
<li>The browser only downloads the matching source, so your users on mobile stop subsidising your enthusiasm for 4K photography</li>
</ul>
