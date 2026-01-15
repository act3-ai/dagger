# Changelog

All notable changes to this project will be documented in this file.

## [0.2.8] - 2026-01-15

### 🐛 Bug Fixes

- Updating dagger engine to v0.19.10 [cdb0db5](https://github.com/act3-ai/dagger/commit/cdb0db51f7c8d3568c2785de1bd5a39af5aad37e) by @pspurlock

- Updating dagger engine to v0.19.10 in tests [3909a06](https://github.com/act3-ai/dagger/commit/3909a06c0d0356147cfce99d2f57f23c3a0f0f16) by @pspurlock


## [0.2.7] - 2025-12-16

### 🐛 Bug Fixes

- Updating dagger engine to v0.19.8 [53dc2a3](https://github.com/act3-ai/dagger/commit/53dc2a39fe3fe72d525e6f312250e6932276ba07) by @pspurlock


## [0.2.6] - 2025-12-12

### 🐛 Bug Fixes

- Switch tests to checks [42ce61d](https://github.com/act3-ai/dagger/commit/42ce61df01cb8bce197fbd916d62aabcb6608a24) by @pspurlock

- Disable default caching to fix gitRef caching bug [244456a](https://github.com/act3-ai/dagger/commit/244456ae8ec1a9345ef4ca7c3e118d4b4e997f36) by @pspurlock

- Add +cache=never to fix gitref caching bug [f530608](https://github.com/act3-ai/dagger/commit/f530608d97104fb6fb51bf5bd9674c1f0a111eb4) by @pspurlock


## [0.2.5] - 2025-12-03

### 🐛 Bug Fixes

- General cleanup [03cc59f](https://github.com/act3-ai/dagger/commit/03cc59f49992a4ba8620795bd3e73b19fc8090fb) by @ktarplee, Signed-off-by:Kyle M. Tarplee <kyle.tarplee@udri.udayton.edu>

- Bug with workdir when using WithMountedDirectory [2987b96](https://github.com/act3-ai/dagger/commit/2987b969853ebedb867d4c014b7d426289423ab4) by @pspurlock


## [0.2.4] - 2025-11-20

### 🐛 Bug Fixes

- Updating dagger engine to v0.19.6 [6cb6c6c](https://github.com/act3-ai/dagger/commit/6cb6c6c9844abe06de382ddbaee06a2bd8d68be1) by @pspurlock


## [0.2.3] - 2025-11-20

### 🐛 Bug Fixes

- Upgrade dagger to v0.19.4 [f68a739](https://github.com/act3-ai/dagger/commit/f68a7390c6800f0dda3d10fd1f9a34a9c7eb182c) by @pspurlock

- Bug with optional config path, add version checking to BumpedVersion [f0a748f](https://github.com/act3-ai/dagger/commit/f0a748f50083bb10c1f84c86a68cd14e42e0042f) by @pspurlock

- Add cliff.toml for git-cliff [df83b2f](https://github.com/act3-ai/dagger/commit/df83b2fb9f7226fcd293a85086fd64ece9b29059) by @pspurlock

- More error log handling changes in bumped-version [8a1cf97](https://github.com/act3-ai/dagger/commit/8a1cf97f1e9df38f3ef0f0fa0ecf7f9423265f36) by @pspurlock

- Tests [cf0f37e](https://github.com/act3-ai/dagger/commit/cf0f37e026ecd09d2d67b16eb0c55e4b80d27d56) by @pspurlock

- Move config out of New and into functions instead [41ca0d4](https://github.com/act3-ai/dagger/commit/41ca0d43bdb3be7f5adb002fe145c0bd9ffd1e94) by @pspurlock

- Upgrade dagger engine v0.19.5 [701a04f](https://github.com/act3-ai/dagger/commit/701a04fe098c496d733464d2032f2b42eea5aaff) by @pspurlock

- Tests for refactor and upgrade dagger engine to v0.19.5 in tests [fd949c8](https://github.com/act3-ai/dagger/commit/fd949c8e547a12226a7b8c0e5a5e76114de128c4) by @pspurlock

- Trim space when returning version as a string [f0c90a6](https://github.com/act3-ai/dagger/commit/f0c90a67ec38d5468a70c4691ffc936da000c525) by @pspurlock

- Use workdir instead of config [453b21c](https://github.com/act3-ai/dagger/commit/453b21c7db256898d498f11b2cd798c7748efab8) by @pspurlock


### 🚜 Refactor

- Gitcliff to accept GitCliffOpts instead in conjuction with Run [d7445c5](https://github.com/act3-ai/dagger/commit/d7445c5dac7af1c60f6bb113cf95ad2262c31eae) by @pspurlock

- Create generate changelog and notes functions [7c1937b](https://github.com/act3-ai/dagger/commit/7c1937b62c5c9ed988931d044c948eb910cea1df) by @pspurlock

- Change --gitref to --git-ref, add extranotes to releasenotes [31dfc80](https://github.com/act3-ai/dagger/commit/31dfc80d0e22ec97f4068bda984d05717c556ef9) by @pspurlock


## [git-cliff/v0.2.2] - 2025-11-05

### 🐛 Bug Fixes

- Upgrade dagger to v0.19.4 [56794ad](https://github.com/act3-ai/dagger/commit/56794ad9a351d580bd35e2e30a1196c3e10616c4) by @pspurlock

- Bug with optional config path, add version checking to BumpedVersion [1eff2a6](https://github.com/act3-ai/dagger/commit/1eff2a69c85934080337171bc01adc71a7adeaea) by @pspurlock

- Add cliff.toml for git-cliff [0a868fb](https://github.com/act3-ai/dagger/commit/0a868fb2dfff1eb4e2bb8ef303afe2b7c203ca57) by @pspurlock

- More error log handling changes in bumped-version [a3388c4](https://github.com/act3-ai/dagger/commit/a3388c4210379f29d85f8c79b798a737295c8700) by @pspurlock

- Tests [6b5cbd2](https://github.com/act3-ai/dagger/commit/6b5cbd2faf75b349d72c17ea17c1a11f3cd4b410) by @pspurlock

- Move config out of New and into functions instead [730de2d](https://github.com/act3-ai/dagger/commit/730de2ded4773b679f4d205bd8cbe38bc57d1932) by @pspurlock


## [git-cliff/v0.2.1] - 2025-10-23

### 🐛 Bug Fixes

- Add WithTagPattern and WithBumpedVersion functions [e48fa16](https://github.com/act3-ai/dagger/commit/e48fa162cb0c65f2e5763dd0a2fc00c823694e33) by **Paul Spurlock**


## [git-cliff/v0.2.0] - 2025-10-17

### 🚀 Features

- Change --src to --gitref and only accept a dagger.GitRef instead of a directory [848fd67](https://github.com/act3-ai/dagger/commit/848fd674342b1a77296a23d6907857b4fc11dec1) by **Paul Spurlock**


### 🐛 Bug Fixes

- Upgrade dagger engine to v0.19.2 [71853cb](https://github.com/act3-ai/dagger/commit/71853cbbccbb65652efddff50a972241b943542a) by **Paul Spurlock**


## [git-cliff/v0.1.4] - 2025-09-19

### 🐛 Bug Fixes

- Upgrade git-cliff dagger engine to v.0.18.19 [9dbb228](https://github.com/act3-ai/dagger/commit/9dbb228948f68b0b20b5ec03802fb0c48cc32f9b) by **Paul Spurlock**


## [git-cliff/v0.1.3] - 2025-08-19

### 🐛 Bug Fixes

- Upgrade dagger to v0.18.16 [38c27fc](https://github.com/act3-ai/dagger/commit/38c27fc73e6e00f48160398766eed994a26efc4f) by **Paul Spurlock**


## [git-cliff/v0.1.2] - 2025-06-25

### 💼 Other

- Bump dagger engine version v0.18.10 to v0.18.11 [fcc9b6e](https://github.com/act3-ai/dagger/commit/fcc9b6e1e68c7d7c009a68a46b6f93489467853c) by **nathan-joslin**


## [git-cliff/v0.1.1] - 2025-06-23

### 🐛 Bug Fixes

- *(git-cliff)* Upgrade dagger to v0.18.10

## [git-cliff/v0.1.0] - 2025-06-18

