---
title: "Bringing Liquidsoap to Homebrew: A Formula for Community Radio"
subtitle: "A twenty-year-old streaming language with no Homebrew formula. That seemed worth fixing."
date: 2026-03-22
description: "Liquidsoap is a programming language for audio streaming with no Homebrew formula. Here is how I packaged it for macOS — OCaml, OPAM, and all."
summary: "Liquidsoap has powered community radio for twenty years, but installing it on macOS meant wrestling with OPAM or Docker. A Homebrew formula fixes that."
topics:
  - Developer Tools
  - Infrastructure
type: Code
read_time: "5 min read"
footer: "If you have ever lost an evening to crossfade timing or codec arguments, come say hello. I am on [Bluesky](https://bsky.app/profile/attilagyorffy.com), [Mastodon](https://fosstodon.org/@attila), [~~Twitter~~ X](https://twitter.com/attilagyorffy), and technically [LinkedIn](https://linkedin.com/in/attilagyorffy), though discussing OCaml packaging there feels like bringing vinyl to a networking event."
---

If you have ever tried to run an internet radio station, you have probably hit the same wall. You start with a playlist — maybe a folder of MP3s and a script that feeds them to Icecast. It works, until it does not. The stream dies at 3 AM and nobody notices. Two tracks play back-to-back with wildly different volumes. There is no crossfade, no jingles, no way to cut to a live input when a DJ shows up. You bolt on more scripts, more cron jobs, more glue. It gets ugly fast.

This is the problem Liquidsoap solves.

## What Liquidsoap is

Liquidsoap is a programming language designed specifically for audio and video streaming. Not a playlist manager. Not a config file for a streaming server. An actual statically typed language where sources of audio are first-class values that you can compose, transform, and route.

The idea started at the Ecole Normale Superieure de Lyon around 2004, when a group of students wanted a better way to run their campus radio. Twenty years later, it powers Radio France, AzuraCast, Radionomy, and countless community stations. The language has grown to handle video, HLS output, SRT ingest, speech synthesis, YouTube streaming, and more — but the core philosophy has not changed: simple things should be simple, complex things should be possible.

A basic Liquidsoap script reads like a description of what you want your radio to do:

```plaintext
music = playlist("~/Music")
jingles = playlist("~/Jingles")
s = rotate(weights=[1, 4], [jingles, music])
s = crossfade(s)
output.icecast(%mp3, host="localhost", port=8000, password="hackme", mount="/stream", s)
```

That is a radio station. Jingles every four tracks, crossfaded transitions, streaming to Icecast in MP3. If any of these sources fail — say the music directory gets unmounted — Liquidsoap detects it at type-check time, before the script even runs. If you want a live input that takes priority over the playlist, you add a `fallback`. If you want time-based scheduling, there is `switch`. The type system ensures you cannot accidentally create a stream that might go silent.

This is fundamentally different from chaining together ffmpeg commands, cron jobs, and shell scripts. <mark>Liquidsoap understands audio at a semantic level — clocks, synchronisation, fallibility, track boundaries — and handles the hard parts so you do not have to.</mark>

I have been building custom audio pipelines and wanted to iterate on `.liq` scripts locally — test the playlist logic, tweak fallback chains, verify encoding settings. The kind of thing you want a fast feedback loop for. So I reached for `brew install` and found nothing. No formula. Your options were wrestling with OPAM directly (slow, fragile, leaves a `.opam` directory in your home) or running the Docker image. <mark>Every friction point in installation is a potential community radio builder who gives up before writing their first script.</mark> I decided to fix that.

## Bringing it to Homebrew

Liquidsoap is written in OCaml, which makes packaging non-trivial. It uses OPAM to pull in dozens of OCaml libraries at build time, dune as its build system, and has a complex relationship with system libraries like ffmpeg and libcurl. The build process is not a simple `./configure && make` — it is closer to bootstrapping a small OCaml ecosystem inside a temporary directory, compiling everything, then extracting the binary and its runtime files. The binary also needs to find its standard library (`.liq` scripts) and unicode data at runtime, and the default build mode bakes in paths that assume a Linux FHS layout. Not exactly the kind of thing `brew create` handles for you.

The formula follows the same OPAM-based pattern that the handful of other OCaml formulas in homebrew-core (semgrep, flow, zero-install) use, with some liquidsoap-specific adaptations. Liquidsoap has three build modes for resolving runtime paths. The "posix" mode hardcodes paths like `/usr/share/liquidsoap/libs` — obviously wrong for Homebrew. The formula patches these at build time to point to Homebrew's prefix, then builds with `LIQUIDSOAP_BUILD_TARGET=posix`. Instead of building the OCaml compiler from source (which takes 10+ minutes), the formula uses Homebrew's OCaml package directly via `--compiler=ocaml-system`. OPAM gets a temporary root inside the build directory — nothing touches your home directory or persists after install.

**The build steps boil down to:**

1. Patch runtime paths and a vendored library's dune file (the `cry` module references `bytes`, a compat library removed in OCaml 5.x)
2. Initialize OPAM with the system compiler
3. Install OCaml dependencies from the project's opam files
4. Install the ffmpeg OCaml bindings
5. Build with dune
6. Install, relocate man pages and stdlib files to Homebrew-conventional locations
7. Copy camomile unicode data from the OPAM switch before it is cleaned up

The whole thing builds in about 2.5 minutes on an M1 Max.

## Room to grow

The current formula is intentionally minimal — ffmpeg covers the most important use case. But there are bindings that ffmpeg does not replace: **LADSPA/Lilv** for external audio effect plugins, **OSC** for real-time control from hardware or apps, **Prometheus** for production monitoring, and **sqlite3** for persistent metadata-driven playlists. These get compiled into the binary at build time — there is no runtime plugin system, and homebrew-core does not support `--with-foo` build options. If you need them today, install via OPAM directly or use a third-party tap. They are on my list to explore adding as default dependencies in future updates.

Note that the formula does not output audio to your speakers directly — `output()` falls back to `output.dummy` without a native audio backend. For local testing, write to a file and play it back:

```bash
liquidsoap 'output.file(%wav, fallible=true, "/tmp/test.wav", sine(duration=3.))'
afplay /tmp/test.wav
```

One thing that bothered me about the formula was the source patching. The posix build target hardcodes its runtime paths in a static OCaml file, so the formula had to `inreplace` six paths before building — fragile, version-sensitive, and the kind of thing that breaks silently on upgrades. The right fix was upstream: I [sent a patch to Liquidsoap](https://github.com/savonet/liquidsoap/pull/5045) that replaces the static file with a dune generation rule, so packagers can set paths through environment variables instead. The defaults stay identical, but the Homebrew formula can now just set `LIQUIDSOAP_LIBS_DIR` and friends instead of rewriting source code. That PR has been accepted, so future formula updates will be considerably cleaner.

<h2 class="conclusion">Three words from your own radio</h2>

The [PR is up on homebrew-core](https://github.com/Homebrew/homebrew-core/pull/273636). Once merged, anyone on macOS is three words from building their own radio. If you have been curious about internet radio, the [Liquidsoap book](http://www.liquidsoap.info/book/) and the [community Discord](https://discord.gg/rsay42QJYU) are good places to start.

<ul class="takeaway">
<li><code>brew install liquidsoap</code> gives you a working binary with ffmpeg support — encode, decode, and stream in virtually any format</li>
<li>The formula builds from source in about 2.5 minutes using your system OCaml — nothing persists in your home directory</li>
<li>LADSPA, OSC, Prometheus, and sqlite3 support are candidates for future updates</li>
</ul>
