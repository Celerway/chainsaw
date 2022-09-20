# Changelog

All notable changes to this project will be documented in this file. See [standard-version](https://github.com/conventional-changelog/standard-version) for commit guidelines.

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
