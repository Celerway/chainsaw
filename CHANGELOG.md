# Changelog

All notable changes to this project will be documented in this file. See [standard-version](https://github.com/conventional-changelog/standard-version) for commit guidelines.

### [1.3.1](https://github.com/Celerway/chainsaw/compare/v1.3.0...v1.3.1) (2023-03-21)


### Bug Fixes

* make the TestRemoveWriter use a fresh logger. ([4f36d0f](https://github.com/Celerway/chainsaw/commit/4f36d0f6dbbe3ea2164f50c649c1acea34140038))
* set the default print level to whatever the default logger has. ([7477204](https://github.com/Celerway/chainsaw/commit/7477204c9f7c3407513ad33e8b625ed630aa99eb))

## [1.3.0](https://github.com/Celerway/chainsaw/compare/v1.2.1...v1.3.0) (2022-11-25)


### Features

* add Print and Printf support so we can pass this to APIs which expect the stdlib logger. ([dd425e7](https://github.com/Celerway/chainsaw/commit/dd425e7007208f66dfa14246a10a3ad1abf1aea1))

### [1.2.1](https://github.com/Celerway/chainsaw/compare/v1.2.0...v1.2.1) (2022-11-18)


### Bug Fixes

* handle logging calls when the logger is nil. ([b524fab](https://github.com/Celerway/chainsaw/commit/b524fab18ec811bc11d9867426a8d8d9bbcdcb3f))

## [1.2.0](https://github.com/Celerway/chainsaw/compare/v1.1.5...v1.2.0) (2022-10-27)


### Features

* add logger name to the backtrace begin and end fields. ([b7d47db](https://github.com/Celerway/chainsaw/commit/b7d47db9c4fe119f69c05168309a17b60d3d4bd7))


### Bug Fixes

* unit test for backtrace ([fccee8d](https://github.com/Celerway/chainsaw/commit/fccee8dd924f8bae9aec48e67f0a5ca0b432a838))

### [1.1.5](https://github.com/Celerway/chainsaw/compare/v1.1.4...v1.1.5) (2022-09-22)


### Bug Fixes

* fields where duplicated. fix and add tests for fields usage. ([f02c186](https://github.com/Celerway/chainsaw/commit/f02c1869de96c2d094eff7f4a22a2cfc4cc9766f))

### [1.1.4](https://github.com/Celerway/chainsaw/compare/v1.1.3...v1.1.4) (2022-09-22)


### Bug Fixes

* use short form of loglevel in the logs. So InfoLevel ==> info, etc. ([d5b7147](https://github.com/Celerway/chainsaw/commit/d5b7147c82e8ea4bdcb38816b47e344dc23300eb))

### [1.1.3](https://github.com/Celerway/chainsaw/compare/v1.1.2...v1.1.3) (2022-09-22)


### Bug Fixes

* remove erroneous whitespace in front of each line. ([f811860](https://github.com/Celerway/chainsaw/commit/f8118605974bc25a2d23c47784292492cfec45e0))

### [1.1.2](https://github.com/Celerway/chainsaw/compare/v1.1.1...v1.1.2) (2022-09-20)


### Bug Fixes

* fields from the logger were not added to messages. ([8611f0f](https://github.com/Celerway/chainsaw/commit/8611f0f5917d574d7c5f136f0476db33de4f4146))

### [1.1.1](https://github.com/Celerway/chainsaw/compare/v1.1.0...v1.1.1) (2022-09-20)


### Bug Fixes

* replace stdout with stderr, as loggers should do. This was a bit late, I know. ([2111eba](https://github.com/Celerway/chainsaw/commit/2111eba2fc48a270034792a34f03394acbcf10c6))

## 1.1.0 (2022-09-19)


### Features

* add support for structured logging. ([008b720](https://github.com/Celerway/chainsaw/commit/008b7204929c647727dbf33bab8244300b53bed6))


### Bug Fixes

* add backtrace support ([079a238](https://github.com/Celerway/chainsaw/commit/079a23830e3f156a92839476bb9206502d4d8350))
* add new mock writer that allows us to see what writes are being done against it. ([3ba2b74](https://github.com/Celerway/chainsaw/commit/3ba2b74f9af479b241fb943488655ef619690954))
* check if the stream output is being serviced. if a stream doesn't accept a log message within a second then that channel gets removed. ([00f96d8](https://github.com/Celerway/chainsaw/commit/00f96d8cba1eac0a8f14bea92d13d841e52b9c56))
* remove bool and running flag to simplify. ([9341bad](https://github.com/Celerway/chainsaw/commit/9341bad081c23d83bb085e905a7b4c66d7575eae))
