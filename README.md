# `git-split`, an interactive commit-splitting tool for git

It's generally better in a code review that individual commits be focused to complete exactly one
logical action. The reviewer doesn't have to maintain and switch contexts, and reviews are generally
much more accurate and faster.

But, often development happens non-linearly, and even when it does manage to happen in a linear
fashion, it's easy to create something monolithic before you remember to create a new commit.

`git-split` means to address that by allowing you to split individual commits into multiple commits
interactively, with a simple file-chunk-line tree view and checkbox selection. Simply select what
you want to be in the first and subsequent commits with each iteration. Once there is no difference
between `HEAD` and your target commit, `git-split` exits and rebases the branch you were previously
in on top of the tip of the split commits.

Something went wrong? Don't like the result? `git-split` also creates a backup of the old branch as
`git-split-backup/<branchname>[.#]`, appending an incrementing number if the same branch is split
multiple times.

## Can't you do this with git-rebase?
The official documentation for splitting commits in git is something like as follows:

1. Start an interactive rebase (`git rebase -i/--interactive`) and change the action on the commit
   you wish to split to `e/edit`
2. Do a soft reset to the commit's target branch parent (`git reset HEAD^`), which leaves the target
   commit's changes unstaged.
3. Edit the files, likely through the `git add -i/--interactive` or `-p/--patch` functionality,
   which sequentially iterates through chunks, choosing what to do with each chunk. If you want to
   edit less than a chunk, you then enter an edit mode for the chunk before continuing.

So the short answer is: yes, but as with most git things, it's harder than it needs to be and could
easily result in wasting time fixing your broken commits or branch through `reflog` or such,
especially if you didn't create a backup branch on the old tip.