# Changelog

All notable changes to this project will be documented in this file.

## [0.11.3] - 2026-04-13

### 🐛 Bug Fixes

- Update dagger engine to v0.20.5 [2ccba02](https://github.com/act3-ai/dagger/commit/2ccba022827795d08c6e0c3d8cf16c87d2963a4e) by @pspurlock


## [0.11.2] - 2026-04-09

### 🐛 Bug Fixes

- Update dagger engine to v0.20.4 [9bd37a2](https://github.com/act3-ai/dagger/commit/9bd37a219ef8b67788c8d8ddb12f30e803697bf6) by @pspurlock


## [0.11.1] - 2026-03-20

### 🐛 Bug Fixes

- Update dagger engine to v0.20.3 [c6247c6](https://github.com/act3-ai/dagger/commit/c6247c60d752157a56265fcd54b61ee97537ca19) by @pspurlock

- Add ignore for .venv in src [bc61156](https://github.com/act3-ai/dagger/commit/bc611567ed447e96905865381864a2c2665d020c) by @pspurlock


## [0.11.0] - 2026-03-18

### 🚀 Features

- Separate uv sync installs to improve caching, move base out of constructor [de1a47b](https://github.com/act3-ai/dagger/commit/de1a47b605f58bddf8e8d9b32a1e1c542050ded1) by @pspurlock


### 🐛 Bug Fixes

- Add default exclude for .venv in cog complexity [508858c](https://github.com/act3-ai/dagger/commit/508858c512aa6a6f601f95058dbd6e12f277c1b0) by @pspurlock

- Rename Runtime to DevContainer, make Project() public [fd3c0d4](https://github.com/act3-ai/dagger/commit/fd3c0d4061a43ac1a2bc352f37ce5676bb43d6ea) by @pspurlock


## [0.10.1] - 2026-03-09

### 🐛 Bug Fixes

- Update dagger engine to v0.20.1 [0c35251](https://github.com/act3-ai/dagger/commit/0c3525176cc6855e753a65d5c6dc4499b7fd4e1a) by @pspurlock


## [0.10.0] - 2026-03-03

### 🚀 Features

- Add WithGitAuth for private git packages [816a53a](https://github.com/act3-ai/dagger/commit/816a53a4243f7d091277ad2935ba97287c065933) by @pspurlock


### 🐛 Bug Fixes

- Remove path parsing [ff1a5ae](https://github.com/act3-ai/dagger/commit/ff1a5ae13381ae2e9dbfa819ef84c1b6b2b17a2e) by @pspurlock

- Switch to gitcred struct [ff0c483](https://github.com/act3-ai/dagger/commit/ff0c483c20e5c3dde46fe869e1a766be2842edc6) by @pspurlock

- Switch to module git cred script [ab40be7](https://github.com/act3-ai/dagger/commit/ab40be7a97e0a9972f5c86429b27a568c2700288) by @pspurlock

- Convert host to correct env format [2206f26](https://github.com/act3-ai/dagger/commit/2206f26c7f134a91b94ddfaec0304bcad453bafd) by @pspurlock

- Add host converstion to script also [a4f3d3e](https://github.com/act3-ai/dagger/commit/a4f3d3e2e4b22c3990af53963036c0a017d45cf6) by @pspurlock

- Get rid gitcred struct [0301cdc](https://github.com/act3-ai/dagger/commit/0301cdc07fa4f1a1afcc487429e0b5c6ca860534) by @pspurlock

- Move git-credential script to constructor [8837b75](https://github.com/act3-ai/dagger/commit/8837b7565dbe2b23c1f88da2ce36ed3d60b5848c) by @pspurlock

- SetSecret to same name as secretvar [d490e00](https://github.com/act3-ai/dagger/commit/d490e00da2346a04735b6adf30fa3791d71192ce) by @pspurlock


## [0.9.2] - 2026-02-27

### 🐛 Bug Fixes

- Add flake8-cognitive-complexity scan [6bb994b](https://github.com/act3-ai/dagger/commit/6bb994b7815be012b3facb0882a25744d2ac2d20) by @pspurlock

- Add cognitive_complexity func, replace flake8 [f5bd984](https://github.com/act3-ai/dagger/commit/f5bd98482d2e02d1a12dfc4dcb8fa6c5fe883aef) by @pspurlock

- Add maxComplexity arg [d0f5c64](https://github.com/act3-ai/dagger/commit/d0f5c64736534df63dc0764a68314433ffff654a) by @pspurlock


## [0.9.1] - 2026-02-06

### 🐛 Bug Fixes

- Updating dagger engine to v0.19.11 [fffadd0](https://github.com/act3-ai/dagger/commit/fffadd099a36442ccebeaef23b25a1618f0f2974) by @pspurlock


## [0.9.0] - 2026-02-05

### 🚀 Features

- Allow specifying ruff report output formats [3443959](https://github.com/act3-ai/dagger/commit/3443959c45233258804fb56e50df85c1d94079c7) by @gresavage, Signed-off-by:Tom Gresavage <tomgresavage@gmail.com>


### 🐛 Bug Fixes

- Make ruff more DRY,  add format output options to reports [d473dc6](https://github.com/act3-ai/dagger/commit/d473dc6002a63e1904172b03a401d23e87a16574) by @pspurlock

- Update description [cb7adb3](https://github.com/act3-ai/dagger/commit/cb7adb36d6c6aca1c737e3af661e2ad52bd3f260) by @pspurlock


## [0.8.0] - 2026-02-04

### 🚀 Features

- Allow for unit test results in addition to coverage reports [add9728](https://github.com/act3-ai/dagger/commit/add9728e931613e1c5a0b56d236750a289f457ff) by @gresavage, Signed-off-by:Tom Gresavage <tomgresavage@gmail.com>


### 🐛 Bug Fixes

- Add tests for pytest Report() [fc6b33f](https://github.com/act3-ai/dagger/commit/fc6b33ff231c0161d3ec1161577a6c66d06ac0ee) by @pspurlock


## [0.7.0] - 2026-02-01

### 🚀 Features

- Add FormatRepor()  and LintReport(), add autodetect ruff version and use Base for faster scans [4f50bc5](https://github.com/act3-ai/dagger/commit/4f50bc5821f412e0a71a6f7be6eb2fe688bb60b1) by @pspurlock


### 🐛 Bug Fixes

- Make version checking private [7299bd3](https://github.com/act3-ai/dagger/commit/7299bd328a2b8c7b2cd74197c6e8628765a792c8) by @pspurlock


## [0.6.0] - 2026-01-28

### 🚀 Features

- Add fix function for lint, add cache for ruff_cache. [24ee443](https://github.com/act3-ai/dagger/commit/24ee443e6feb2190fd4df464af78096655859f79) by @pspurlock


### 🐛 Bug Fixes

- Remove unnecessary ctx from tests [e8de01f](https://github.com/act3-ai/dagger/commit/e8de01f1abdc5237eab6b7f427c531eaf5776900) by @pspurlock

- Add UV_LINK_MODE to remove warning [b6ba31d](https://github.com/act3-ai/dagger/commit/b6ba31d7fa73ae76de9971a766c30d2e3d8d8bd4) by @pspurlock


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
