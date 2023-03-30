# How to contribute

## Welcome

Welcome to HorizonCD, and thanks for your interest in contributing to us!

Below is a set of things you should know before you open a PR.

## Issues

Opening an issue is a great way to contribute to this project.

Before opening an issue, you should always check if there's one containing your good thought exists.

If there's feedback\comments or anything you are not sure about, feel free to open an issue, whenever day and night.

To make the issue details as standard as possible, we setup an [ISSUE TEMPLATE](https://github.com/horizoncd/horizon/tree/main/.github/ISSUE_TEMPLATE) for issue reporters. You can find two kinds of issue templates there: bug report and feature request. Please BE SURE to follow the instructions to fill fields in template.
An issue about bugs should tell us:

There are a lot of cases when you could open an issue:

+ bug report
+ feature request
+ feature proposal
+ feature design
+ help wanted
+ doc incomplete
+ test improvement
+ any questions on project
+ and so on
> We must remind that when filing a new issue, please remember to remove the sensitive data from your post. Sensitive data could be password, secret key, network locations, private business data and so on.
## Commit Rules
To better comply with Angular specifications, it is recommended that you make a habit of committing code without the `git commit -m` option, Instead, directly use `git commit` or `git commit -a` to enter the interactive interface and edit the Commit Message. This will better format the Commit Message.

You might think that code commits are heavy and seem arbitrary. Or maybe we want to wait until we've built a complete feature and commit it together in a commit. In this case, we can perform `git rebase -i` commits before the final merge or Pull Request.
=======
We are eager to discuss with you.
## Pull Requests
You'll need to sign off your commits, before we could accept your pull requests.

Pull requests should always get an associated issue. Before making a PR, please open an issue to discuss with us, or you are at a risk of spending lots of time on a PR that we don't need.


### Code review

We recommend you describe what your PR do in the PR's comment box, that will save a lot of time for you and us.

Before you call someone to review your code, check the [CI](https://github.com/horizoncd/horizon/actions) is passed.

### CI

We host CI on [GitHub Actions](https://github.com/horizoncd/horizon/actions), we will make sure PR pass tests before we can merge it.

These two kind of tests: `lint` and `unit test`

`lint` tests if your code matches our code conventions, please consult [golangci-lint](https://golangci-lint.run/) and [lint config](./.golangci.yml)

`unit test` runs all the test in code, and the code coverage should not less than 70 percent.

Try your best to keep every function has been tested, it keeps the function behaves as intended.

### Developing

The [development guide](DEVELOPMENT.md) could help you easily develop and debug the features you are interested in.