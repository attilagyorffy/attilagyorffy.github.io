---
title: "Postgres Refuses Your NOT NULL Migration"
subtitle: "Postgres is smarter than your migration, and it is not going to pretend otherwise."
date: 2016-07-25
description: "How to add a NOT NULL column to an existing table in Rails without Postgres losing its mind. Default values, backfills, and the migration nobody warns you about"
summary: "Postgres is smarter than your migration and it's not going to pretend otherwise. Add the column, backfill, then add the constraint."
topics:
  - Ruby on Rails
type: Code
read_time: "3 min read"
footer: "If you have somehow been adding NOT NULL columns to populated tables and wondering why the universe is against you, now you know. Come commiserate on [Bluesky](https://bsky.app/profile/attilagyorffy.com), [Mastodon](https://fosstodon.org/@attila), [~~Twitter~~ X](https://twitter.com/attilagyorffy), or even [LinkedIn](https://linkedin.com/in/attilagyorffy), where I am sure someone will tell me this is what `default:` is for. You can also find my code on [GitHub](https://github.com/attilagyorffy), where the migrations all pass on the first try. Mostly."
---

Right, so here is a scene I have watched play out more times than I care to admit. Someone new to Rails decides they want to add a column to an existing table. Lovely. Very ambitious. They also want a `NOT NULL` constraint on it, because they have heard that data integrity is important, which is technically true, and then they write something like this:

```ruby
class AddFirstAndLastNameToUsers < ActiveRecord::Migration
  def change
    add_column :users, :first_name, :string, null: false
    add_column :users, :last_name, :string, null: false
  end
end
```

And then they run it. And then Postgres tells them, in no uncertain terms, to get stuffed.

You get a `PG::NotNullViolation` because, and I cannot stress this enough, Postgres is not stupid. It looks at your table, sees that there are already rows in it, realises that your shiny new column would have no value in any of those rows, and quite reasonably refuses to let you turn your database into a liar. It has no idea what the value of the new columns should be for the existing rows, and unlike some people, it is not prepared to just wing it.

<mark>Postgres will not let you add a NOT NULL column to a table that already has data, because it refuses to participate in your fantasy that those existing rows can just have nothing in a column that explicitly forbids nothing.</mark>

Which, when you think about it, is actually the database doing its job. The constraint says "this column must always have a value," and you are trying to create it on a table where it immediately would not. That is not a bug. That is the database respecting your own rules more than you do.

## The fix, which is embarrassingly simple

The trick is to break the operation into three steps, all of which can live in a single migration so nobody has to file a support ticket about it. First, you add the column *without* the constraint. Then you backfill the existing rows with some sensible default. Then you add the `NOT NULL` constraint after everything already has a value. It is the database equivalent of putting your trousers on before your shoes.

For example, say you are adding first and last names to your Devise users, because apparently you launched an application where you did not bother collecting anyone's name. Bold choice. Here is how you do it without Postgres throwing a tantrum:

```ruby
class AddFirstAndLastNameToUsers < ActiveRecord::Migration
  def up
    add_column :users, :first_name, :string
    add_column :users, :last_name, :string

    execute <<-SQL.strip_heredoc
      UPDATE users
      SET first_name = '[[UPDATEME]]'
      WHERE first_name IS NULL
    SQL

    execute <<-SQL.strip_heredoc
      UPDATE users
      SET last_name = '[[UPDATEME]]'
      WHERE last_name IS NULL
    SQL

    change_column :users, :first_name, :string, null: false
    change_column :users, :last_name, :string, null: false
  end

  def down
    remove_column :users, :first_name
    remove_column :users, :last_name
  end
end
```

Notice you are using `up` and `down` instead of `change` here, because this is a multi-step migration and Rails cannot magically reverse an `execute` block. If you try to use `change` for this, you deserve whatever happens next.

The `[[UPDATEME]]` placeholder is there to remind you to put in an actual sensible default. Maybe it is an empty string, maybe it is "Unknown," maybe it is the name of your first pet. The point is that every row gets a value before the constraint goes on, so Postgres has nothing to complain about. And Postgres *loves* having nothing to complain about.

Also worth noting: the SQL `WHERE first_name IS NULL` clause means this migration is idempotent in the backfill step. If you somehow end up running it twice, or if some rows already have values because of reasons, it will not clobber them. That is the kind of defensive coding that separates people who sleep at night from people who get paged at 3 AM.

<h2 class="conclusion">The whole point</h2>

This is not complicated. It is barely even interesting. But I keep seeing people get bitten by it, so apparently it needs to be written down somewhere.

The database is not being difficult. It is doing exactly what you asked it to do: enforce your constraints. The fact that your migration contradicts those constraints is a you problem, not a Postgres problem. Add the column first, fill in the blanks, then add the constraint. Three steps. One migration. Zero drama.

<ul class="takeaway">
<li>Never add a <code>NOT NULL</code> column directly to a table that already has rows</li>
<li>Add the column, backfill existing data, then apply the constraint in one migration</li>
<li>Use <code>up</code>/<code>down</code> instead of <code>change</code> when your migration includes raw SQL</li>
</ul>

Do that, and Postgres will stop yelling at you. Or at least it will stop yelling at you about *this*. I make no promises about the rest of your schema.
