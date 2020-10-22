<div align="center">
  <h1>Daily Redirector</h1>
  <strong>Redirect daily.dev links</strong>
</div>
<br>
<p align="center">
  <a href="https://circleci.com/gh/dailydotdev/daily-redirector">
    <img src="https://img.shields.io/circleci/build/github/dailydotdev/daily-redirector/master.svg" alt="Build Status">
  </a>
  <a href="https://github.com/dailydotdev/daily-redirector/blob/master/LICENSE">
    <img src="https://img.shields.io/github/license/dailydotdev/daily-redirector.svg" alt="License">
  </a>
  <a href="https://stackshare.io/daily/daily">
    <img src="http://img.shields.io/badge/tech-stack-0690fa.svg?style=flat" alt="StackShare">
  </a>
</p>

The redirector service is in charge of redirecting users to the relevant post based on a cutom daily.dev link.
The service also publishes a message upon every view of a human (not a bot) so the ranking system can use it for calculating the score of every post.

## Stack

* Go v1.11
* Go dep managing dependencies.
* `net/http` as the web framework

## Local environment

At the moment it is not possible to run the Redirector without access to a Google Cloud Pub/Sub instance or an emulator.

Daily Redirector requires a running instance of [Dail API](https://github.com/dailydotdev/daily-api), you can set it up by following the [instructions](https://github.com/dailydotdev/daily-api). The service is required for getting the post link.

Make sure to use Go 1.11 and install all dependencies using [dep](https://github.com/golang/dep)

Environment variables:
* `API_URL` - Root url to the API service or the Gateway (if exists in the environment)


## Want to Help?

So you want to contribute to Daily Redirector and make an impact, we are glad to hear it. :heart_eyes:

Before you proceed we have a few guidelines for contribution that will make everything much easier.
We would appreciate if you dedicate the time and read them carefully:
https://github.com/dailydotdev/.github/blob/master/CONTRIBUTING.md
