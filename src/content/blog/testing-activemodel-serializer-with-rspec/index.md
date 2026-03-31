---
title: "Testing ActiveModel::Serializer with RSpec"
subtitle: "How to unit-test ActiveModel::Serializer objects with RSpec now that the adapter layer has changed how serialization works."
date: 2015-01-24
description: "How to unit-test ActiveModel::Serializer objects with RSpec now that the adapter layer has changed how serialization works. The old approach quietly broke."
summary: "They rewrote the serializer gem and broke how you test it. Here's how to wire up adapters so your specs actually test what your API sends."
topics:
  - Ruby on Rails
  - Testing
type: Code
read_time: "3 min read"
footer: "If you've read this far and have opinions on testing serializers, I'd be glad to hear them. I'm on [Bluesky](https://bsky.app/profile/attilagyorffy.com), [Mastodon](https://fosstodon.org/@attila), [~~Twitter~~ X](https://twitter.com/attilagyorffy), and nominally on [LinkedIn](https://linkedin.com/in/attilagyorffy), though discussing RSpec edge cases there feels like wearing a suit to a hackathon. You can also find my projects on [GitHub](https://github.com/attilagyorffy), where the specs always pass. Eventually."
---

Nobody tests their serializers. Be honest. You've got model specs, controller specs, maybe even some request specs if you're feeling virtuous -- but serializers? They just sit there between your models and your API, quietly mangling data, and <mark>when they break the failure mode is not a 500 error but a subtly wrong JSON payload</mark> that breezes past every test you've written and blows up in a client's face on a Friday afternoon. Testing them in isolation catches that before you ruin your weekend.

So the brilliant minds behind [ActiveModel::Serializer](https://github.com/rails-api/active_model_serializers) decided what we really needed was another abstraction. Because why have one layer when you can have two? Enter the Adapter. Now a serializer defines *what* gets serialized (the attributes on the model), while an adapter controls *how* it gets serialized. Want a flat JSON response for one endpoint and a [JSONAPI](http://jsonapi.org/)-compliant envelope for another? Same serializer, different adapter. It's actually a decent idea, which is rare enough in the Ruby ecosystem that it deserves a standing ovation.

The catch -- because there's always a catch -- is that `ActionController::Serialization` no longer calls `to_json` on serializer objects directly. That'd be too simple. Instead, serializers get wrapped in an adapter via `ActiveModel::Serializer::Adapter.create`, which takes a serializer instance and returns whatever adapter your config says you want (`ActiveModel::Serializer.config.adapter` -- in my case, `:json_api`, because I enjoy pain).

Here's what this whole song and dance looks like when you actually fire up a console:

```ruby
profile = Profile.first
# => #<Profile _id: 544d2cc466756b966b010000, name: "Liquid", permalink: "liquid">

serializer = ProfileSerializer.new(profile)
# => #<ProfileSerializer:0x007fbeda323de0 @_routes=nil, @object=#<Profile _id: 544d2cc466756b966b010000, name: "Liquid", permalink: "liquid">, @root=false, @meta=nil, @meta_key=nil>

adapter = ActiveModel::Serializer::Adapter.create(serializer)
# => #<ActiveModel::Serializer::Adapter::JsonApi:0x007fbedd8c3cc0 @serializer=#<ProfileSerializer:0x007fbeda323de0 @_routes=nil, @object=#<Profile _id: 544d2cc466756b966b010000, name: "Liquid", permalink: "liquid">, @root=true, @meta=nil, @meta_key=nil>, @options={}, @hash={}, @top={}, @fieldset=nil>

adapter.as_json
# => {:profiles=>{:id=>"liquid", :name=>"Liquid"}
```

So if you had nice, tidy serializer specs that just called `to_json` and went home early -- congratulations, they're all lying to you now. You need to go through the adapter layer to test the actual output your API produces, or you're essentially testing a thing that doesn't exist.

Right. So given a serializer like this:

```ruby
class ProfileSerializer < ActiveModel::Serializer

  attributes :id, :name

  def id
    object.permalink
  end
end
```

Here's the spec that actually works. It wraps the serializer in an adapter -- exactly the way Rails does it -- and asserts against the resulting JSON. Full serialization pipeline, no controller, no routes, no mucking about. Just the truth:

```ruby
require 'rails_helper'

RSpec.describe ProfileSerializer, :type => :serializer do

  context 'Individual Resource Representation' do
    let(:resource) { build(:profile) }

    let(:serializer) { ProfileSerializer.new(resource) }
    let(:serialization) { ActiveModel::Serializer::Adapter.create(serializer) }

    subject do
      # I'm using a JSONAPI adapter, which means my profile is wrapped in a
      # top level `profiles` object.
      JSON.parse(serialization.to_json)['profiles']
    end

    it 'has an id that matches #permalink' do
      expect(subject['id']).to eql(resource.permalink)
    end

    it 'has a name' do
      expect(subject['name']).to eql(resource.name)
    end
  end
end
```

Why bother? Because these specs are stupidly fast -- we're talking milliseconds -- and each one <mark>catches attribute-mapping bugs that controller tests gloss over</mark>. They'll scream at you the second someone adds a field to the model and conveniently forgets to expose it in the API. Which is everyone. Everyone forgets. For more on the adapter rewrite and its glorious indirection, see the [ActiveModel::Serializers project on GitHub](http://github.com/rails-api/active_model_serializers).

<ul class="takeaway">
<li>Wrap serializers in <code>ActiveModel::Serializer::Adapter.create</code> to test the actual output your API produces</li>
<li>Serializer specs run in milliseconds and catch attribute-mapping bugs that controller tests miss</li>
<li>Always test through the adapter layer — calling <code>to_json</code> directly no longer matches what the controller sends</li>
</ul>
