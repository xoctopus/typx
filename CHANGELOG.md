
<a name="HEAD"></a>
## [HEAD](https://github.com/xoctopus/x/compare/v0.4.4...HEAD)

> 2026-07-20

### Chore

* remove ptrx.Ptr dependency
* format code
* **deps:** bump xoctopus/x to latest
* **deps:** bump lint actions

### Doc

* update README


<a name="v0.4.4"></a>
## [v0.4.4](https://github.com/xoctopus/x/compare/v0.4.3...v0.4.4)

> 2026-07-15

### Chore

* upgrade dependencies
* fmt code
* regen scripts
* **deps:** bump dependencies
* **deps:** bump dependencies


<a name="v0.4.3"></a>
## [v0.4.3](https://github.com/xoctopus/x/compare/v0.4.2...v0.4.3)

> 2026-03-17


<a name="v0.4.2"></a>
## [v0.4.2](https://github.com/xoctopus/x/compare/v0.4.1...v0.4.2)

> 2026-03-17

### Feat

* **helper:** exports Literal


<a name="v0.4.1"></a>
## [v0.4.1](https://github.com/xoctopus/x/compare/v0.4.0...v0.4.1)

> 2026-03-17

### Chore

* go fix for mordernization

### Ci

* update Makefile
* update Makefile

### Doc

* update README
* update README

### Feat

* **helper:** exports LitType meta


<a name="v0.4.0"></a>
## [v0.4.0](https://github.com/xoctopus/x/compare/v0.3.4...v0.4.0)

> 2026-03-08

### Chore

* format config and regen
* **deps:** bump xoctopus/x to v0.4.0
* **deps:** bump dependencies
* **deps:** bump github.com/xoctopus/x from 0.2.11 to 0.3.0
* **deps:** bump actions/setup-go from 5 to 6 ([#1](https://github.com/xoctopus/x/issues/1))
* **deps:** bump actions/checkout from 4 to 6 ([#2](https://github.com/xoctopus/x/issues/2))
* **deps:** bump codecov/codecov-action from 4 to 5 ([#3](https://github.com/xoctopus/x/issues/3))
* **deps:** bump github.com/xoctopus/x from 0.2.7 to 0.2.11 ([#4](https://github.com/xoctopus/x/issues/4))

### Doc

* add changelog and makefile entry


<a name="v0.3.4"></a>
## [v0.3.4](https://github.com/xoctopus/x/compare/v0.3.3...v0.3.4)

> 2025-12-30

### Chore

* add linter and fixing


<a name="v0.3.3"></a>
## [v0.3.3](https://github.com/xoctopus/x/compare/v0.3.2...v0.3.3)

> 2025-12-22

### Chore

* bump dependencies
* bump dependencies


<a name="v0.3.2"></a>
## [v0.3.2](https://github.com/xoctopus/x/compare/v0.3.1...v0.3.2)

> 2025-12-03

### Feat

* remove Implements/ConvertibleTo/AssignableTo from diff underlying


<a name="v0.3.1"></a>
## [v0.3.1](https://github.com/xoctopus/x/compare/v0.3.0...v0.3.1)

> 2025-12-02

### Fix

* dumping package path


<a name="v0.3.0"></a>
## [v0.3.0](https://github.com/xoctopus/x/compare/v0.2.1...v0.3.0)

> 2025-12-02

### Chore

* bump dependencies
* remove dep of pkgx
* bump dependencies
* bump dependencies
* bump x; update Makefile
* bump dependencies and adaption
* add helper.go

### Ci

* use latest go version

### Feat

* TypeLit for dumping
* remove LitType.Type, use NewTTByRT
* add PosOfStructField to help query field token.Pos
* BREAKING CHAGE. NewTType/NewRType with context to support package scanning with workdir

### Fix

* default load mode

### Refact

* rename module name => typx
* remove pkg loading dep


<a name="v0.2.1"></a>
## [v0.2.1](https://github.com/xoctopus/x/compare/v0.2.0...v0.2.1)

> 2025-10-23

### Feat

* add Deref


<a name="v0.2.0"></a>
## [v0.2.0](https://github.com/xoctopus/x/compare/v0.1.2...v0.2.0)

> 2025-10-23

### Chore

* use syncx.Map instead of mapx
* upgrade modules
* upgrade models
* **deps:** bump dependencies
* **deps:** bump dependencies
* **deps:** bump dependencies
* **deps:** bump `x` to latest

### Feat

* upgrade modules
* use testx and fix parsex
* upgrade dep x and fixes
* **namer:** add PackageNamer. support define TypeLit renaming outside

### Fix

* remove reflect.Invalid from builtins
* check namer returns empty package path

### Refactor

* **pkgutil:** move pkgx out of internal to pkgutil as a simplified package parser

### Test

* instead of ExpectPanic for unit testing


<a name="v0.1.2"></a>
## [v0.1.2](https://github.com/xoctopus/x/compare/v0.1.1...v0.1.2)

> 2025-03-23

### Test

* move testdata out of internal


<a name="v0.1.1"></a>
## [v0.1.1](https://github.com/xoctopus/x/compare/v0.1.0...v0.1.1)

> 2025-03-23

### Refactor

* rename Typename to TypeLit to show type literal in source code


<a name="v0.1.0"></a>
## v0.1.0

> 2025-03-22

### Docs

* add README

### Feat

* initical commit

