# Changelog

All notable changes to this project will be documented in this file.

## [0.3.0] - 2025-11-05

### 🚀 Features

- Prepare should return a changeset [ca90a8c](https://github.com/act3-ai/dagger/commit/ca90a8c60d306d5c43a25f40b551a4b85f4dc3ca) by **Kyle M. Tarplee**, Signed-off-by:Kyle M. Tarplee <kyle.tarplee@udri.udayton.edu>

- Add CreateExtraTags [fdad583](https://github.com/act3-ai/dagger/commit/fdad583795c8da1ae164498cd2455a85fc06568c) by **Kyle M. Tarplee**, Signed-off-by:Kyle M. Tarplee <kyle.tarplee@udri.udayton.edu>


### 🐛 Bug Fixes

- Remove gitRefAsDir() [fa0fa1c](https://github.com/act3-ai/dagger/commit/fa0fa1c4cf8eee88159705875e9977d8b53420b8) by **Kyle M. Tarplee**, Signed-off-by:Kyle M. Tarplee <kyle.tarplee@udri.udayton.edu>

- ExtraTags should not care if the target tag already exists [5220c91](https://github.com/act3-ai/dagger/commit/5220c917cf3e54d98dc8725bc51b9ed09966a751) by **Kyle M. Tarplee**, Signed-off-by:Kyle M. Tarplee <kyle.tarplee@udri.udayton.edu>

- Correctly parse OCI references in CreateExtraTags [02835a8](https://github.com/act3-ai/dagger/commit/02835a880f48b49755507bef0a5694a3e315d150) by **Kyle M. Tarplee**, Signed-off-by:Kyle M. Tarplee <kyle.tarplee@udri.udayton.edu>

- Upgrade dagger to v0.19.4 [f68a739](https://github.com/act3-ai/dagger/commit/f68a7390c6800f0dda3d10fd1f9a34a9c7eb182c) by **Paul Spurlock**

- Bug with optional config path, add version checking to BumpedVersion [f0a748f](https://github.com/act3-ai/dagger/commit/f0a748f50083bb10c1f84c86a68cd14e42e0042f) by **Paul Spurlock**

- Add cliff.toml for git-cliff [df83b2f](https://github.com/act3-ai/dagger/commit/df83b2fb9f7226fcd293a85086fd64ece9b29059) by **Paul Spurlock**

- More error log handling changes in bumped-version [8a1cf97](https://github.com/act3-ai/dagger/commit/8a1cf97f1e9df38f3ef0f0fa0ecf7f9423265f36) by **Paul Spurlock**

- Tests [cf0f37e](https://github.com/act3-ai/dagger/commit/cf0f37e026ecd09d2d67b16eb0c55e4b80d27d56) by **Paul Spurlock**

- Move config out of New and into functions instead [41ca0d4](https://github.com/act3-ai/dagger/commit/41ca0d43bdb3be7f5adb002fe145c0bd9ffd1e94) by **Paul Spurlock**

- Test commit [11b383f](https://github.com/act3-ai/dagger/commit/11b383f84ef3cbbc59df1949a4acb532f8a6e505) by **Paul Spurlock**

- Upgrade git-cliff to v0.2.2 [a60d5a5](https://github.com/act3-ai/dagger/commit/a60d5a5f0739fcc3b891909e58501cdeb9422271) by **Paul Spurlock**

- Add cliff.toml to git-cliff [fb81c7c](https://github.com/act3-ai/dagger/commit/fb81c7c91f3af6afb1e249db6a9196c54cf116f0) by **Paul Spurlock**

- Tests [276a94e](https://github.com/act3-ai/dagger/commit/276a94e6ccd7c8b47a5917efe4db247d4411fe14) by **Paul Spurlock**

- Release notes name was wrong [46dbf44](https://github.com/act3-ai/dagger/commit/46dbf4493c6d7e47d8b501e2aaecc49f553b9bd3) by **Kyle M. Tarplee**, Signed-off-by:Kyle M. Tarplee <kyle.tarplee@udri.udayton.edu>

- Remove unused dependencies [7601391](https://github.com/act3-ai/dagger/commit/7601391a0821f2b180ffeaef52e1a0ad6e7a4300) by **Paul Spurlock**


### 🚜 Refactor

- Move tag related function into tags.go [4ec14ef](https://github.com/act3-ai/dagger/commit/4ec14efa9f17e0a9d8d0591c60b858c90c82fad4) by **Kyle M. Tarplee**, Signed-off-by:Kyle M. Tarplee <kyle.tarplee@udri.udayton.edu>

- Prepare.go, now accepts optional git-cliff config and token [5489cba](https://github.com/act3-ai/dagger/commit/5489cba3bd9366634c3c23830f6441538569a9e7) by **Paul Spurlock**


## [release/v0.2.3] - 2025-10-24

### 🐛 Bug Fixes

- Update git-cliff to v0.2.1 [3994eec](https://github.com/act3-ai/dagger/commit/3994eeca1df882ca96bf5d7345baa92c4edfd9b7) by **Paul Spurlock**

- Improve error handling when git-cliff bump fails to bump version [b12ed52](https://github.com/act3-ai/dagger/commit/b12ed52d00a84ac0b053bc92ac573d65b1f840c0) by **Paul Spurlock**

- Add test for prepare [5f8231a](https://github.com/act3-ai/dagger/commit/5f8231a286aa4b5256f081047b625966e099a19f) by **Paul Spurlock**

- Remove gitStatus check, no longer needed since gitref is the only accepted source [c19e219](https://github.com/act3-ai/dagger/commit/c19e2194b1c6be02daf219b257960c06a8b26021) by **Paul Spurlock**


## [release/v0.2.2] - 2025-10-22

### 🐛 Bug Fixes

- Update dependency act3-ai/dagger to git-cliff/v0.2.0 [6248f0f](https://github.com/act3-ai/dagger/commit/6248f0f4caff0ff5103b946e3c29b2a1df321a85) by **Paul Spurlock**

- Update git-cliff functions to use gitRef instead of src [7ad89bf](https://github.com/act3-ai/dagger/commit/7ad89bf2aec8a8f73002bff76c7989e47082c09d) by **Paul Spurlock**


## [release/v0.2.1] - 2025-10-20

### 🐛 Bug Fixes

- Upgrade dagger engine to v0.19.2 [0a98ad4](https://github.com/act3-ai/dagger/commit/0a98ad41e05f1831f16a61f2c072424ffccb9ce4) by **Paul Spurlock**

- Remove gitignore flag [13b62be](https://github.com/act3-ai/dagger/commit/13b62be9d700ff9acab941c61f4efee35f260533) by **Paul Spurlock**


## [release/v0.2.0] - 2025-10-08

### 🚀 Features

- Switch to using a git ref instead of local directory for the source [ffe7f21](https://github.com/act3-ai/dagger/commit/ffe7f21a643b125da43e3c7c2d44373c50aec4e4) by **Paul Spurlock**

BREAKING CHANGE: `--src` has been changed to `--gitref` and only accepts a !GitRef instead of a !Directory.


### 🐛 Bug Fixes

- Remove linting functions from release module [5349a55](https://github.com/act3-ai/dagger/commit/5349a553bfb9345d740d913876de118d3b4c35eb) by **Paul Spurlock**

- Ineffective usage of title option in CreateGitlab [8ce6fc6](https://github.com/act3-ai/dagger/commit/8ce6fc64829263db8bbe39a9d6c787ea82e0b0ec) by **nathan-joslin**

- Add check for if version bump already exists [72977e7](https://github.com/act3-ai/dagger/commit/72977e7d5f0f273184b1c7d2df31a725460520b5) by **Paul Spurlock**


### 💼 Other

- Bump yamllint to v0.1.5 [d1f39f7](https://github.com/act3-ai/dagger/commit/d1f39f7a92ac29015a7a03d1e06dfd0481e78528) by **nathan-joslin**


## [release/v0.1.13] - 2025-09-23

### 💼 Other

- Bump yamllint to v0.1.5 [e0206c1](https://github.com/act3-ai/dagger/commit/e0206c1e4d4aada68aaaea186a42ba88e8f08a60) by **nathan-joslin**


## [release/v0.1.12] - 2025-09-19

### 🐛 Bug Fixes

- Update dagger enving to v0.18.19 [fa5e287](https://github.com/act3-ai/dagger/commit/fa5e287957879c806f5bbc03bda8a2cd29ddf8cb) by **Paul Spurlock**

- Update git-cliff to v0.1.4 [4d3bb34](https://github.com/act3-ai/dagger/commit/4d3bb349ed0883ea5038321de6c8f5347f8796d2) by **Paul Spurlock**

- Updae govulncheck to v0.1.4 [a701714](https://github.com/act3-ai/dagger/commit/a701714886d48e42474e075216b075625b199f38) by **Paul Spurlock**

- Update markdownlint to v0.1.3 [058198e](https://github.com/act3-ai/dagger/commit/058198e576dc88c0a24e7298db3c4c329c8db0f1) by **Paul Spurlock**

- Update python to v0.1.6 [5c405f8](https://github.com/act3-ai/dagger/commit/5c405f8a239c615170bb33b40a6764713045a1c7) by **Paul Spurlock**

- Update yamllint to v0.1.4 [10dcae5](https://github.com/act3-ai/dagger/commit/10dcae53bc9c0e8e39ee91abe4b7855ee9f7ef0a) by **Paul Spurlock**

- Update wolfi to v0.18.19 [0eab6b9](https://github.com/act3-ai/dagger/commit/0eab6b9fb9572cd1d641b80c7d53df76d32f9eac) by **Paul Spurlock**

- Update shellcheck to v0.18.19 [20c792d](https://github.com/act3-ai/dagger/commit/20c792d101de7d9b2a4936c4d4b1cb33325e044a) by **Paul Spurlock**

- Update golangci-lint [c6bcdd7](https://github.com/act3-ai/dagger/commit/c6bcdd72e2ea214b343bdb7eca49f2a7d60831d2) by **Paul Spurlock**

- Update registry-config [6989c07](https://github.com/act3-ai/dagger/commit/6989c07070009915c48abfbe2e95a33135605341) by **Paul Spurlock**

- Update go [38b1216](https://github.com/act3-ai/dagger/commit/38b12168f7fff8b3531b5f0e1946817b494aeb98) by **Paul Spurlock**

- Update gh [f78bc89](https://github.com/act3-ai/dagger/commit/f78bc899df502f1b8a363b7ee73c52a9220e70fa) by **Paul Spurlock**


## [release/v0.1.11] - 2025-09-18

### 🐛 Bug Fixes

- create new changelog if one is not found at changelogPath [a1c8c6d](https://github.com/act3-ai/dagger/commit/0ef49970376756f9af1f0dd604d04906aa1c8c6d) by **Paul Spurlock**


## [release/v0.1.10] - 2025-08-19

### 🐛 Bug Fixes

- Upgrade dagger to v0.18.16 [7fc7c7e](https://github.com/act3-ai/dagger/commit/7fc7c7ed3d9abeb42c9f8ebfa611998ac18ef427) by **Paul Spurlock**

- Update git-cliff to v0.1.3 [0db1f06](https://github.com/act3-ai/dagger/commit/0db1f065ce7543f06918830eec52e93d22d9a49a) by **Paul Spurlock**

- Update govulncheck to v0.1.3 [04b97de](https://github.com/act3-ai/dagger/commit/04b97de5a5ef2aa3d88fa8cfe889e3e4f452f721) by **Paul Spurlock**

- Update markdownlint to v0.1.2 [11a121d](https://github.com/act3-ai/dagger/commit/11a121d169bd02bcefe4338ca7009b6c7230e4e8) by **Paul Spurlock**

- Update python to v0.1.5 [ec9e915](https://github.com/act3-ai/dagger/commit/ec9e9155962f0ce2366debce56a97f391a2afa13) by **Paul Spurlock**

- Update yamllint to v0.1.3 [5685746](https://github.com/act3-ai/dagger/commit/5685746967f99737b8b89b9bb1daadf9a4cab45e) by **Paul Spurlock**


## [release/v0.1.9] - 2025-07-08

### 🐛 Bug Fixes

- Update dagger to v0.18.12 [7313af8](https://github.com/act3-ai/dagger/commit/7313af897d78c3b6ff0003e8d07ab428066c06ee) by **Paul Spurlock**

- Update python module to v0.1.3 [c9e510c](https://github.com/act3-ai/dagger/commit/c9e510c54c63cd6c7a65779416fbe71a7e4b3bc8) by **Paul Spurlock**


## [release/v0.1.8] - 2025-07-03

### 🐛 Bug Fixes

- Prepare extra notes whitespace [3761187](https://github.com/act3-ai/dagger/commit/376118780091241b32c65d6c46343b2be679fbab) by @nathan-joslin


## [release/v0.1.7] - 2025-07-02

### 🐛 Bug Fixes

- Add support for additional .gitignore file [431a422](https://github.com/act3-ai/dagger/commit/431a422d079793b3ea7d3b28f5e939b63b16a912) by @nathan-joslin


## [release/v0.1.6] - 2025-07-02

### 🐛 Bug Fixes

- Update go verify to return an output string to propagatre warnings [6a68d9c](https://github.com/act3-ai/dagger/commit/6a68d9c31374f6baa3b3f42ff570b5c09a6054db) by @nathan-joslin


## [release/v0.1.5] - 2025-07-02

### 🐛 Bug Fixes

- Use fork of golang.org/x/exp/cmd/gorelease [8ba7fa7](https://github.com/act3-ai/dagger/commit/8ba7fa7b4ade369a9d3910efb52d31922210ab2f) by @nathan-joslin


## [release/v0.1.4] - 2025-06-27

### 🐛 Bug Fixes

- Parallelize linters in golang check and python check [1e7499d](https://github.com/act3-ai/dagger/commit/1e7499de32cf85d41cc4d0ec5e5b668d6d3915a3) by @nathan-joslin


## [release/v0.1.3] - 2025-06-25

### 🐛 Bug Fixes

- Handling of v version prefixes and whitespace [5893094](https://github.com/act3-ai/dagger/commit/589309438d7c27fa812fe3b47f26b1a17ca4b43d) by **nathan-joslin**

- Remove unused option to disable unit tests [aaafeb3](https://github.com/act3-ai/dagger/commit/aaafeb39645dac52e4ed105b6ed1290480f75ac0) by **nathan-joslin**


### 💼 Other

- Bump dagger engine version v0.18.10 to v0.18.11 [4acb5ff](https://github.com/act3-ai/dagger/commit/4acb5ff5b0cf8806206c49fa0bbd264800fb82ea) by **nathan-joslin**

- Bump module git-cliff v0.1.1 to v0.1.2 [a54ffe4](https://github.com/act3-ai/dagger/commit/a54ffe4e0383f1a419165ba23f2582fac0f62a16) by **nathan-joslin**

- Bump module govulncheck v0.1.1 to v0.1.2 [f646cb5](https://github.com/act3-ai/dagger/commit/f646cb56c72fa816ee7e22cf61f043eaa019b741) by **nathan-joslin**

- Bump module markdownlint v0.1.0 to v0.1.1 [1e9cd9e](https://github.com/act3-ai/dagger/commit/1e9cd9eda71cb0d36d6edd8439641db9985e04c6) by **nathan-joslin**

- Bump module python v0.1.1 to v0.1.2 [ea35619](https://github.com/act3-ai/dagger/commit/ea3561906d11d3da50d463219fef6e285b25b590) by **nathan-joslin**

- Bump module yamllint v0.1.1 to v0.1.2 [ed0b052](https://github.com/act3-ai/dagger/commit/ed0b052ff2881875d849db9c39bba95f4e59d9a4) by **nathan-joslin**

- Bump module shellcheck v0.18.10 to v0.18.11 [cba2af7](https://github.com/act3-ai/dagger/commit/cba2af7f52ca32eabb17b3c72bd2aa721aea4c3d) by **nathan-joslin**

- Bump module wolfi v0.18.10 to v0.18.11 [76c44cc](https://github.com/act3-ai/dagger/commit/76c44cc7e8368b1e0e759f2d78801859643bdaaa) by **nathan-joslin**


### 📚 Documentation

- Fix typo in long description [1425169](https://github.com/act3-ai/dagger/commit/14251695031b5d471c392b905bd45cddc454f519) by **nathan-joslin**


## [release/v0.1.2] - 2025-06-24

### 🐛 Bug Fixes

- *(release)* Update dagger module deps

## [release/v0.1.1] - 2025-06-18

### 🐛 Bug Fixes

- *(release)* Use tagged versions of act3-ai modules

### 💼 Other

- *(release)* Bump dagger/modules/shellcheck@v0.18.2 to v0.18.10
- *(release)* Bump dagger/modules/wolfi@v0.18.5 to v0.18.10

## [release/v0.1.0] - 2025-06-18

🚀 Initial release 🚀
