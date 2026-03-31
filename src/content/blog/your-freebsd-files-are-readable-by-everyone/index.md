---
title: "Your FreeBSD Files Are Readable by Everyone"
subtitle: "FreeBSD's default umask lets every user on the system read every new file. Here is why that matters and how to fix it."
date: 2016-01-29
description: "FreeBSD ships with a default umask that lets every user on the system read every new file. Here is why that matters and how to tighten it up properly."
summary: "Your database credentials, API keys, and deployment scripts are world-readable by default. Three digits in one config file fix it. You're welcome."
topics:
  - FreeBSD
  - Security
type: Code
read_time: "3 min read"
footer: "If you enjoyed this little tour of FreeBSD permissions and want to chat about server hardening, find me on [Bluesky](https://bsky.app/profile/attilagyorffy.com), [Mastodon](https://fosstodon.org/@attila), [~~Twitter~~ X](https://twitter.com/attilagyorffy), or even [LinkedIn](https://linkedin.com/in/attilagyorffy) if you're feeling professional about it. The rest of my tinkering lives on [GitHub](https://github.com/attilagyorffy)."
---

<mark>Every file you create on a stock FreeBSD system is readable by every other user on that machine</mark>. Your database credentials, your API keys, your private config -- all of it just sitting there with `-rw-r--r--` permissions like a diary left open on a park bench. On a shared server, that is not a quirky default. It is the operating system actively working against you.

Right, quick refresher for those of you who blocked this out after university. UNIX file permissions come in three flavours: the owning user, the owning group, and everyone else (charmingly called "others" or "world," because apparently the whole world deserves to read your production secrets). When you create a file, the system decides what permissions it gets based on something called the umask. Here is what the default looks like:

```bash
$ touch test.txt
$ ls -la test.txt
-rw-r--r--  1 vagrant  vagrant  0 Jan 28 22:53 test.txt
```

The permissions `-rw-r--r--` mean the owner can read and write, and literally everyone else on the machine can read it too. For some throwaway temp file, sure, who cares. But think about what you actually create on a server: application configs, log files, database dumps, deployment scripts. All of it is world-readable by default. It is like leaving your front door open and calling it a feature.

Now, you could manually `chmod` every single file after you create it, but let's be honest -- you won't. You will forget by the second file. Your scripts definitely will not do it. What you actually need is to fix the default so you stop shooting yourself in the foot. That is what umask does.

## What is umask

Umask stands for "user file-creation mask," which sounds like something a sysadmin made up to feel important, but it is genuinely useful. Every process in a POSIX environment carries a umask value that specifies which permission bits to *remove* from newly created files. Think of it as a bouncer for your filesystem: the system starts with the maximum permissions a program asks for, and the umask strips away whatever you have told it to refuse at the door.

The default umask in FreeBSD is `022`. Because of course it is. Each octal digit maps to a permission scope:

- 0: no permissions are removed from the owner
- 2: write permission is removed for the group
- 2: write permission is removed for others

So the group and everyone else can still read your files. They just cannot modify them. Great, so strangers can look through your stuff but not rearrange the furniture. How generous. On a multi-user server this is hilariously permissive.

## Setting a stricter default

A umask of `027` is what a reasonable person would have chosen in the first place. It gives the owner full access, the group read-only access (handy for services that share a group), and tells everyone else to get stuffed. No world-readable files, no accidental data leaks to random unprivileged users snooping around on your box.

FreeBSD keeps this little gem tucked away in `/etc/login.conf`:

```bash
default:\
        :umask=022:
```

Swap `022` for `027` in your editor, save it, and then -- because nothing on FreeBSD can ever just work without a second step -- rebuild the login capability database:

```bash
$ sudo cap_mkdb /etc/login.conf
```

Fair warning: the new umask only kicks in for new login sessions. Anything already running keeps its old, recklessly generous umask until you log out and back in. Yes, you have to turn it off and on again.

Once you have done the sacred log-out-log-in dance, verify it actually worked:

```bash
$ touch test2.txt
$ ls -la test2.txt
-rw-r-----  1 vagrant  vagrant  0 Jan 28 23:19 test2.txt
```

Look at that -- the `others` column is blissfully empty. Randos on your system can no longer casually browse your files like it is a public library. If you are feeling particularly paranoid, you can go nuclear with `077` and kill group access too, but `027` strikes a decent balance between security and the practical reality that services often need group-level read access to not fall over.

<ul class="takeaway">
<li>Change the default umask from <code>022</code> to <code>027</code> in <code>/etc/login.conf</code></li>
<li>Rebuild the login database with <code>cap_mkdb /etc/login.conf</code></li>
<li>New files will deny all access to others — existing files are unaffected</li>
</ul>
