How to contribute
=================

.. note::

    Pure feature requests will currently **NOT BE** considered.
    Read on for the reasoning and details.

This software is still in very early stages and still needs to find the
direction where it's usable for a high number of people. Implementing features
that only a very limited number of users will benefit from is one of the
highest risk currently. Since we also want to make sure that the feature set of
»brig« makes sense as a whole and all features are orthogonal, we will ignore
typical feature request at the moment.

What we want instead are *experience reports*. We want you to use the current state
of the software and write down the following:

- Was it easy to get »brig« running?
- Was it easy to understand it's concepts?
- What is your intended usecase for it? Could you make it work?
- If no, what's missing in your opinion to make the usecase possible?
- Anything else that you feel free to share.

After this we'll try to analyze your reports and create feature requests
acccordingly. Once those were implemented we will probably allow traditional
feature requests to be made.

**What do we want to prevent by this?** Getting 20 feature requests, 5 of them
contradicting each other and. In short: featuritis. It's quite hard to figure
out what users wants from a developers standpoint. So this will hopefully give
us some more insights.

**Are bug reports okay?** Sure. If you already fix the bug it's even better.
Please use the ``brig bug`` command to get a template with all the info we need.

**Are very small feature requests okay?** If it's only about changing or
extending an existing feature, it's probably fine. Feel free to create an issue
on GitHub to check back on this before you do any actual change.

Also, the developer of this software is currently doing all of this is in his
free time. If you're willing to offer any financial support feel free to
contact me.

What to improve
---------------

The following improvements are greatly appreciated:

- Bug reports & fixes.
- Documentation improvements.
- Porting to other platforms.
- Writing tests.

Workflow
--------

Please adhere to the general `GitHub workflow_`, i.e. fork the repository,
make your changes and open a pull request that can be discussed.

.. _`Github workflow`: https://help.github.com/articles/about-pull-requests

If you contribute code, make sure:

- Tests are still running and you wrote test for your new code.
- You ran ``gofmt`` over your code.
- Your pull requests is opended against the ``develop`` branch.
