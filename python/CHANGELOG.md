# Changelog

All notable changes to this project will be documented in this file.

## [0.5.2] - 2026-01-26

### 🐛 Bug Fixes

- Bug with changeset diff with source [20ba7b7](https://github.com/act3-ai/dagger/commit/20ba7b78e469331e0b65ce3758987e712b0856b3) by @pspurlock


## [0.5.1] - 2026-01-15

### 🐛 Bug Fixes

- Updating dagger engine to v0.19.10 [b274f00](https://github.com/act3-ai/dagger/commit/b274f0059db2eedf25704a739239c0b2b5612479) by @pspurlock

- Updating dagger engine to v0.19.10 in tests [c2bb264](https://github.com/act3-ai/dagger/commit/c2bb264b320d1b78314295e9a0417776f0847495) by @pspurlock


## [0.5.0] - 2026-01-08

### 🚀 Features

- Rename Format to Fix, made pytest args more DRY [82332b1](https://github.com/act3-ai/dagger/commit/82332b1d479f644d35d35c68ed2afce1da6fd312) by @pspurlock


### 🐛 Bug Fixes

- Bug with Report [9d07825](https://github.com/act3-ai/dagger/commit/9d078257d896cd3c464ab282ac2e193f37e9b481) by @pspurlock


## [0.4.0] - 2026-01-06

### 🐛 Bug Fixes

- Tests [a13d249](https://github.com/act3-ai/dagger/commit/a13d24979968d5502c03b6321bc98039da4f1877) by @pspurlock


### 🚜 Refactor

- Mypy check renamed lint and now returns a container  instead of err, report returns a json file [e7cfdbd](https://github.com/act3-ai/dagger/commit/e7cfdbdc077e18c9dae45a98991673b46ccc1163) by @pspurlock

- Pylint check renamed lint and now returns a container  instead of err, report returns a json file [38d2dc5](https://github.com/act3-ai/dagger/commit/38d2dc54e2653fcd1590631986969ee0ae56e5b8) by @pspurlock

- Pyright check renamed lint and now returns a container  instead of err, report returns a json file [22cc6a4](https://github.com/act3-ai/dagger/commit/22cc6a46fd402f91a3612664bd9c53ddbe18334b) by @pspurlock

- Pytest check renamed test and now returns a container  instead of err, removed exitcode and output from report results [d7fc348](https://github.com/act3-ai/dagger/commit/d7fc34826c66f6378f5c2051fea8c3538dc27f45) by @pspurlock

- Ruff lint now returns a container  instead of err, check func removed, report returns a json file [9c75036](https://github.com/act3-ai/dagger/commit/9c7503683f83fcc3209a6b8098ff978fe7630643) by @pspurlock

- Ruff format returns a changeset directly [b79c48e](https://github.com/act3-ai/dagger/commit/b79c48ed8192548253e69fa2302ee58ed632e00e) by @pspurlock


## [0.3.3] - 2025-12-19

### 🐛 Bug Fixes

- Descriptions [1ff427c](https://github.com/act3-ai/dagger/commit/1ff427cd5e4038f818c8db20217b9e8fe8c45e4a) by @pspurlock

- Formatting [9ad4116](https://github.com/act3-ai/dagger/commit/9ad41168af663c43e98eee0923778cfe744cb7e5) by @pspurlock


## [0.3.2] - 2025-12-19

### 🐛 Bug Fixes

- Expose the base container [71355f6](https://github.com/act3-ai/dagger/commit/71355f68426214ab0b2f7b1689d086945ce4b70e) by @ktarplee, Closes #89, Signed-off-by:Kyle M. Tarplee <kyle.tarplee@udri.udayton.edu>

- Remove "--cov-fail-under=100" [43be5c3](https://github.com/act3-ai/dagger/commit/43be5c383b04b8949458aa053c3e3ae38193619d) by @ktarplee, Signed-off-by:Kyle M. Tarplee <kyle.tarplee@udri.udayton.edu>


## [0.3.1] - 2025-12-16

### 🐛 Bug Fixes

- Updating dagger engine to v0.19.8 [645a634](https://github.com/act3-ai/dagger/commit/645a634f6ac77db09e413d16ad7f6b0e31945815) by @pspurlock


## [0.3.0] - 2025-12-12

### 🚀 Features

- Refactor python to return results structs, add new Check() commands for errors [832be91](https://github.com/act3-ai/dagger/commit/832be91953d882c48a2ae466635777e6f5b90476) by @pspurlock


### 🐛 Bug Fixes

- Tests for refactor [9166c48](https://github.com/act3-ai/dagger/commit/9166c4845ff190f8f83fc7af04879ba3d0f28ed6) by @pspurlock

- Make exit-code private [7531671](https://github.com/act3-ai/dagger/commit/75316710ef65f56f7c2dbae8d1ebe7ed4f83231c) by @pspurlock


## [0.2.2] - 2025-12-08

### 🐛 Bug Fixes

- General cleanup [03cc59f](https://github.com/act3-ai/dagger/commit/03cc59f49992a4ba8620795bd3e73b19fc8090fb) by @ktarplee, Signed-off-by:Kyle M. Tarplee <kyle.tarplee@udri.udayton.edu>

- Bug netrc still in python struct [92194e9](https://github.com/act3-ai/dagger/commit/92194e9864af622a8665e89a8693152d497681f2) by @pspurlock

- Tests to work with new refactor [6a0efbb](https://github.com/act3-ai/dagger/commit/6a0efbb5200c07d6b46774d512cc77be3064e983) by @pspurlock


### 🚜 Refactor

- Linters to return structs with results file and exit code. [709a273](https://github.com/act3-ai/dagger/commit/709a2731fc26a5caaf91e6f12ab5adabac2dafad) by @pspurlock

- Make all linters subcommands with checks instead, remove UV() and move to Base instead [0905e28](https://github.com/act3-ai/dagger/commit/0905e28e028779d316e92bb16574e955d016abe3) by @pspurlock


## [0.2.1] - 2025-11-26

### 🐛 Bug Fixes

- Update dagger to v0.19.7 [2a37e6b](https://github.com/act3-ai/dagger/commit/2a37e6b48a4e91a603f4caf21618941233f5dc4c) by @pspurlock


### 🚜 Refactor

- WithRegistryAuth [f1190ce](https://github.com/act3-ai/dagger/commit/f1190ce4fe83fc42f175cfb51b9de57acb658e03) by @pspurlock


## [python/v0.2.0] - 2025-10-28

### 🚀 Features

- Add WithRegistryCreds function and remove netrc flag [75a5fb1](https://github.com/act3-ai/dagger/commit/75a5fb1ddc3cb07d8c41873561f17d38526f009a) by @pspurlock


### 🐛 Bug Fixes

- Upgrade dagger engine to v0.19.3 [036d46d](https://github.com/act3-ai/dagger/commit/036d46d1f04addf2bbf4f9c92c90a5c883ca8050) by @pspurlock

- Improve descriptions on publish [a180cc0](https://github.com/act3-ai/dagger/commit/a180cc0fb25a9c673d8e1e760c8c746ab36ba57b) by @pspurlock

- Add WithNetrc function [09e8181](https://github.com/act3-ai/dagger/commit/09e8181a341dadd068836a84a20e2df85905ef72) by @pspurlock


## [0.1.8] - 2025-11-20

### 🐛 Bug Fixes

- Updating dagger engine to v0.19.6 [86e1674](https://github.com/act3-ai/dagger/commit/86e1674d8f775e4ced19750c224a1d696ba6607b) by @pspurlock


## [0.1.7] - 2025-11-06

### 🐛 Bug Fixes

- Test commit [11b383f](https://github.com/act3-ai/dagger/commit/11b383f84ef3cbbc59df1949a4acb532f8a6e505) by @pspurlock


## [python/v0.1.6] - 2025-09-19

### 🐛 Bug Fixes

- Upgrade python dagger engine to v0.18.19 [1bd6ef4](https://github.com/act3-ai/dagger/commit/1bd6ef462bf64bb19d996c6ea91310475279cad4) by **Paul Spurlock**


## [python/v0.1.5] - 2025-08-19

### 🐛 Bug Fixes

- Upgrade dagger to v0.18.16 [b401190](https://github.com/act3-ai/dagger/commit/b40119027dcfe796cc40778f7a442eb7660d1656) by **Paul Spurlock**


## [python/v0.1.4] - 2025-07-16

### 🐛 Bug Fixes

- Add dagger tests for python [f9a32ed](https://github.com/act3-ai/dagger/commit/f9a32ed6b0d79c48ba09e3dc71023a49fb34a0e7) by **Paul Spurlock**

- Upgrade dagger engine to v0.18.12 [a8363af](https://github.com/act3-ai/dagger/commit/a8363af58bd4e54a3c400a8bfc9165e2c000c60a) by **Paul Spurlock**

- Add option for additional build args in publish [0177d94](https://github.com/act3-ai/dagger/commit/0177d9436b41399de2338d8a9e6781bb5c54d7f8) by **Paul Spurlock**


## [python/v0.1.3] - 2025-07-03

### 🐛 Bug Fixes

- Add dagger tests for python [8ccdf18](https://github.com/act3-ai/dagger/commit/8ccdf186c934860030ca1eb2b2018553e533d040) by **Paul Spurlock**

- Upgrade dagger engine to v0.18.12 [2315fa8](https://github.com/act3-ai/dagger/commit/2315fa812e8e41a9389b3bbdc83edb01f07276fa) by **Paul Spurlock**


## [python/v0.1.2] - 2025-06-25

### 💼 Other

- Bump dagger engine version v0.18.10 to v0.18.11 [15b19f5](https://github.com/act3-ai/dagger/commit/15b19f514982382566b852e7aac94d574e3ed997) by **nathan-joslin**


## [python/v0.1.1] - 2025-06-23

### 🐛 Bug Fixes

- *(python)* Update dagger to v0.18.10

## [python/v0.1.0] - 2025-06-18

🚀 Initial release 🚀
