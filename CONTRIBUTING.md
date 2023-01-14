# How to contribute

## Welcome

Welcome to HorizonCD and thanks for your interest in contributing to us!

Below is a set of things you should know before you open a PR.

## Issues

Opening an issue is a great way to contribute to this project.

Before opening an issue, you should always check if there's one containing your good thought exists.

If there's feedback\comments or anything you are not sure about, feel free to open an issue, whenever day and night.

An issue about bugs should tells us:

    * What do you want to achieve.
    * What is the expected result.
    * What are the steps to reproduce this bug.
    * What environment you are using.

An issue about new features should tells us:

    * Senario the new feature works in.

We are eager discuss with you.

## Pull Requests

You'll need to sign a [Contributer License Agreement(CLA)](./CLA.md), before we could accept your pull requests.

Pull requests should always get an associated issue. Before making a PR, please open an issue to discuss with us, or you are at a risk of spending lots of time on a PR that we don't need.

### Code Style

This project adheres to coding conventions [`Effective Go`](https://go.dev/doc/effective_go), please make sure your code follows it either.

### Code review

We recommend you describe what your PR do in the PR's comment box, that's will save a lot of time for you and us.

Before you call someone to reivew your code, check the [CI](https://github.com/horizoncd/horizon/actions) is passed.

### CI

We host CI on [github actions](https://github.com/horizoncd/horizon/actions), we will make sure PR pass tests before we can merge it.

These two kind of tests: `lint` and `unit test`

`lint` tests if your code matches our code conventions, please consult [golangci-lint](https://golangci-lint.run/) and [lint config](./.golangci.yml)

`unit test` runs all the test in code, and the code coverage should not less than 70 percent.

Try your best to keep every function has been tested, it keeps the funtion behaves as intended.
