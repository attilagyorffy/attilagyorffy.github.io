---
title: "Stop Naming Your Go Constructors New"
subtitle: "Why Parse is better than New when your Go constructor is really parsing a string."
date: 2026-03-12
description: "Why Parse is a better name than New when your Go constructor is really validating and parsing a string. Naming matters more than you think, especially in Go."
summary: "A grown man loses his mind over a three-letter word in a function name. Turns out he has a point, but still."
topics:
  - Go
type: Code
read_time: "4 min read"
footer: "If you've read this far and you still think `NewFoo` is fine, I respect your commitment to being wrong. Either way, come tell me about it. I'm on [Bluesky](https://bsky.app/profile/attilagyorffy.com), [Mastodon](https://fosstodon.org/@attila), [~~Twitter~~ X](https://twitter.com/attilagyorffy), and technically [LinkedIn](https://linkedin.com/in/attilagyorffy), though God knows why anyone would discuss Go naming conventions there. You can also find the code that started all of this on [GitHub](https://github.com/attilagyorffy), where I promise not to name anything `New` unless it bloody well deserves it."
---

A validated value type takes a raw string and gives you back something typed and trustworthy. In Go, the default instinct is often to expose that as a pair like `NewFoo` and `MustNewFoo`. Which sounds fine right up until you think about it for more than six seconds.

`NewFoo` tells you only that a `Foo` comes back. Great. Spectacular. Very informative. `MustNewFoo` adds only the error-handling strategy, which is basically the API equivalent of saying, "same mystery, but louder." Neither name tells you what the function is actually doing with the input.

And that matters, because when a function accepts a raw string and returns a validated type, it is usually not *creating* something from thin air like a magician pulling a rabbit out of a hat. It is interpreting input, checking whether that input is valid, and only then reifying it as a typed value. In other words: it is parsing. Calling that `New` is technically survivable, but semantically a bit rubbish.

Take an [AT Protocol](https://atproto.com/) record key: a string that must match a specific format. This is exactly the sort of API where `NewFoo` / `MustNewFoo` looks conventional at first, but hides the semantics that matter.

```go
type RecordKey struct{ value string }

// "New" — but what is actually happening here?
// Is this for code reconstructing from storage?
// For HTTP input?
// For a test using a known-valid literal?
func NewRecordKey(v string) (RecordKey, error) { ... }

// "Must" explains error handling (panic vs return),
// but still says nothing about intent.
func MustNewRecordKey(v string) RecordKey { ... }
```

Consider two very different callers:

```go
// Store: "I read this from SQLite; turn it back into the typed value."
rk, err := atproto.NewRecordKey(rkeyStr)

// Test: "I know this literal is valid; just give me the type."
rk := atproto.MustNewRecordKey("3jt5k2e4xab2s")
```

Both callers use the same conceptual entry point, but they are doing very different jobs.

The store layer is not creating a new record key in any meaningful domain sense. It is reconstructing one that already exists in persisted form. <mark>Nothing is being born. No miracle is occurring.</mark> A string came out of SQLite and you are trying to turn it back into a proper type without lying to yourself about what just happened.

The test is different again. It is not parsing untrusted input in the ordinary sense; it is asserting that a hardcoded literal is valid. That is closer to saying, "this value had better be valid or the programmer has done something daft."

Those are distinct relationships to data, yet `New` makes them look like the same bland little ceremony. That is the real weakness of `New` here: not that it is technically wrong, but that it hides intent behind a name so generic it may as well be called `DoThing`.

Go's `net/netip` package shows what this looks like done right:

```go
// "Parse" — this string comes from outside the type.
addr, err := netip.ParseAddr("192.168.1.1")

// "MustParse" — this should be valid; panic if it is not.
loopback := netip.MustParseAddr("127.0.0.1")
```

`Parse` means: take a textual representation, validate it, and turn it into a typed value. `MustParse` means: do the same, but treat failure as a programmer error --- somebody in the codebase deserves a long, disappointed stare.

The pattern appears throughout the standard library:

```go
url.Parse("https://example.com")                     // net/url
time.Parse(time.RFC3339, "2024-01-01T00:00:00Z")     // time
template.Must(template.New("t").Parse("..."))        // text/template
regexp.MustCompile(`\d+`)                            // regexp
```

## Applying the pattern

Using that convention, the record key API becomes clearer:

```go
// Validation lives in one place.
// Call sites: store layer, HTTP handlers, config parsing.
func ParseRecordKey(v string) (RecordKey, error) {
    if !tidRegexp.MatchString(v) {
        return RecordKey{}, fmt.Errorf("record key must be a valid TID: %q", v)
    }
    return RecordKey{value: v}, nil
}

// Panic wrapper calls Parse — no duplicated logic.
// Call sites: test fixtures, package-level constants.
func MustParseRecordKey(v string) RecordKey {
    rk, err := ParseRecordKey(v)
    if err != nil {
        panic(err)
    }
    return rk
}
```

Now the call sites read with much more precision:

```go
// Store: reconstructing from persistence.
rk, err := atproto.ParseRecordKey(rkeyStr)

// Test: asserting a known-valid literal.
rk := atproto.MustParseRecordKey("3jt5k2e4xab2s")
```

The benefit is not just stylistic. The API now nudges the caller toward the right mental model. Instead of reaching for a generic `New`, they must implicitly answer a useful question: am I parsing input, or am I asserting a trusted literal?

That is the whole game. Good APIs make the correct path feel obvious. Bad names do the opposite: they smear together different situations and then act surprised when readers have to squint.

## When New is still right

None of this means `New` is bad. `New` is perfectly fine. Lovely even. It just needs to stop turning up to jobs that belong to `Parse`. `New` is best reserved for cases where the operation is genuinely construction rather than interpretation.

Aggregate constructors that assemble already-typed values should usually keep `New`:

```go
// Takes validated types — no string parsing happens here.
func NewActor(username Username, domain Domain, publicKey *rsa.PublicKey) (Actor, error)

// Assembles an aggregate root from typed parts.
func NewContent(id ContentID, kind ContentKind, body Body) (Content, error)

// Service constructor — wires dependencies.
func NewAdapter(client *Client, domain CanonicalDomain, store RecordKeyStore) *Adapter
```

Likewise, simple constructors where any input is acceptable can still use `New`:

```go
func NewSummary(v string) Summary
func NewPublishedAt(t time.Time) PublishedAt
```

In both cases, the function is not trying to interpret a serialized representation. It is either composing typed parts or wrapping a value where no meaningful parsing step exists.

<h2 class="conclusion">The real point</h2>

Names are not decorative. They are part of the contract you present to the caller. The point is not to be clever; the point is to be honest.

<mark>Good names should reflect the semantic operation, not just the return type.</mark>

Because if your API takes a dodgy little string from the outside world, interrogates it, validates it, and only then agrees to let it into polite society as a proper type, that function did not *new* anything. It parsed. Pretending otherwise is like watching airport security frisk a man for ten minutes and then calling the whole process `NewPassenger`.

<ul class="takeaway">
<li>If the input is a raw string that must be validated against some textual format, call it <code>Parse</code></li>
<li>If the inputs are already typed values being assembled into something larger, call it <code>New</code></li>
</ul>

Do that, and your code reads more clearly, your APIs carry their own intent, and the next poor bastard reading your package will not have to perform forensic analysis on a function called `MustNewFoo` just to work out that it was parsing a bloody string all along.
