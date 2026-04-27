# Changelog

All notable changes to this project will be documented in this file. See [standard-version](https://github.com/conventional-changelog/standard-version) for commit guidelines.

## [2.7.0](https://github.com/steierma/cogni-cash/compare/v2.0.0...v2.7.0) (2026-04-27)


### Features

* add dynamic log level control and enhance bank connection debugging ([59a74c9](https://github.com/steierma/cogni-cash/commit/59a74c9f1f11ecdebe4ac7c0a30d838999caae1a))
* add Frontend Type Guardian skill and mandate to prevent TypeScript syntax errors ([12d5cff](https://github.com/steierma/cogni-cash/commit/12d5cff84dc1a13418dd1025a461c9d551ce356a))
* **backend:** implement robust graceful shutdown and background worker tracking ([b4159c3](https://github.com/steierma/cogni-cash/commit/b4159c3d14808c2fcb97128b5e909988ea09e04d))
* **backend:** optimize categorization with bulk updates and improve test coverage ([f7ba37e](https://github.com/steierma/cogni-cash/commit/f7ba37e57c32a0cdbd97730a328c928854ac56af))
* **backend:** support AES-encrypted PDFs for corporate payslips and bank statements ([d27804c](https://github.com/steierma/cogni-cash/commit/d27804c35c53cafc8006ab98f1b01251c0320232))
* **bank-statements:** exhaustive parsing and localized high-signal error handling ([f1059fc](https://github.com/steierma/cogni-cash/commit/f1059fc9eea770a6b2486c3d60a40da2f86e345e))
* **bank:** enhance bank connection UX and fix multi-tenancy account isolation ([bf330bd](https://github.com/steierma/cogni-cash/commit/bf330bd946f4f7e5c7d06232e6cbf5c2c17f98b3))
* **bank:** implement counterparty IBAN and transaction code mapping for Enable Banking ([8137052](https://github.com/steierma/cogni-cash/commit/813705257c248288d2ffdc06b605a96ae07281de))
* **categories:** make it possible to change the forecast based on the historical data by changing the time period for the past ([965bcef](https://github.com/steierma/cogni-cash/commit/965bcefe09fcddae541ac997863bb4037d4586ef))
* **cogni-cash:** multi currency support ([020d637](https://github.com/steierma/cogni-cash/commit/020d637200e52d5531e5968e73026ad7d1d8a477))
* **discovery:** consolidate subscription discovery engine and implement strict name-based deduplication ([9003880](https://github.com/steierma/cogni-cash/commit/9003880f28b362f2433a3ef7cad767c49bf351b0))
* **discovery:** implement user-configurable discovery tolerances and UI settings ([6d6e21c](https://github.com/steierma/cogni-cash/commit/6d6e21c2459def8751706f64af67139658e0e496))
* **docs:** add Subscription Management & One-Click Cancellation concept ([f5144c7](https://github.com/steierma/cogni-cash/commit/f5144c7e3833cd282bbea3f891b5ca23414f091b))
* **documents:** implement document vault and refactor frontend api into modular services ([b66d438](https://github.com/steierma/cogni-cash/commit/b66d438f62f927dbcb755e054012c2394d8f4a27))
* forecast engine simplification + subscription-driven projections ([f72c258](https://github.com/steierma/cogni-cash/commit/f72c258950b64cd76df44ccb85ecbf75ff494321))
* **forecasting:** implement last bank day scheduling for planned transactions ([a311d3e](https://github.com/steierma/cogni-cash/commit/a311d3e53aae6e8a0e713ac8357588913b760bb4))
* **forecasting:** implement recurring manual transactions and smart auto-suppression ([4c15e9d](https://github.com/steierma/cogni-cash/commit/4c15e9d0cf00235cda4fca987b66ff89a958194d))
* **forecasting:** small improvements for smartphones ([10beb4e](https://github.com/steierma/cogni-cash/commit/10beb4e610c5ed7d1b5a9b429549b8412ebdfe9e))
* **frontend:** add actual vs. burn-rate comparison and 24-month forecast range ([9e68a44](https://github.com/steierma/cogni-cash/commit/9e68a447041ea90bbeb761fab9465da22afd1657))
* **frontend:** add loading state for subscription discovery ([3c2e187](https://github.com/steierma/cogni-cash/commit/3c2e18706ce1e6f11924da58d62c1806bfe1779a))
* **frontend:** refactor SubscriptionsPage layout and split active/canceled tables ([977f9c8](https://github.com/steierma/cogni-cash/commit/977f9c8244b48e01982e548e1984db098065cff8))
* implement batch transaction linking and improve forecasting accuracy ([fdbd3d9](https://github.com/steierma/cogni-cash/commit/fdbd3d9aed444639166add757f8008d8e17f352b))
* implement document re-upload for vault recovery ([ec70fc8](https://github.com/steierma/cogni-cash/commit/ec70fc887fcabd9a8dddca2a460cac205d24d868))
* implement editable billing intervals and enhance bank connection documentation ([cdbe0bb](https://github.com/steierma/cogni-cash/commit/cdbe0bb05ecb9b022bc94f57b331199a30d0077f))
* implement LLM error email notifications and client-specific settings ([922308f](https://github.com/steierma/cogni-cash/commit/922308f3dcc2c8b22c159a9907659d9e03a6623b))
* implement mandatory document encryption for payslips, invoices, and bank statements ([1cfe804](https://github.com/steierma/cogni-cash/commit/1cfe8041e91901e957decf12093c9b7346f881ca))
* implement multi-LLM profiles with global enforcement and secure token management ([5db5e79](https://github.com/steierma/cogni-cash/commit/5db5e7903fbc59e7b1e6ac13a40984c89b25dbf1))
* implement proactive bank connection management, document vault filtering/previews, and security hardening ([76dc0b0](https://github.com/steierma/cogni-cash/commit/76dc0b0b5216777ef7532985566d0e38d14f2648))
* **ingcsv:** enhance parser with explicit statement typing and joint account support ([d88204d](https://github.com/steierma/cogni-cash/commit/d88204d191f0eabbfb8791b116ccc0a189c6727d))
* **invoices:** attribute remaining unsplit amount to main category in breakdown ([90dd5b2](https://github.com/steierma/cogni-cash/commit/90dd5b2ee7eb30141417e5610c8bf656653098bf))
* **invoices:** implement multi-category split tracking and update documentation ([623a873](https://github.com/steierma/cogni-cash/commit/623a8733be6c22c9b93f871e0c5e5ca542690279))
* make subscription approval asynchronous with background AI enrichment ([3431a9d](https://github.com/steierma/cogni-cash/commit/3431a9dd5c917e7e6412f4d91db3b7597a8935e7))
* **mobile:** align menu and features with web frontend ([171fab0](https://github.com/steierma/cogni-cash/commit/171fab044519ee76585b4c8e6ac8e69202ca9837))
* **mobile:** implement guided onboarding, sharing dashboard, and multi-page document scanning ([d191927](https://github.com/steierma/cogni-cash/commit/d191927acdcd4918c716bc1401920dc3b4dd242c)), closes [#37](https://github.com/steierma/cogni-cash/issues/37) [#38](https://github.com/steierma/cogni-cash/issues/38) [#39](https://github.com/steierma/cogni-cash/issues/39)
* **mobile:** implement subscription parity and fix soft-delete category leak in analytics ([41e57bb](https://github.com/steierma/cogni-cash/commit/41e57bb80073dbc99c1c8126b3245516bea24b0c))
* **multi-currency:** decouple AI prompts, implement dynamic UI formatting, and localize subscription enrichment ([7108fd8](https://github.com/steierma/cogni-cash/commit/7108fd8c7cc741402c41a6d85824fae0f016e81b))
* **security:** enforce role-based access for sensitive settings and bank integration ([f76b63a](https://github.com/steierma/cogni-cash/commit/f76b63aec01e327c8da068c4186058e33be4ff4f))
* **settings:** harden settings service with encryption and masking ([09a84d9](https://github.com/steierma/cogni-cash/commit/09a84d9c8343ae93af5663948bbca489db63d1de))
* **sharing:** add backend, frontend and first mobile draft ([a4937e7](https://github.com/steierma/cogni-cash/commit/a4937e71bbcfe6fbafad1aa78a238a3207fcb419))
* **sharing:** implement bank account centric collaborative forecasting and virtual accounts ([19131bb](https://github.com/steierma/cogni-cash/commit/19131bbcebe40c5884f7719dce1133b136cec270))
* standardize column configuration UI and persistence across web and mobile ([fa53041](https://github.com/steierma/cogni-cash/commit/fa53041cf50ba3633b38d9362e124518bf2499b0))
* **subscription-management:** broad backfill, activity log fix, cancellation UI fix, and extended field editing ([87312b0](https://github.com/steierma/cogni-cash/commit/87312b0f5212b202535762a3ac8a063d367d3f96))
* **subscription-management:** enhance discovery stability, AI clarity, and manual deactivation ([db27507](https://github.com/steierma/cogni-cash/commit/db275076130750631441b555d6a186cfd668b299))
* **subscription-management:** implement full subscription lifecycle including discovery, AI enrichment, and one-click cancellation ([87cebcd](https://github.com/steierma/cogni-cash/commit/87cebcd71bb8b35380aea9ff0c006a476e1c8dc6))
* **subscriptions:** implement ability to permanently decline subscription suggestions ([d949cd0](https://github.com/steierma/cogni-cash/commit/d949cd0d70a2affae7046a06c69b2ee07ec965cc))
* **subscriptions:** implement ability to undecline/restore ignored subscription suggestions ([7cded9f](https://github.com/steierma/cogni-cash/commit/7cded9f573f46e9f30b062821351c23062e38b87))
* **subscriptions:** implement manual subscription creation from transactions ([6dd324c](https://github.com/steierma/cogni-cash/commit/6dd324cbec463ee7dfe08e7e67a9960150028238))
* **subscriptions:** make discovery lookback period configurable ([2ad257c](https://github.com/steierma/cogni-cash/commit/2ad257ccfd0c25d3d11f7e74340a54ce8c2a0668))
* **subscriptions:** mandate-first discovery consolidation and automatic enrichment ([3bb7b65](https://github.com/steierma/cogni-cash/commit/3bb7b654228f106c096cd97bbebb6e0c9f27cb21))
* **subscriptions:** support bulk linking of transactions to subscriptions ([d16b304](https://github.com/steierma/cogni-cash/commit/d16b304d452fdfa06b80dbcaea38a74acaedb6d2))
* **testing:** add E2E tests for Payslips page and fix in-memory seeding ([6aae46d](https://github.com/steierma/cogni-cash/commit/6aae46dd2a702d8b0345055caaf1fc5bd3ddbee1))
* **testing:** implement E2E integration concept with Playwright and Forgejo CI ([e8dc94e](https://github.com/steierma/cogni-cash/commit/e8dc94e4f5010199cb8938ea0d89956693e3691e))
* **transaction:** improve categorisation and review all ([d291ed3](https://github.com/steierma/cogni-cash/commit/d291ed3eaff36c91e6b4eff1eeb3873988ea8d4a))
* **transactions:** add 'Search similar' feature to transaction list ([72fa4f0](https://github.com/steierma/cogni-cash/commit/72fa4f0c9267f11d50c3511024de96ff9d02e0f8))
* **types:** add subscription_id to Transaction interface and document manual creation concept ([7400d13](https://github.com/steierma/cogni-cash/commit/7400d1343aef7e5c82f092c415fa35b5340a424f))
* **web:** implement bank reauthentication and PSD2 lifecycle UI ([1bf59ff](https://github.com/steierma/cogni-cash/commit/1bf59ff74d037512544ab2c332a2da7e092b65ef))


### Bug Fixes

* **api:** correct MIME type detection for XLS bank statements ([87fc2f3](https://github.com/steierma/cogni-cash/commit/87fc2f3f3a1c3fbb15b8f2daad62cc0d0f09f091))
* **backend:** correct NewInvoiceService call in main.go to match new signature ([c2078bc](https://github.com/steierma/cogni-cash/commit/c2078bc0733ef9c1e096f8ae6d7fa10bb070caa2))
* **backend:** handle parser correctly to return specific error if vw parser fails ([342b20c](https://github.com/steierma/cogni-cash/commit/342b20c5519fe0e739fd3ca64688bf902e785ab8))
* **backend:** import strings ([0581978](https://github.com/steierma/cogni-cash/commit/05819788f18ff3f96323f3b50625b514c3c8466b))
* **backend:** improve subscription discovery filtering to check both description and counterparty ([eab0af0](https://github.com/steierma/cogni-cash/commit/eab0af075c634f85cd5052080c310c78ddc87515))
* **backend:** refine forecasting burn-rate and subscription reconciliation ([75d1533](https://github.com/steierma/cogni-cash/commit/75d1533938a170f5d13b6f7311ad57631e954cd7))
* **bank_statement:** improve logging for parsing and validation failures ([4eca776](https://github.com/steierma/cogni-cash/commit/4eca7768032a528299490c669ae8dd4928fe5ff3))
* **ci:** fix backend startup in E2E and increase test timeouts ([bf332c3](https://github.com/steierma/cogni-cash/commit/bf332c3c702bd91fe7451ae813d9a7fca049ad15))
* **ci:** restore rsync-based deployment and enable local builds ([f47e182](https://github.com/steierma/cogni-cash/commit/f47e182b5f18efcc623e53b92873908a8fd7b928))
* **core:** improve reconciliation accuracy and subscription discovery ([f4ee995](https://github.com/steierma/cogni-cash/commit/f4ee995a483f9fc3f6374d5de9d60850f35e5fa4))
* **discovery:** improve subscription discovery robustness and fix LLM adapter panic ([39c882d](https://github.com/steierma/cogni-cash/commit/39c882d96036a133224a61ba0869b6125ff92ea5))
* **forecasting:** align api paths for planned transactions and improve document sorting ([0541744](https://github.com/steierma/cogni-cash/commit/05417447749c02bfe60c4e04f57984a6f962b2ee))
* **forecasting:** fix saving of recurrence fields and improve soft suppression logic ([7613959](https://github.com/steierma/cogni-cash/commit/7613959b343b953ec53d1f0dc5c3f086c0bb0adc))
* forward client IP and User-Agent to Enable Banking for improved bank compatibility ([d5fd196](https://github.com/steierma/cogni-cash/commit/d5fd1963d5bba83675b44c13734ae11b5cda30dd))
* **frontend:** add missing RefreshCcw import in TransactionsPage ([65c5bab](https://github.com/steierma/cogni-cash/commit/65c5bab0d104a1320ac66fdcac8f6d2e1264c79b))
* **frontend:** implement robust local date formatting to fix timezone boundary issues in forecasting and reconciliation ([43f12f1](https://github.com/steierma/cogni-cash/commit/43f12f1d5e42dd3a1b9187182a67b8846f3e1ae7))
* **frontend:** import LucideIcon as type to fix SyntaxError ([abb0611](https://github.com/steierma/cogni-cash/commit/abb0611fb77edc6da4c1d6064e05fdb96a0b8c25))
* **frontend:** resolve all typescript and syntax errors (type-only imports and casting) ([be13785](https://github.com/steierma/cogni-cash/commit/be1378532be8efeedecc3b88a00c415dfa2da2ba))
* **frontend:** resolve build errors in SubscriptionsPage ([5f925d0](https://github.com/steierma/cogni-cash/commit/5f925d00c282ec210540ac2eb730b1ac99729f19))
* **i18n:** add missing translation keys for forecasting and settings ([eb54d1f](https://github.com/steierma/cogni-cash/commit/eb54d1ff378fd34d13755b2d4bac2e3e769ede18))
* **ingcsv:** correctly terminate metadata loop when header is reached ([3c03bde](https://github.com/steierma/cogni-cash/commit/3c03bde36e58011cd01a81693761bb9d88295cde))
* **invoices:** allow clearing all splits by sending an empty array instead of undefined ([cbb144f](https://github.com/steierma/cogni-cash/commit/cbb144f9c42ef283428c02a4898256fd7e8589b2))
* **invoices:** fix 400 Bad Request on invoice update by ensuring ISO date format and filtering splits ([65195f1](https://github.com/steierma/cogni-cash/commit/65195f173dca1da389011ff76071f4d9389b43c8))
* **mobile:** add missing Category import in forecast_view ([728ebe9](https://github.com/steierma/cogni-cash/commit/728ebe9aa40f365fa8cf8b1326975ebe3825f330))
* **multi-currency:** provide drop-down for currency ([a57d57e](https://github.com/steierma/cogni-cash/commit/a57d57e880bf2ad3ba642986fc57eb9e3ec17fa3))
* **reconcile:** align JSON payload keys with backend and improve error feedback ([e23d7ed](https://github.com/steierma/cogni-cash/commit/e23d7edef19e2704cc964b6bc683c862a4398946))
* remove unused import in vault_check script to fix CI build ([8773a22](https://github.com/steierma/cogni-cash/commit/8773a226385492886f8c5371c46f1938b9333545))
* resolve bank sync SQL error and mobile build failure; feat: add burn rate expansion in mobile forecast ([5c22daf](https://github.com/steierma/cogni-cash/commit/5c22daf64990332e7e3c52a9390518f61d808fe5))
* resolve build failures caused by redeclared identifiers in HTTP adapter ([75879cd](https://github.com/steierma/cogni-cash/commit/75879cda9e413506dda4dab40c756344521d8efe))
* resolve test failures across full stack and implement resilient document decryption ([a2e6efa](https://github.com/steierma/cogni-cash/commit/a2e6efa3b617d9081d1b5d6283ac123150de6a90))
* **service:** resolve build error in forecasting and align tests with discovery tolerances ([99f275f](https://github.com/steierma/cogni-cash/commit/99f275f5bfedf164ce00809f90dd314e569f9d85))
* standardize LLM profile active state field to 'is_active' ([dee10d6](https://github.com/steierma/cogni-cash/commit/dee10d6bb1ffaf619ad9ae2fa0e3ef6c3cce9313))
* **subscription-management:** log entries for historical activity ([a4e6387](https://github.com/steierma/cogni-cash/commit/a4e63877765756a4e252ea7f7b57df85d6d6803e))
* **subscription-management:** smaller fixes of approving subscriptions ([6f85e35](https://github.com/steierma/cogni-cash/commit/6f85e3598854314e4a74d29c9acdd3fd9ceeeff5))
* **subscription:** correct sign for manually created subscriptions ([c412dc9](https://github.com/steierma/cogni-cash/commit/c412dc9814796e60a1555cf3ad9611a3fc6adc4d))
* **subscription:** correct spelling to 'cancelled' and improve manual creation reliability ([8e1913c](https://github.com/steierma/cogni-cash/commit/8e1913c2975e56be8ba6ef88d5b33cb7c815eb18))
* **subscriptions:** prevent duplicate suggestions after approval via hash-based filtering and improved backfill ([1d705bb](https://github.com/steierma/cogni-cash/commit/1d705bb851f38f1509a29feb142c45d3b2fa1250))
* **ui:** ensure LLM action icons are visible on mobile devices ([549802a](https://github.com/steierma/cogni-cash/commit/549802a3bfc4ab6cf3a377b5d6474879599e82e3))
* **ui:** ensure subscription suggestion tooltips are visible and opaque ([4bbdfdf](https://github.com/steierma/cogni-cash/commit/4bbdfdf8810150aae24579544556aec2ae680d4a))
* **vw-parser:** return FormatMismatch on non-PDF files to allow parser chain to continue ([a337fdf](https://github.com/steierma/cogni-cash/commit/a337fdfccbe09b2c455d88e1db836e6f812762c6))


### Performance Improvements

* **frontend:** implement row virtualization across all list views ([3a72fb3](https://github.com/steierma/cogni-cash/commit/3a72fb34c74c8d1b89ca99796287b4c02bb9ae89))
* **frontend:** implement row virtualization for ForecastTable ([3a140de](https://github.com/steierma/cogni-cash/commit/3a140de36c24c40fffe315a12fd71b6dcdc40261))
* **frontend:** optimize transaction table and fix layout overflows ([776c6ea](https://github.com/steierma/cogni-cash/commit/776c6ea879e1edd333e8441a333caa31bab05600))
* optimize transaction management with row virtualization and bulk API ([874799c](https://github.com/steierma/cogni-cash/commit/874799c9b05f1a1171f8fb6175103c3b133ae229))


### Styles

* align manual subscription creation icon with main menu (RefreshCcw) ([bcc5b9c](https://github.com/steierma/cogni-cash/commit/bcc5b9c3a4af0f50abe93d851652b9e5077fd777))


### Tests

* anonymize discovery service test data ([e43948c](https://github.com/steierma/cogni-cash/commit/e43948cbc9bcdc16ef41251b31908b0ef362ddfc))
* **backend:** significantly improve test coverage and consolidate redundant tests ([881f0d7](https://github.com/steierma/cogni-cash/commit/881f0d740138533bc0913e1dc5c8c32870f7e0f8))
* fix mockDiscoveryUseCase to implement LinkTransactions ([4f72f97](https://github.com/steierma/cogni-cash/commit/4f72f97b8d4f5fba127e553e5e5f45a375a529dd))
* **settings:** fix tests and improve coverage for hardened settings service ([b490cba](https://github.com/steierma/cogni-cash/commit/b490cba1e42adb63d14e994816649e098d1b6fd7))


### Code Refactoring

* **frontend:** fix 66 lint errors (types, cascading renders, unused vars) ([4cfbb45](https://github.com/steierma/cogni-cash/commit/4cfbb4530bcbfa32e6fa2c9d4fa655925a3d9884))
* **parser:** split VW bank parser into PDF and CSV variants ([2d74b8e](https://github.com/steierma/cogni-cash/commit/2d74b8e9d13af966bf08c8d98050fa27720f4c8a))
* unify docker image names and fix backend build network timeout ([3a29025](https://github.com/steierma/cogni-cash/commit/3a29025e4b211bd038cf53216d931041a6ee8334))


### Documentation

* add concept for shared bank account forecasting and virtual accounts ([01e30b4](https://github.com/steierma/cogni-cash/commit/01e30b4fc7c7e9163c057fd13eb6b725f51064cd))
* **core:** reorganize documentation and update architecture navigation ([1c42cfa](https://github.com/steierma/cogni-cash/commit/1c42cfac44c6cfa38a108624dab7e11b17909cd4))
* finalize windows installation guide title ([15a3e16](https://github.com/steierma/cogni-cash/commit/15a3e16e4ba4725d88955400cef8de730f63b63b))
* **stories:** add implementation roadmap for Subscription Management ([c5454fc](https://github.com/steierma/cogni-cash/commit/c5454fc7f4ce200f6fb2d4a94ac2687180f4294f))
* synchronize documentation for collaborative forecasting and virtual accounts ([2ade1c7](https://github.com/steierma/cogni-cash/commit/2ade1c7fecca9a589aadb1ba5a6f6d7590a03719))
* synchronize memory, readme, and schema documentation ([2b46101](https://github.com/steierma/cogni-cash/commit/2b46101b4d8b6a41fa2fc2aa3f014827afcdf5f2))
* update MEMORY.md with asynchronous approval info ([12f35f9](https://github.com/steierma/cogni-cash/commit/12f35f923ef949e07530190378e1f28430010692))
* update MEMORY.md with coverage improvements ([73e3397](https://github.com/steierma/cogni-cash/commit/73e3397eaef08ae560420aed2e734bf64b6be986))
* update MEMORY.md with graceful shutdown improvements ([78882d7](https://github.com/steierma/cogni-cash/commit/78882d79ced2cf571b6f0a254bead9b09f45cec0))
* update MEMORY.md with invoice split logic improvements ([0b619d6](https://github.com/steierma/cogni-cash/commit/0b619d601050e441713e533821f681cacb297f2d))
* update MEMORY.md with invoice update bug fix info ([184083b](https://github.com/steierma/cogni-cash/commit/184083b2d1d51d447e77ac10de5cc735ea132150))
* update MEMORY.md with split clearing fix info ([3365c5e](https://github.com/steierma/cogni-cash/commit/3365c5efed984e16305383528abfc85a76d6d1cf))
* update MEMORY.md with subscription discovery loading state ([edfc6d7](https://github.com/steierma/cogni-cash/commit/edfc6d730576fc2ed74bf1523f421be1111f7344))
* update MEMORY.md with XLS download fix ([f578b63](https://github.com/steierma/cogni-cash/commit/f578b639b45ddd9b89bb565661f648b8cb6bdfe3))


### Maintenance

* **backen:** add tests for coverage ([542b87c](https://github.com/steierma/cogni-cash/commit/542b87cbdfe7902a1d02f2b39423011d47702616))
* **backen:** add tests for coverage ([8a93036](https://github.com/steierma/cogni-cash/commit/8a93036af116c0d5bd8075478c3f5b2c7bb4a0e2))
* **ci:** refactor to separate public-release workflow ([a325384](https://github.com/steierma/cogni-cash/commit/a3253840eaf0efa6d2387fdf463facc7285680b7))
* **docs:** update MEMORY.md for public release v2.3.0 ([5de0538](https://github.com/steierma/cogni-cash/commit/5de05382dc72dd212903ad022206cd23d1cd704b))
* **makefile:** add test-e2e target ([83e6651](https://github.com/steierma/cogni-cash/commit/83e665192e70a6352d002bd6d3b1f1509248ba10))
* **migrations:** squash migrations ([133a871](https://github.com/steierma/cogni-cash/commit/133a871ee1df4b37e14735971b09fb9d97da6c89))
* **mobile:** finalize mobile app for sharing information ([a0b9ea8](https://github.com/steierma/cogni-cash/commit/a0b9ea8cafeed5aa85eb25d24a2726859eb44e6f))
* **mobile:** start to fix mobile app for sharing ([ff722af](https://github.com/steierma/cogni-cash/commit/ff722af07a0104ece1b29a609f6f194d46d1eb99))
* **release:** 2.1.0 ([9d32fc6](https://github.com/steierma/cogni-cash/commit/9d32fc676a771bef00d9e1ca3da218fe4e6e2578))
* **release:** 2.2.0 ([22bb85c](https://github.com/steierma/cogni-cash/commit/22bb85c5b671909da76f4f9f439b2fb0b8b9fe8e))
* **release:** 2.3.0 ([9c038db](https://github.com/steierma/cogni-cash/commit/9c038dbbbf836e112ea1646754f35e55b98786cd))
* **release:** 2.4.0 ([80b6692](https://github.com/steierma/cogni-cash/commit/80b669251c7fdb065bcb32bb497f0349a8763e03))
* **release:** 2.5.0 ([9871ca6](https://github.com/steierma/cogni-cash/commit/9871ca603143ce4eef7ece130d10221c21db9875))
* **release:** 2.6.0 ([916eedb](https://github.com/steierma/cogni-cash/commit/916eedb4ddf8c6a8d9de211d0636dc1e845781d2))
* **release:** bump version to 2.0.1 ([a600726](https://github.com/steierma/cogni-cash/commit/a600726878dcee57169aedfe81b4dc37723a6677))
* **security:** add penetration-test agent ([fdfeb16](https://github.com/steierma/cogni-cash/commit/fdfeb16fe78cb93fb87f53401b21bb5acb1e9123))
* **skills:** add public-release-snapshotter skill ([2c6fef1](https://github.com/steierma/cogni-cash/commit/2c6fef155d955da98210b672893e0d813cc5c148))
* synchronize documentation for v2.5.0 release ([f98bd60](https://github.com/steierma/cogni-cash/commit/f98bd604eeb26a16ec09316b915514a82a924e1f))

## [2.6.0](https://github.com/steierma/cogni-cash/compare/v2.0.0...v2.6.0) (2026-04-26)


### Features

* add dynamic log level control and enhance bank connection debugging ([59a74c9](https://github.com/steierma/cogni-cash/commit/59a74c9f1f11ecdebe4ac7c0a30d838999caae1a))
* add Frontend Type Guardian skill and mandate to prevent TypeScript syntax errors ([12d5cff](https://github.com/steierma/cogni-cash/commit/12d5cff84dc1a13418dd1025a461c9d551ce356a))
* **backend:** implement robust graceful shutdown and background worker tracking ([b4159c3](https://github.com/steierma/cogni-cash/commit/b4159c3d14808c2fcb97128b5e909988ea09e04d))
* **backend:** support AES-encrypted PDFs for corporate payslips and bank statements ([d27804c](https://github.com/steierma/cogni-cash/commit/d27804c35c53cafc8006ab98f1b01251c0320232))
* **bank-statements:** exhaustive parsing and localized high-signal error handling ([f1059fc](https://github.com/steierma/cogni-cash/commit/f1059fc9eea770a6b2486c3d60a40da2f86e345e))
* **bank:** enhance bank connection UX and fix multi-tenancy account isolation ([bf330bd](https://github.com/steierma/cogni-cash/commit/bf330bd946f4f7e5c7d06232e6cbf5c2c17f98b3))
* **bank:** implement counterparty IBAN and transaction code mapping for Enable Banking ([8137052](https://github.com/steierma/cogni-cash/commit/813705257c248288d2ffdc06b605a96ae07281de))
* **categories:** make it possible to change the forecast based on the historical data by changing the time period for the past ([965bcef](https://github.com/steierma/cogni-cash/commit/965bcefe09fcddae541ac997863bb4037d4586ef))
* **cogni-cash:** multi currency support ([020d637](https://github.com/steierma/cogni-cash/commit/020d637200e52d5531e5968e73026ad7d1d8a477))
* **discovery:** consolidate subscription discovery engine and implement strict name-based deduplication ([9003880](https://github.com/steierma/cogni-cash/commit/9003880f28b362f2433a3ef7cad767c49bf351b0))
* **discovery:** implement user-configurable discovery tolerances and UI settings ([6d6e21c](https://github.com/steierma/cogni-cash/commit/6d6e21c2459def8751706f64af67139658e0e496))
* **docs:** add Subscription Management & One-Click Cancellation concept ([f5144c7](https://github.com/steierma/cogni-cash/commit/f5144c7e3833cd282bbea3f891b5ca23414f091b))
* **documents:** implement document vault and refactor frontend api into modular services ([b66d438](https://github.com/steierma/cogni-cash/commit/b66d438f62f927dbcb755e054012c2394d8f4a27))
* forecast engine simplification + subscription-driven projections ([f72c258](https://github.com/steierma/cogni-cash/commit/f72c258950b64cd76df44ccb85ecbf75ff494321))
* **forecasting:** implement last bank day scheduling for planned transactions ([a311d3e](https://github.com/steierma/cogni-cash/commit/a311d3e53aae6e8a0e713ac8357588913b760bb4))
* **forecasting:** implement recurring manual transactions and smart auto-suppression ([4c15e9d](https://github.com/steierma/cogni-cash/commit/4c15e9d0cf00235cda4fca987b66ff89a958194d))
* **forecasting:** small improvements for smartphones ([10beb4e](https://github.com/steierma/cogni-cash/commit/10beb4e610c5ed7d1b5a9b429549b8412ebdfe9e))
* **frontend:** add actual vs. burn-rate comparison and 24-month forecast range ([9e68a44](https://github.com/steierma/cogni-cash/commit/9e68a447041ea90bbeb761fab9465da22afd1657))
* **frontend:** add loading state for subscription discovery ([3c2e187](https://github.com/steierma/cogni-cash/commit/3c2e18706ce1e6f11924da58d62c1806bfe1779a))
* **frontend:** refactor SubscriptionsPage layout and split active/canceled tables ([977f9c8](https://github.com/steierma/cogni-cash/commit/977f9c8244b48e01982e548e1984db098065cff8))
* implement batch transaction linking and improve forecasting accuracy ([fdbd3d9](https://github.com/steierma/cogni-cash/commit/fdbd3d9aed444639166add757f8008d8e17f352b))
* implement document re-upload for vault recovery ([ec70fc8](https://github.com/steierma/cogni-cash/commit/ec70fc887fcabd9a8dddca2a460cac205d24d868))
* implement editable billing intervals and enhance bank connection documentation ([cdbe0bb](https://github.com/steierma/cogni-cash/commit/cdbe0bb05ecb9b022bc94f57b331199a30d0077f))
* implement mandatory document encryption for payslips, invoices, and bank statements ([1cfe804](https://github.com/steierma/cogni-cash/commit/1cfe8041e91901e957decf12093c9b7346f881ca))
* implement multi-LLM profiles with global enforcement and secure token management ([5db5e79](https://github.com/steierma/cogni-cash/commit/5db5e7903fbc59e7b1e6ac13a40984c89b25dbf1))
* **ingcsv:** enhance parser with explicit statement typing and joint account support ([d88204d](https://github.com/steierma/cogni-cash/commit/d88204d191f0eabbfb8791b116ccc0a189c6727d))
* **invoices:** attribute remaining unsplit amount to main category in breakdown ([90dd5b2](https://github.com/steierma/cogni-cash/commit/90dd5b2ee7eb30141417e5610c8bf656653098bf))
* **invoices:** implement multi-category split tracking and update documentation ([623a873](https://github.com/steierma/cogni-cash/commit/623a8733be6c22c9b93f871e0c5e5ca542690279))
* make subscription approval asynchronous with background AI enrichment ([3431a9d](https://github.com/steierma/cogni-cash/commit/3431a9dd5c917e7e6412f4d91db3b7597a8935e7))
* **mobile:** align menu and features with web frontend ([171fab0](https://github.com/steierma/cogni-cash/commit/171fab044519ee76585b4c8e6ac8e69202ca9837))
* **mobile:** implement subscription parity and fix soft-delete category leak in analytics ([41e57bb](https://github.com/steierma/cogni-cash/commit/41e57bb80073dbc99c1c8126b3245516bea24b0c))
* **multi-currency:** decouple AI prompts, implement dynamic UI formatting, and localize subscription enrichment ([7108fd8](https://github.com/steierma/cogni-cash/commit/7108fd8c7cc741402c41a6d85824fae0f016e81b))
* **security:** enforce role-based access for sensitive settings and bank integration ([f76b63a](https://github.com/steierma/cogni-cash/commit/f76b63aec01e327c8da068c4186058e33be4ff4f))
* **settings:** harden settings service with encryption and masking ([09a84d9](https://github.com/steierma/cogni-cash/commit/09a84d9c8343ae93af5663948bbca489db63d1de))
* **sharing:** add backend, frontend and first mobile draft ([a4937e7](https://github.com/steierma/cogni-cash/commit/a4937e71bbcfe6fbafad1aa78a238a3207fcb419))
* **sharing:** implement bank account centric collaborative forecasting and virtual accounts ([19131bb](https://github.com/steierma/cogni-cash/commit/19131bbcebe40c5884f7719dce1133b136cec270))
* **subscription-management:** broad backfill, activity log fix, cancellation UI fix, and extended field editing ([87312b0](https://github.com/steierma/cogni-cash/commit/87312b0f5212b202535762a3ac8a063d367d3f96))
* **subscription-management:** enhance discovery stability, AI clarity, and manual deactivation ([db27507](https://github.com/steierma/cogni-cash/commit/db275076130750631441b555d6a186cfd668b299))
* **subscription-management:** implement full subscription lifecycle including discovery, AI enrichment, and one-click cancellation ([87cebcd](https://github.com/steierma/cogni-cash/commit/87cebcd71bb8b35380aea9ff0c006a476e1c8dc6))
* **subscriptions:** implement ability to permanently decline subscription suggestions ([d949cd0](https://github.com/steierma/cogni-cash/commit/d949cd0d70a2affae7046a06c69b2ee07ec965cc))
* **subscriptions:** implement ability to undecline/restore ignored subscription suggestions ([7cded9f](https://github.com/steierma/cogni-cash/commit/7cded9f573f46e9f30b062821351c23062e38b87))
* **subscriptions:** implement manual subscription creation from transactions ([6dd324c](https://github.com/steierma/cogni-cash/commit/6dd324cbec463ee7dfe08e7e67a9960150028238))
* **subscriptions:** make discovery lookback period configurable ([2ad257c](https://github.com/steierma/cogni-cash/commit/2ad257ccfd0c25d3d11f7e74340a54ce8c2a0668))
* **subscriptions:** mandate-first discovery consolidation and automatic enrichment ([3bb7b65](https://github.com/steierma/cogni-cash/commit/3bb7b654228f106c096cd97bbebb6e0c9f27cb21))
* **subscriptions:** support bulk linking of transactions to subscriptions ([d16b304](https://github.com/steierma/cogni-cash/commit/d16b304d452fdfa06b80dbcaea38a74acaedb6d2))
* **testing:** add E2E tests for Payslips page and fix in-memory seeding ([6aae46d](https://github.com/steierma/cogni-cash/commit/6aae46dd2a702d8b0345055caaf1fc5bd3ddbee1))
* **testing:** implement E2E integration concept with Playwright and Forgejo CI ([e8dc94e](https://github.com/steierma/cogni-cash/commit/e8dc94e4f5010199cb8938ea0d89956693e3691e))
* **transaction:** improve categorisation and review all ([d291ed3](https://github.com/steierma/cogni-cash/commit/d291ed3eaff36c91e6b4eff1eeb3873988ea8d4a))
* **transactions:** add 'Search similar' feature to transaction list ([72fa4f0](https://github.com/steierma/cogni-cash/commit/72fa4f0c9267f11d50c3511024de96ff9d02e0f8))
* **types:** add subscription_id to Transaction interface and document manual creation concept ([7400d13](https://github.com/steierma/cogni-cash/commit/7400d1343aef7e5c82f092c415fa35b5340a424f))


### Bug Fixes

* **api:** correct MIME type detection for XLS bank statements ([87fc2f3](https://github.com/steierma/cogni-cash/commit/87fc2f3f3a1c3fbb15b8f2daad62cc0d0f09f091))
* **backend:** correct NewInvoiceService call in main.go to match new signature ([c2078bc](https://github.com/steierma/cogni-cash/commit/c2078bc0733ef9c1e096f8ae6d7fa10bb070caa2))
* **backend:** handle parser correctly to return specific error if vw parser fails ([342b20c](https://github.com/steierma/cogni-cash/commit/342b20c5519fe0e739fd3ca64688bf902e785ab8))
* **backend:** import strings ([0581978](https://github.com/steierma/cogni-cash/commit/05819788f18ff3f96323f3b50625b514c3c8466b))
* **backend:** improve subscription discovery filtering to check both description and counterparty ([eab0af0](https://github.com/steierma/cogni-cash/commit/eab0af075c634f85cd5052080c310c78ddc87515))
* **backend:** refine forecasting burn-rate and subscription reconciliation ([75d1533](https://github.com/steierma/cogni-cash/commit/75d1533938a170f5d13b6f7311ad57631e954cd7))
* **bank_statement:** improve logging for parsing and validation failures ([4eca776](https://github.com/steierma/cogni-cash/commit/4eca7768032a528299490c669ae8dd4928fe5ff3))
* **ci:** fix backend startup in E2E and increase test timeouts ([bf332c3](https://github.com/steierma/cogni-cash/commit/bf332c3c702bd91fe7451ae813d9a7fca049ad15))
* **ci:** restore rsync-based deployment and enable local builds ([f47e182](https://github.com/steierma/cogni-cash/commit/f47e182b5f18efcc623e53b92873908a8fd7b928))
* **core:** improve reconciliation accuracy and subscription discovery ([f4ee995](https://github.com/steierma/cogni-cash/commit/f4ee995a483f9fc3f6374d5de9d60850f35e5fa4))
* **discovery:** improve subscription discovery robustness and fix LLM adapter panic ([39c882d](https://github.com/steierma/cogni-cash/commit/39c882d96036a133224a61ba0869b6125ff92ea5))
* **forecasting:** align api paths for planned transactions and improve document sorting ([0541744](https://github.com/steierma/cogni-cash/commit/05417447749c02bfe60c4e04f57984a6f962b2ee))
* **forecasting:** fix saving of recurrence fields and improve soft suppression logic ([7613959](https://github.com/steierma/cogni-cash/commit/7613959b343b953ec53d1f0dc5c3f086c0bb0adc))
* forward client IP and User-Agent to Enable Banking for improved bank compatibility ([d5fd196](https://github.com/steierma/cogni-cash/commit/d5fd1963d5bba83675b44c13734ae11b5cda30dd))
* **frontend:** add missing RefreshCcw import in TransactionsPage ([65c5bab](https://github.com/steierma/cogni-cash/commit/65c5bab0d104a1320ac66fdcac8f6d2e1264c79b))
* **frontend:** implement robust local date formatting to fix timezone boundary issues in forecasting and reconciliation ([43f12f1](https://github.com/steierma/cogni-cash/commit/43f12f1d5e42dd3a1b9187182a67b8846f3e1ae7))
* **frontend:** import LucideIcon as type to fix SyntaxError ([abb0611](https://github.com/steierma/cogni-cash/commit/abb0611fb77edc6da4c1d6064e05fdb96a0b8c25))
* **frontend:** resolve all typescript and syntax errors (type-only imports and casting) ([be13785](https://github.com/steierma/cogni-cash/commit/be1378532be8efeedecc3b88a00c415dfa2da2ba))
* **frontend:** resolve build errors in SubscriptionsPage ([5f925d0](https://github.com/steierma/cogni-cash/commit/5f925d00c282ec210540ac2eb730b1ac99729f19))
* **i18n:** add missing translation keys for forecasting and settings ([eb54d1f](https://github.com/steierma/cogni-cash/commit/eb54d1ff378fd34d13755b2d4bac2e3e769ede18))
* **ingcsv:** correctly terminate metadata loop when header is reached ([3c03bde](https://github.com/steierma/cogni-cash/commit/3c03bde36e58011cd01a81693761bb9d88295cde))
* **invoices:** allow clearing all splits by sending an empty array instead of undefined ([cbb144f](https://github.com/steierma/cogni-cash/commit/cbb144f9c42ef283428c02a4898256fd7e8589b2))
* **invoices:** fix 400 Bad Request on invoice update by ensuring ISO date format and filtering splits ([65195f1](https://github.com/steierma/cogni-cash/commit/65195f173dca1da389011ff76071f4d9389b43c8))
* **mobile:** add missing Category import in forecast_view ([728ebe9](https://github.com/steierma/cogni-cash/commit/728ebe9aa40f365fa8cf8b1326975ebe3825f330))
* **multi-currency:** provide drop-down for currency ([a57d57e](https://github.com/steierma/cogni-cash/commit/a57d57e880bf2ad3ba642986fc57eb9e3ec17fa3))
* **reconcile:** align JSON payload keys with backend and improve error feedback ([e23d7ed](https://github.com/steierma/cogni-cash/commit/e23d7edef19e2704cc964b6bc683c862a4398946))
* remove unused import in vault_check script to fix CI build ([8773a22](https://github.com/steierma/cogni-cash/commit/8773a226385492886f8c5371c46f1938b9333545))
* resolve build failures caused by redeclared identifiers in HTTP adapter ([75879cd](https://github.com/steierma/cogni-cash/commit/75879cda9e413506dda4dab40c756344521d8efe))
* resolve test failures across full stack and implement resilient document decryption ([a2e6efa](https://github.com/steierma/cogni-cash/commit/a2e6efa3b617d9081d1b5d6283ac123150de6a90))
* **service:** resolve build error in forecasting and align tests with discovery tolerances ([99f275f](https://github.com/steierma/cogni-cash/commit/99f275f5bfedf164ce00809f90dd314e569f9d85))
* **subscription-management:** log entries for historical activity ([a4e6387](https://github.com/steierma/cogni-cash/commit/a4e63877765756a4e252ea7f7b57df85d6d6803e))
* **subscription-management:** smaller fixes of approving subscriptions ([6f85e35](https://github.com/steierma/cogni-cash/commit/6f85e3598854314e4a74d29c9acdd3fd9ceeeff5))
* **subscription:** correct sign for manually created subscriptions ([c412dc9](https://github.com/steierma/cogni-cash/commit/c412dc9814796e60a1555cf3ad9611a3fc6adc4d))
* **subscription:** correct spelling to 'cancelled' and improve manual creation reliability ([8e1913c](https://github.com/steierma/cogni-cash/commit/8e1913c2975e56be8ba6ef88d5b33cb7c815eb18))
* **subscriptions:** prevent duplicate suggestions after approval via hash-based filtering and improved backfill ([1d705bb](https://github.com/steierma/cogni-cash/commit/1d705bb851f38f1509a29feb142c45d3b2fa1250))
* **ui:** ensure subscription suggestion tooltips are visible and opaque ([4bbdfdf](https://github.com/steierma/cogni-cash/commit/4bbdfdf8810150aae24579544556aec2ae680d4a))
* **vw-parser:** return FormatMismatch on non-PDF files to allow parser chain to continue ([a337fdf](https://github.com/steierma/cogni-cash/commit/a337fdfccbe09b2c455d88e1db836e6f812762c6))


### Styles

* align manual subscription creation icon with main menu (RefreshCcw) ([bcc5b9c](https://github.com/steierma/cogni-cash/commit/bcc5b9c3a4af0f50abe93d851652b9e5077fd777))


### Tests

* anonymize discovery service test data ([e43948c](https://github.com/steierma/cogni-cash/commit/e43948cbc9bcdc16ef41251b31908b0ef362ddfc))
* **backend:** significantly improve test coverage and consolidate redundant tests ([881f0d7](https://github.com/steierma/cogni-cash/commit/881f0d740138533bc0913e1dc5c8c32870f7e0f8))
* fix mockDiscoveryUseCase to implement LinkTransactions ([4f72f97](https://github.com/steierma/cogni-cash/commit/4f72f97b8d4f5fba127e553e5e5f45a375a529dd))
* **settings:** fix tests and improve coverage for hardened settings service ([b490cba](https://github.com/steierma/cogni-cash/commit/b490cba1e42adb63d14e994816649e098d1b6fd7))


### Code Refactoring

* **frontend:** fix 66 lint errors (types, cascading renders, unused vars) ([4cfbb45](https://github.com/steierma/cogni-cash/commit/4cfbb4530bcbfa32e6fa2c9d4fa655925a3d9884))
* **parser:** split VW bank parser into PDF and CSV variants ([2d74b8e](https://github.com/steierma/cogni-cash/commit/2d74b8e9d13af966bf08c8d98050fa27720f4c8a))
* unify docker image names and fix backend build network timeout ([3a29025](https://github.com/steierma/cogni-cash/commit/3a29025e4b211bd038cf53216d931041a6ee8334))


### Documentation

* add concept for shared bank account forecasting and virtual accounts ([01e30b4](https://github.com/steierma/cogni-cash/commit/01e30b4fc7c7e9163c057fd13eb6b725f51064cd))
* **core:** reorganize documentation and update architecture navigation ([1c42cfa](https://github.com/steierma/cogni-cash/commit/1c42cfac44c6cfa38a108624dab7e11b17909cd4))
* finalize windows installation guide title ([15a3e16](https://github.com/steierma/cogni-cash/commit/15a3e16e4ba4725d88955400cef8de730f63b63b))
* **stories:** add implementation roadmap for Subscription Management ([c5454fc](https://github.com/steierma/cogni-cash/commit/c5454fc7f4ce200f6fb2d4a94ac2687180f4294f))
* synchronize documentation for collaborative forecasting and virtual accounts ([2ade1c7](https://github.com/steierma/cogni-cash/commit/2ade1c7fecca9a589aadb1ba5a6f6d7590a03719))
* synchronize memory, readme, and schema documentation ([2b46101](https://github.com/steierma/cogni-cash/commit/2b46101b4d8b6a41fa2fc2aa3f014827afcdf5f2))
* update MEMORY.md with asynchronous approval info ([12f35f9](https://github.com/steierma/cogni-cash/commit/12f35f923ef949e07530190378e1f28430010692))
* update MEMORY.md with coverage improvements ([73e3397](https://github.com/steierma/cogni-cash/commit/73e3397eaef08ae560420aed2e734bf64b6be986))
* update MEMORY.md with graceful shutdown improvements ([78882d7](https://github.com/steierma/cogni-cash/commit/78882d79ced2cf571b6f0a254bead9b09f45cec0))
* update MEMORY.md with invoice split logic improvements ([0b619d6](https://github.com/steierma/cogni-cash/commit/0b619d601050e441713e533821f681cacb297f2d))
* update MEMORY.md with invoice update bug fix info ([184083b](https://github.com/steierma/cogni-cash/commit/184083b2d1d51d447e77ac10de5cc735ea132150))
* update MEMORY.md with split clearing fix info ([3365c5e](https://github.com/steierma/cogni-cash/commit/3365c5efed984e16305383528abfc85a76d6d1cf))
* update MEMORY.md with subscription discovery loading state ([edfc6d7](https://github.com/steierma/cogni-cash/commit/edfc6d730576fc2ed74bf1523f421be1111f7344))
* update MEMORY.md with XLS download fix ([f578b63](https://github.com/steierma/cogni-cash/commit/f578b639b45ddd9b89bb565661f648b8cb6bdfe3))


### Maintenance

* **backen:** add tests for coverage ([542b87c](https://github.com/steierma/cogni-cash/commit/542b87cbdfe7902a1d02f2b39423011d47702616))
* **backen:** add tests for coverage ([8a93036](https://github.com/steierma/cogni-cash/commit/8a93036af116c0d5bd8075478c3f5b2c7bb4a0e2))
* **ci:** refactor to separate public-release workflow ([a325384](https://github.com/steierma/cogni-cash/commit/a3253840eaf0efa6d2387fdf463facc7285680b7))
* **docs:** update MEMORY.md for public release v2.3.0 ([5de0538](https://github.com/steierma/cogni-cash/commit/5de05382dc72dd212903ad022206cd23d1cd704b))
* **makefile:** add test-e2e target ([83e6651](https://github.com/steierma/cogni-cash/commit/83e665192e70a6352d002bd6d3b1f1509248ba10))
* **migrations:** squash migrations ([133a871](https://github.com/steierma/cogni-cash/commit/133a871ee1df4b37e14735971b09fb9d97da6c89))
* **mobile:** finalize mobile app for sharing information ([a0b9ea8](https://github.com/steierma/cogni-cash/commit/a0b9ea8cafeed5aa85eb25d24a2726859eb44e6f))
* **mobile:** start to fix mobile app for sharing ([ff722af](https://github.com/steierma/cogni-cash/commit/ff722af07a0104ece1b29a609f6f194d46d1eb99))
* **release:** 2.1.0 ([9d32fc6](https://github.com/steierma/cogni-cash/commit/9d32fc676a771bef00d9e1ca3da218fe4e6e2578))
* **release:** 2.2.0 ([22bb85c](https://github.com/steierma/cogni-cash/commit/22bb85c5b671909da76f4f9f439b2fb0b8b9fe8e))
* **release:** 2.3.0 ([9c038db](https://github.com/steierma/cogni-cash/commit/9c038dbbbf836e112ea1646754f35e55b98786cd))
* **release:** 2.4.0 ([80b6692](https://github.com/steierma/cogni-cash/commit/80b669251c7fdb065bcb32bb497f0349a8763e03))
* **release:** 2.5.0 ([9871ca6](https://github.com/steierma/cogni-cash/commit/9871ca603143ce4eef7ece130d10221c21db9875))
* **release:** bump version to 2.0.1 ([a600726](https://github.com/steierma/cogni-cash/commit/a600726878dcee57169aedfe81b4dc37723a6677))
* **security:** add penetration-test agent ([fdfeb16](https://github.com/steierma/cogni-cash/commit/fdfeb16fe78cb93fb87f53401b21bb5acb1e9123))
* **skills:** add public-release-snapshotter skill ([2c6fef1](https://github.com/steierma/cogni-cash/commit/2c6fef155d955da98210b672893e0d813cc5c148))
* synchronize documentation for v2.5.0 release ([f98bd60](https://github.com/steierma/cogni-cash/commit/f98bd604eeb26a16ec09316b915514a82a924e1f))

## [2.5.0](https://github.com/steierma/cogni-cash/compare/v2.0.0...v2.5.0) (2026-04-25)


### Features

* add dynamic log level control and enhance bank connection debugging ([59a74c9](https://github.com/steierma/cogni-cash/commit/59a74c9f1f11ecdebe4ac7c0a30d838999caae1a))
* add Frontend Type Guardian skill and mandate to prevent TypeScript syntax errors ([12d5cff](https://github.com/steierma/cogni-cash/commit/12d5cff84dc1a13418dd1025a461c9d551ce356a))
* **backend:** implement robust graceful shutdown and background worker tracking ([b4159c3](https://github.com/steierma/cogni-cash/commit/b4159c3d14808c2fcb97128b5e909988ea09e04d))
* **backend:** support AES-encrypted PDFs for corporate payslips and bank statements ([d27804c](https://github.com/steierma/cogni-cash/commit/d27804c35c53cafc8006ab98f1b01251c0320232))
* **bank-statements:** exhaustive parsing and localized high-signal error handling ([f1059fc](https://github.com/steierma/cogni-cash/commit/f1059fc9eea770a6b2486c3d60a40da2f86e345e))
* **bank:** enhance bank connection UX and fix multi-tenancy account isolation ([bf330bd](https://github.com/steierma/cogni-cash/commit/bf330bd946f4f7e5c7d06232e6cbf5c2c17f98b3))
* **bank:** implement counterparty IBAN and transaction code mapping for Enable Banking ([8137052](https://github.com/steierma/cogni-cash/commit/813705257c248288d2ffdc06b605a96ae07281de))
* **categories:** make it possible to change the forecast based on the historical data by changing the time period for the past ([965bcef](https://github.com/steierma/cogni-cash/commit/965bcefe09fcddae541ac997863bb4037d4586ef))
* **cogni-cash:** multi currency support ([020d637](https://github.com/steierma/cogni-cash/commit/020d637200e52d5531e5968e73026ad7d1d8a477))
* **discovery:** consolidate subscription discovery engine and implement strict name-based deduplication ([9003880](https://github.com/steierma/cogni-cash/commit/9003880f28b362f2433a3ef7cad767c49bf351b0))
* **discovery:** implement user-configurable discovery tolerances and UI settings ([6d6e21c](https://github.com/steierma/cogni-cash/commit/6d6e21c2459def8751706f64af67139658e0e496))
* **docs:** add Subscription Management & One-Click Cancellation concept ([f5144c7](https://github.com/steierma/cogni-cash/commit/f5144c7e3833cd282bbea3f891b5ca23414f091b))
* **documents:** implement document vault and refactor frontend api into modular services ([b66d438](https://github.com/steierma/cogni-cash/commit/b66d438f62f927dbcb755e054012c2394d8f4a27))
* forecast engine simplification + subscription-driven projections ([f72c258](https://github.com/steierma/cogni-cash/commit/f72c258950b64cd76df44ccb85ecbf75ff494321))
* **forecasting:** implement last bank day scheduling for planned transactions ([a311d3e](https://github.com/steierma/cogni-cash/commit/a311d3e53aae6e8a0e713ac8357588913b760bb4))
* **forecasting:** implement recurring manual transactions and smart auto-suppression ([4c15e9d](https://github.com/steierma/cogni-cash/commit/4c15e9d0cf00235cda4fca987b66ff89a958194d))
* **forecasting:** small improvements for smartphones ([10beb4e](https://github.com/steierma/cogni-cash/commit/10beb4e610c5ed7d1b5a9b429549b8412ebdfe9e))
* **frontend:** add actual vs. burn-rate comparison and 24-month forecast range ([9e68a44](https://github.com/steierma/cogni-cash/commit/9e68a447041ea90bbeb761fab9465da22afd1657))
* **frontend:** add loading state for subscription discovery ([3c2e187](https://github.com/steierma/cogni-cash/commit/3c2e18706ce1e6f11924da58d62c1806bfe1779a))
* **frontend:** refactor SubscriptionsPage layout and split active/canceled tables ([977f9c8](https://github.com/steierma/cogni-cash/commit/977f9c8244b48e01982e548e1984db098065cff8))
* implement batch transaction linking and improve forecasting accuracy ([fdbd3d9](https://github.com/steierma/cogni-cash/commit/fdbd3d9aed444639166add757f8008d8e17f352b))
* implement document re-upload for vault recovery ([ec70fc8](https://github.com/steierma/cogni-cash/commit/ec70fc887fcabd9a8dddca2a460cac205d24d868))
* implement editable billing intervals and enhance bank connection documentation ([cdbe0bb](https://github.com/steierma/cogni-cash/commit/cdbe0bb05ecb9b022bc94f57b331199a30d0077f))
* implement mandatory document encryption for payslips, invoices, and bank statements ([1cfe804](https://github.com/steierma/cogni-cash/commit/1cfe8041e91901e957decf12093c9b7346f881ca))
* **ingcsv:** enhance parser with explicit statement typing and joint account support ([d88204d](https://github.com/steierma/cogni-cash/commit/d88204d191f0eabbfb8791b116ccc0a189c6727d))
* **invoices:** attribute remaining unsplit amount to main category in breakdown ([90dd5b2](https://github.com/steierma/cogni-cash/commit/90dd5b2ee7eb30141417e5610c8bf656653098bf))
* **invoices:** implement multi-category split tracking and update documentation ([623a873](https://github.com/steierma/cogni-cash/commit/623a8733be6c22c9b93f871e0c5e5ca542690279))
* make subscription approval asynchronous with background AI enrichment ([3431a9d](https://github.com/steierma/cogni-cash/commit/3431a9dd5c917e7e6412f4d91db3b7597a8935e7))
* **mobile:** align menu and features with web frontend ([171fab0](https://github.com/steierma/cogni-cash/commit/171fab044519ee76585b4c8e6ac8e69202ca9837))
* **mobile:** implement subscription parity and fix soft-delete category leak in analytics ([41e57bb](https://github.com/steierma/cogni-cash/commit/41e57bb80073dbc99c1c8126b3245516bea24b0c))
* **multi-currency:** decouple AI prompts, implement dynamic UI formatting, and localize subscription enrichment ([7108fd8](https://github.com/steierma/cogni-cash/commit/7108fd8c7cc741402c41a6d85824fae0f016e81b))
* **security:** enforce role-based access for sensitive settings and bank integration ([f76b63a](https://github.com/steierma/cogni-cash/commit/f76b63aec01e327c8da068c4186058e33be4ff4f))
* **settings:** harden settings service with encryption and masking ([09a84d9](https://github.com/steierma/cogni-cash/commit/09a84d9c8343ae93af5663948bbca489db63d1de))
* **sharing:** add backend, frontend and first mobile draft ([a4937e7](https://github.com/steierma/cogni-cash/commit/a4937e71bbcfe6fbafad1aa78a238a3207fcb419))
* **sharing:** implement bank account centric collaborative forecasting and virtual accounts ([19131bb](https://github.com/steierma/cogni-cash/commit/19131bbcebe40c5884f7719dce1133b136cec270))
* **subscription-management:** broad backfill, activity log fix, cancellation UI fix, and extended field editing ([87312b0](https://github.com/steierma/cogni-cash/commit/87312b0f5212b202535762a3ac8a063d367d3f96))
* **subscription-management:** enhance discovery stability, AI clarity, and manual deactivation ([db27507](https://github.com/steierma/cogni-cash/commit/db275076130750631441b555d6a186cfd668b299))
* **subscription-management:** implement full subscription lifecycle including discovery, AI enrichment, and one-click cancellation ([87cebcd](https://github.com/steierma/cogni-cash/commit/87cebcd71bb8b35380aea9ff0c006a476e1c8dc6))
* **subscriptions:** implement ability to permanently decline subscription suggestions ([d949cd0](https://github.com/steierma/cogni-cash/commit/d949cd0d70a2affae7046a06c69b2ee07ec965cc))
* **subscriptions:** implement ability to undecline/restore ignored subscription suggestions ([7cded9f](https://github.com/steierma/cogni-cash/commit/7cded9f573f46e9f30b062821351c23062e38b87))
* **subscriptions:** implement manual subscription creation from transactions ([6dd324c](https://github.com/steierma/cogni-cash/commit/6dd324cbec463ee7dfe08e7e67a9960150028238))
* **subscriptions:** make discovery lookback period configurable ([2ad257c](https://github.com/steierma/cogni-cash/commit/2ad257ccfd0c25d3d11f7e74340a54ce8c2a0668))
* **subscriptions:** mandate-first discovery consolidation and automatic enrichment ([3bb7b65](https://github.com/steierma/cogni-cash/commit/3bb7b654228f106c096cd97bbebb6e0c9f27cb21))
* **subscriptions:** support bulk linking of transactions to subscriptions ([d16b304](https://github.com/steierma/cogni-cash/commit/d16b304d452fdfa06b80dbcaea38a74acaedb6d2))
* **testing:** add E2E tests for Payslips page and fix in-memory seeding ([6aae46d](https://github.com/steierma/cogni-cash/commit/6aae46dd2a702d8b0345055caaf1fc5bd3ddbee1))
* **testing:** implement E2E integration concept with Playwright and Forgejo CI ([e8dc94e](https://github.com/steierma/cogni-cash/commit/e8dc94e4f5010199cb8938ea0d89956693e3691e))
* **transaction:** improve categorisation and review all ([d291ed3](https://github.com/steierma/cogni-cash/commit/d291ed3eaff36c91e6b4eff1eeb3873988ea8d4a))
* **transactions:** add 'Search similar' feature to transaction list ([72fa4f0](https://github.com/steierma/cogni-cash/commit/72fa4f0c9267f11d50c3511024de96ff9d02e0f8))
* **types:** add subscription_id to Transaction interface and document manual creation concept ([7400d13](https://github.com/steierma/cogni-cash/commit/7400d1343aef7e5c82f092c415fa35b5340a424f))


### Bug Fixes

* **api:** correct MIME type detection for XLS bank statements ([87fc2f3](https://github.com/steierma/cogni-cash/commit/87fc2f3f3a1c3fbb15b8f2daad62cc0d0f09f091))
* **backend:** correct NewInvoiceService call in main.go to match new signature ([c2078bc](https://github.com/steierma/cogni-cash/commit/c2078bc0733ef9c1e096f8ae6d7fa10bb070caa2))
* **backend:** handle parser correctly to return specific error if vw parser fails ([342b20c](https://github.com/steierma/cogni-cash/commit/342b20c5519fe0e739fd3ca64688bf902e785ab8))
* **backend:** import strings ([0581978](https://github.com/steierma/cogni-cash/commit/05819788f18ff3f96323f3b50625b514c3c8466b))
* **backend:** improve subscription discovery filtering to check both description and counterparty ([eab0af0](https://github.com/steierma/cogni-cash/commit/eab0af075c634f85cd5052080c310c78ddc87515))
* **backend:** refine forecasting burn-rate and subscription reconciliation ([75d1533](https://github.com/steierma/cogni-cash/commit/75d1533938a170f5d13b6f7311ad57631e954cd7))
* **bank_statement:** improve logging for parsing and validation failures ([4eca776](https://github.com/steierma/cogni-cash/commit/4eca7768032a528299490c669ae8dd4928fe5ff3))
* **ci:** fix backend startup in E2E and increase test timeouts ([bf332c3](https://github.com/steierma/cogni-cash/commit/bf332c3c702bd91fe7451ae813d9a7fca049ad15))
* **ci:** restore rsync-based deployment and enable local builds ([f47e182](https://github.com/steierma/cogni-cash/commit/f47e182b5f18efcc623e53b92873908a8fd7b928))
* **core:** improve reconciliation accuracy and subscription discovery ([f4ee995](https://github.com/steierma/cogni-cash/commit/f4ee995a483f9fc3f6374d5de9d60850f35e5fa4))
* **discovery:** improve subscription discovery robustness and fix LLM adapter panic ([39c882d](https://github.com/steierma/cogni-cash/commit/39c882d96036a133224a61ba0869b6125ff92ea5))
* **forecasting:** align api paths for planned transactions and improve document sorting ([0541744](https://github.com/steierma/cogni-cash/commit/05417447749c02bfe60c4e04f57984a6f962b2ee))
* **forecasting:** fix saving of recurrence fields and improve soft suppression logic ([7613959](https://github.com/steierma/cogni-cash/commit/7613959b343b953ec53d1f0dc5c3f086c0bb0adc))
* forward client IP and User-Agent to Enable Banking for improved bank compatibility ([d5fd196](https://github.com/steierma/cogni-cash/commit/d5fd1963d5bba83675b44c13734ae11b5cda30dd))
* **frontend:** add missing RefreshCcw import in TransactionsPage ([65c5bab](https://github.com/steierma/cogni-cash/commit/65c5bab0d104a1320ac66fdcac8f6d2e1264c79b))
* **frontend:** implement robust local date formatting to fix timezone boundary issues in forecasting and reconciliation ([43f12f1](https://github.com/steierma/cogni-cash/commit/43f12f1d5e42dd3a1b9187182a67b8846f3e1ae7))
* **frontend:** import LucideIcon as type to fix SyntaxError ([abb0611](https://github.com/steierma/cogni-cash/commit/abb0611fb77edc6da4c1d6064e05fdb96a0b8c25))
* **frontend:** resolve all typescript and syntax errors (type-only imports and casting) ([be13785](https://github.com/steierma/cogni-cash/commit/be1378532be8efeedecc3b88a00c415dfa2da2ba))
* **frontend:** resolve build errors in SubscriptionsPage ([5f925d0](https://github.com/steierma/cogni-cash/commit/5f925d00c282ec210540ac2eb730b1ac99729f19))
* **i18n:** add missing translation keys for forecasting and settings ([eb54d1f](https://github.com/steierma/cogni-cash/commit/eb54d1ff378fd34d13755b2d4bac2e3e769ede18))
* **ingcsv:** correctly terminate metadata loop when header is reached ([3c03bde](https://github.com/steierma/cogni-cash/commit/3c03bde36e58011cd01a81693761bb9d88295cde))
* **invoices:** allow clearing all splits by sending an empty array instead of undefined ([cbb144f](https://github.com/steierma/cogni-cash/commit/cbb144f9c42ef283428c02a4898256fd7e8589b2))
* **invoices:** fix 400 Bad Request on invoice update by ensuring ISO date format and filtering splits ([65195f1](https://github.com/steierma/cogni-cash/commit/65195f173dca1da389011ff76071f4d9389b43c8))
* **mobile:** add missing Category import in forecast_view ([728ebe9](https://github.com/steierma/cogni-cash/commit/728ebe9aa40f365fa8cf8b1326975ebe3825f330))
* **multi-currency:** provide drop-down for currency ([a57d57e](https://github.com/steierma/cogni-cash/commit/a57d57e880bf2ad3ba642986fc57eb9e3ec17fa3))
* **reconcile:** align JSON payload keys with backend and improve error feedback ([e23d7ed](https://github.com/steierma/cogni-cash/commit/e23d7edef19e2704cc964b6bc683c862a4398946))
* remove unused import in vault_check script to fix CI build ([8773a22](https://github.com/steierma/cogni-cash/commit/8773a226385492886f8c5371c46f1938b9333545))
* resolve test failures across full stack and implement resilient document decryption ([a2e6efa](https://github.com/steierma/cogni-cash/commit/a2e6efa3b617d9081d1b5d6283ac123150de6a90))
* **service:** resolve build error in forecasting and align tests with discovery tolerances ([99f275f](https://github.com/steierma/cogni-cash/commit/99f275f5bfedf164ce00809f90dd314e569f9d85))
* **subscription-management:** log entries for historical activity ([a4e6387](https://github.com/steierma/cogni-cash/commit/a4e63877765756a4e252ea7f7b57df85d6d6803e))
* **subscription-management:** smaller fixes of approving subscriptions ([6f85e35](https://github.com/steierma/cogni-cash/commit/6f85e3598854314e4a74d29c9acdd3fd9ceeeff5))
* **subscription:** correct sign for manually created subscriptions ([c412dc9](https://github.com/steierma/cogni-cash/commit/c412dc9814796e60a1555cf3ad9611a3fc6adc4d))
* **subscription:** correct spelling to 'cancelled' and improve manual creation reliability ([8e1913c](https://github.com/steierma/cogni-cash/commit/8e1913c2975e56be8ba6ef88d5b33cb7c815eb18))
* **subscriptions:** prevent duplicate suggestions after approval via hash-based filtering and improved backfill ([1d705bb](https://github.com/steierma/cogni-cash/commit/1d705bb851f38f1509a29feb142c45d3b2fa1250))
* **ui:** ensure subscription suggestion tooltips are visible and opaque ([4bbdfdf](https://github.com/steierma/cogni-cash/commit/4bbdfdf8810150aae24579544556aec2ae680d4a))
* **vw-parser:** return FormatMismatch on non-PDF files to allow parser chain to continue ([a337fdf](https://github.com/steierma/cogni-cash/commit/a337fdfccbe09b2c455d88e1db836e6f812762c6))


### Styles

* align manual subscription creation icon with main menu (RefreshCcw) ([bcc5b9c](https://github.com/steierma/cogni-cash/commit/bcc5b9c3a4af0f50abe93d851652b9e5077fd777))


### Tests

* anonymize discovery service test data ([e43948c](https://github.com/steierma/cogni-cash/commit/e43948cbc9bcdc16ef41251b31908b0ef362ddfc))
* **backend:** significantly improve test coverage and consolidate redundant tests ([881f0d7](https://github.com/steierma/cogni-cash/commit/881f0d740138533bc0913e1dc5c8c32870f7e0f8))
* fix mockDiscoveryUseCase to implement LinkTransactions ([4f72f97](https://github.com/steierma/cogni-cash/commit/4f72f97b8d4f5fba127e553e5e5f45a375a529dd))
* **settings:** fix tests and improve coverage for hardened settings service ([b490cba](https://github.com/steierma/cogni-cash/commit/b490cba1e42adb63d14e994816649e098d1b6fd7))


### Code Refactoring

* **frontend:** fix 66 lint errors (types, cascading renders, unused vars) ([4cfbb45](https://github.com/steierma/cogni-cash/commit/4cfbb4530bcbfa32e6fa2c9d4fa655925a3d9884))
* **parser:** split VW bank parser into PDF and CSV variants ([2d74b8e](https://github.com/steierma/cogni-cash/commit/2d74b8e9d13af966bf08c8d98050fa27720f4c8a))
* unify docker image names and fix backend build network timeout ([3a29025](https://github.com/steierma/cogni-cash/commit/3a29025e4b211bd038cf53216d931041a6ee8334))


### Documentation

* add concept for shared bank account forecasting and virtual accounts ([01e30b4](https://github.com/steierma/cogni-cash/commit/01e30b4fc7c7e9163c057fd13eb6b725f51064cd))
* **core:** reorganize documentation and update architecture navigation ([1c42cfa](https://github.com/steierma/cogni-cash/commit/1c42cfac44c6cfa38a108624dab7e11b17909cd4))
* finalize windows installation guide title ([15a3e16](https://github.com/steierma/cogni-cash/commit/15a3e16e4ba4725d88955400cef8de730f63b63b))
* **stories:** add implementation roadmap for Subscription Management ([c5454fc](https://github.com/steierma/cogni-cash/commit/c5454fc7f4ce200f6fb2d4a94ac2687180f4294f))
* synchronize documentation for collaborative forecasting and virtual accounts ([2ade1c7](https://github.com/steierma/cogni-cash/commit/2ade1c7fecca9a589aadb1ba5a6f6d7590a03719))
* synchronize memory, readme, and schema documentation ([2b46101](https://github.com/steierma/cogni-cash/commit/2b46101b4d8b6a41fa2fc2aa3f014827afcdf5f2))
* update MEMORY.md with asynchronous approval info ([12f35f9](https://github.com/steierma/cogni-cash/commit/12f35f923ef949e07530190378e1f28430010692))
* update MEMORY.md with coverage improvements ([73e3397](https://github.com/steierma/cogni-cash/commit/73e3397eaef08ae560420aed2e734bf64b6be986))
* update MEMORY.md with graceful shutdown improvements ([78882d7](https://github.com/steierma/cogni-cash/commit/78882d79ced2cf571b6f0a254bead9b09f45cec0))
* update MEMORY.md with invoice split logic improvements ([0b619d6](https://github.com/steierma/cogni-cash/commit/0b619d601050e441713e533821f681cacb297f2d))
* update MEMORY.md with invoice update bug fix info ([184083b](https://github.com/steierma/cogni-cash/commit/184083b2d1d51d447e77ac10de5cc735ea132150))
* update MEMORY.md with split clearing fix info ([3365c5e](https://github.com/steierma/cogni-cash/commit/3365c5efed984e16305383528abfc85a76d6d1cf))
* update MEMORY.md with subscription discovery loading state ([edfc6d7](https://github.com/steierma/cogni-cash/commit/edfc6d730576fc2ed74bf1523f421be1111f7344))
* update MEMORY.md with XLS download fix ([f578b63](https://github.com/steierma/cogni-cash/commit/f578b639b45ddd9b89bb565661f648b8cb6bdfe3))


### Maintenance

* **backen:** add tests for coverage ([542b87c](https://github.com/steierma/cogni-cash/commit/542b87cbdfe7902a1d02f2b39423011d47702616))
* **backen:** add tests for coverage ([8a93036](https://github.com/steierma/cogni-cash/commit/8a93036af116c0d5bd8075478c3f5b2c7bb4a0e2))
* **ci:** refactor to separate public-release workflow ([a325384](https://github.com/steierma/cogni-cash/commit/a3253840eaf0efa6d2387fdf463facc7285680b7))
* **docs:** update MEMORY.md for public release v2.3.0 ([5de0538](https://github.com/steierma/cogni-cash/commit/5de05382dc72dd212903ad022206cd23d1cd704b))
* **makefile:** add test-e2e target ([83e6651](https://github.com/steierma/cogni-cash/commit/83e665192e70a6352d002bd6d3b1f1509248ba10))
* **migrations:** squash migrations ([133a871](https://github.com/steierma/cogni-cash/commit/133a871ee1df4b37e14735971b09fb9d97da6c89))
* **mobile:** finalize mobile app for sharing information ([a0b9ea8](https://github.com/steierma/cogni-cash/commit/a0b9ea8cafeed5aa85eb25d24a2726859eb44e6f))
* **mobile:** start to fix mobile app for sharing ([ff722af](https://github.com/steierma/cogni-cash/commit/ff722af07a0104ece1b29a609f6f194d46d1eb99))
* **release:** 2.1.0 ([9d32fc6](https://github.com/steierma/cogni-cash/commit/9d32fc676a771bef00d9e1ca3da218fe4e6e2578))
* **release:** 2.2.0 ([22bb85c](https://github.com/steierma/cogni-cash/commit/22bb85c5b671909da76f4f9f439b2fb0b8b9fe8e))
* **release:** 2.3.0 ([9c038db](https://github.com/steierma/cogni-cash/commit/9c038dbbbf836e112ea1646754f35e55b98786cd))
* **release:** 2.4.0 ([80b6692](https://github.com/steierma/cogni-cash/commit/80b669251c7fdb065bcb32bb497f0349a8763e03))
* **release:** bump version to 2.0.1 ([a600726](https://github.com/steierma/cogni-cash/commit/a600726878dcee57169aedfe81b4dc37723a6677))
* **security:** add penetration-test agent ([fdfeb16](https://github.com/steierma/cogni-cash/commit/fdfeb16fe78cb93fb87f53401b21bb5acb1e9123))
* **skills:** add public-release-snapshotter skill ([2c6fef1](https://github.com/steierma/cogni-cash/commit/2c6fef155d955da98210b672893e0d813cc5c148))

## [2.4.0](https://github.com/steierma/cogni-cash/compare/v2.0.0...v2.4.0) (2026-04-24)


### Features

* add dynamic log level control and enhance bank connection debugging ([59a74c9](https://github.com/steierma/cogni-cash/commit/59a74c9f1f11ecdebe4ac7c0a30d838999caae1a))
* add Frontend Type Guardian skill and mandate to prevent TypeScript syntax errors ([12d5cff](https://github.com/steierma/cogni-cash/commit/12d5cff84dc1a13418dd1025a461c9d551ce356a))
* **backend:** implement robust graceful shutdown and background worker tracking ([b4159c3](https://github.com/steierma/cogni-cash/commit/b4159c3d14808c2fcb97128b5e909988ea09e04d))
* **backend:** support AES-encrypted PDFs for corporate payslips and bank statements ([d27804c](https://github.com/steierma/cogni-cash/commit/d27804c35c53cafc8006ab98f1b01251c0320232))
* **bank-statements:** exhaustive parsing and localized high-signal error handling ([f1059fc](https://github.com/steierma/cogni-cash/commit/f1059fc9eea770a6b2486c3d60a40da2f86e345e))
* **bank:** enhance bank connection UX and fix multi-tenancy account isolation ([bf330bd](https://github.com/steierma/cogni-cash/commit/bf330bd946f4f7e5c7d06232e6cbf5c2c17f98b3))
* **bank:** implement counterparty IBAN and transaction code mapping for Enable Banking ([8137052](https://github.com/steierma/cogni-cash/commit/813705257c248288d2ffdc06b605a96ae07281de))
* **categories:** make it possible to change the forecast based on the historical data by changing the time period for the past ([965bcef](https://github.com/steierma/cogni-cash/commit/965bcefe09fcddae541ac997863bb4037d4586ef))
* **cogni-cash:** multi currency support ([020d637](https://github.com/steierma/cogni-cash/commit/020d637200e52d5531e5968e73026ad7d1d8a477))
* **discovery:** consolidate subscription discovery engine and implement strict name-based deduplication ([9003880](https://github.com/steierma/cogni-cash/commit/9003880f28b362f2433a3ef7cad767c49bf351b0))
* **discovery:** implement user-configurable discovery tolerances and UI settings ([6d6e21c](https://github.com/steierma/cogni-cash/commit/6d6e21c2459def8751706f64af67139658e0e496))
* **docs:** add Subscription Management & One-Click Cancellation concept ([f5144c7](https://github.com/steierma/cogni-cash/commit/f5144c7e3833cd282bbea3f891b5ca23414f091b))
* **documents:** implement document vault and refactor frontend api into modular services ([b66d438](https://github.com/steierma/cogni-cash/commit/b66d438f62f927dbcb755e054012c2394d8f4a27))
* forecast engine simplification + subscription-driven projections ([f72c258](https://github.com/steierma/cogni-cash/commit/f72c258950b64cd76df44ccb85ecbf75ff494321))
* **forecasting:** implement last bank day scheduling for planned transactions ([a311d3e](https://github.com/steierma/cogni-cash/commit/a311d3e53aae6e8a0e713ac8357588913b760bb4))
* **forecasting:** implement recurring manual transactions and smart auto-suppression ([4c15e9d](https://github.com/steierma/cogni-cash/commit/4c15e9d0cf00235cda4fca987b66ff89a958194d))
* **forecasting:** small improvements for smartphones ([10beb4e](https://github.com/steierma/cogni-cash/commit/10beb4e610c5ed7d1b5a9b429549b8412ebdfe9e))
* **frontend:** add actual vs. burn-rate comparison and 24-month forecast range ([9e68a44](https://github.com/steierma/cogni-cash/commit/9e68a447041ea90bbeb761fab9465da22afd1657))
* **frontend:** add loading state for subscription discovery ([3c2e187](https://github.com/steierma/cogni-cash/commit/3c2e18706ce1e6f11924da58d62c1806bfe1779a))
* **frontend:** refactor SubscriptionsPage layout and split active/canceled tables ([977f9c8](https://github.com/steierma/cogni-cash/commit/977f9c8244b48e01982e548e1984db098065cff8))
* implement batch transaction linking and improve forecasting accuracy ([fdbd3d9](https://github.com/steierma/cogni-cash/commit/fdbd3d9aed444639166add757f8008d8e17f352b))
* implement editable billing intervals and enhance bank connection documentation ([cdbe0bb](https://github.com/steierma/cogni-cash/commit/cdbe0bb05ecb9b022bc94f57b331199a30d0077f))
* **ingcsv:** enhance parser with explicit statement typing and joint account support ([d88204d](https://github.com/steierma/cogni-cash/commit/d88204d191f0eabbfb8791b116ccc0a189c6727d))
* **invoices:** attribute remaining unsplit amount to main category in breakdown ([90dd5b2](https://github.com/steierma/cogni-cash/commit/90dd5b2ee7eb30141417e5610c8bf656653098bf))
* **invoices:** implement multi-category split tracking and update documentation ([623a873](https://github.com/steierma/cogni-cash/commit/623a8733be6c22c9b93f871e0c5e5ca542690279))
* make subscription approval asynchronous with background AI enrichment ([3431a9d](https://github.com/steierma/cogni-cash/commit/3431a9dd5c917e7e6412f4d91db3b7597a8935e7))
* **mobile:** align menu and features with web frontend ([171fab0](https://github.com/steierma/cogni-cash/commit/171fab044519ee76585b4c8e6ac8e69202ca9837))
* **mobile:** implement subscription parity and fix soft-delete category leak in analytics ([41e57bb](https://github.com/steierma/cogni-cash/commit/41e57bb80073dbc99c1c8126b3245516bea24b0c))
* **multi-currency:** decouple AI prompts, implement dynamic UI formatting, and localize subscription enrichment ([7108fd8](https://github.com/steierma/cogni-cash/commit/7108fd8c7cc741402c41a6d85824fae0f016e81b))
* **security:** enforce role-based access for sensitive settings and bank integration ([f76b63a](https://github.com/steierma/cogni-cash/commit/f76b63aec01e327c8da068c4186058e33be4ff4f))
* **settings:** harden settings service with encryption and masking ([09a84d9](https://github.com/steierma/cogni-cash/commit/09a84d9c8343ae93af5663948bbca489db63d1de))
* **sharing:** add backend, frontend and first mobile draft ([a4937e7](https://github.com/steierma/cogni-cash/commit/a4937e71bbcfe6fbafad1aa78a238a3207fcb419))
* **sharing:** implement bank account centric collaborative forecasting and virtual accounts ([19131bb](https://github.com/steierma/cogni-cash/commit/19131bbcebe40c5884f7719dce1133b136cec270))
* **subscription-management:** broad backfill, activity log fix, cancellation UI fix, and extended field editing ([87312b0](https://github.com/steierma/cogni-cash/commit/87312b0f5212b202535762a3ac8a063d367d3f96))
* **subscription-management:** enhance discovery stability, AI clarity, and manual deactivation ([db27507](https://github.com/steierma/cogni-cash/commit/db275076130750631441b555d6a186cfd668b299))
* **subscription-management:** implement full subscription lifecycle including discovery, AI enrichment, and one-click cancellation ([87cebcd](https://github.com/steierma/cogni-cash/commit/87cebcd71bb8b35380aea9ff0c006a476e1c8dc6))
* **subscriptions:** implement ability to permanently decline subscription suggestions ([d949cd0](https://github.com/steierma/cogni-cash/commit/d949cd0d70a2affae7046a06c69b2ee07ec965cc))
* **subscriptions:** implement ability to undecline/restore ignored subscription suggestions ([7cded9f](https://github.com/steierma/cogni-cash/commit/7cded9f573f46e9f30b062821351c23062e38b87))
* **subscriptions:** implement manual subscription creation from transactions ([6dd324c](https://github.com/steierma/cogni-cash/commit/6dd324cbec463ee7dfe08e7e67a9960150028238))
* **subscriptions:** make discovery lookback period configurable ([2ad257c](https://github.com/steierma/cogni-cash/commit/2ad257ccfd0c25d3d11f7e74340a54ce8c2a0668))
* **subscriptions:** mandate-first discovery consolidation and automatic enrichment ([3bb7b65](https://github.com/steierma/cogni-cash/commit/3bb7b654228f106c096cd97bbebb6e0c9f27cb21))
* **subscriptions:** support bulk linking of transactions to subscriptions ([d16b304](https://github.com/steierma/cogni-cash/commit/d16b304d452fdfa06b80dbcaea38a74acaedb6d2))
* **testing:** add E2E tests for Payslips page and fix in-memory seeding ([6aae46d](https://github.com/steierma/cogni-cash/commit/6aae46dd2a702d8b0345055caaf1fc5bd3ddbee1))
* **testing:** implement E2E integration concept with Playwright and Forgejo CI ([e8dc94e](https://github.com/steierma/cogni-cash/commit/e8dc94e4f5010199cb8938ea0d89956693e3691e))
* **transaction:** improve categorisation and review all ([d291ed3](https://github.com/steierma/cogni-cash/commit/d291ed3eaff36c91e6b4eff1eeb3873988ea8d4a))
* **transactions:** add 'Search similar' feature to transaction list ([72fa4f0](https://github.com/steierma/cogni-cash/commit/72fa4f0c9267f11d50c3511024de96ff9d02e0f8))
* **types:** add subscription_id to Transaction interface and document manual creation concept ([7400d13](https://github.com/steierma/cogni-cash/commit/7400d1343aef7e5c82f092c415fa35b5340a424f))


### Bug Fixes

* **api:** correct MIME type detection for XLS bank statements ([87fc2f3](https://github.com/steierma/cogni-cash/commit/87fc2f3f3a1c3fbb15b8f2daad62cc0d0f09f091))
* **backend:** correct NewInvoiceService call in main.go to match new signature ([c2078bc](https://github.com/steierma/cogni-cash/commit/c2078bc0733ef9c1e096f8ae6d7fa10bb070caa2))
* **backend:** handle parser correctly to return specific error if vw parser fails ([342b20c](https://github.com/steierma/cogni-cash/commit/342b20c5519fe0e739fd3ca64688bf902e785ab8))
* **backend:** import strings ([0581978](https://github.com/steierma/cogni-cash/commit/05819788f18ff3f96323f3b50625b514c3c8466b))
* **backend:** improve subscription discovery filtering to check both description and counterparty ([eab0af0](https://github.com/steierma/cogni-cash/commit/eab0af075c634f85cd5052080c310c78ddc87515))
* **backend:** refine forecasting burn-rate and subscription reconciliation ([75d1533](https://github.com/steierma/cogni-cash/commit/75d1533938a170f5d13b6f7311ad57631e954cd7))
* **bank_statement:** improve logging for parsing and validation failures ([4eca776](https://github.com/steierma/cogni-cash/commit/4eca7768032a528299490c669ae8dd4928fe5ff3))
* **ci:** fix backend startup in E2E and increase test timeouts ([bf332c3](https://github.com/steierma/cogni-cash/commit/bf332c3c702bd91fe7451ae813d9a7fca049ad15))
* **ci:** restore rsync-based deployment and enable local builds ([f47e182](https://github.com/steierma/cogni-cash/commit/f47e182b5f18efcc623e53b92873908a8fd7b928))
* **core:** improve reconciliation accuracy and subscription discovery ([f4ee995](https://github.com/steierma/cogni-cash/commit/f4ee995a483f9fc3f6374d5de9d60850f35e5fa4))
* **discovery:** improve subscription discovery robustness and fix LLM adapter panic ([39c882d](https://github.com/steierma/cogni-cash/commit/39c882d96036a133224a61ba0869b6125ff92ea5))
* **forecasting:** align api paths for planned transactions and improve document sorting ([0541744](https://github.com/steierma/cogni-cash/commit/05417447749c02bfe60c4e04f57984a6f962b2ee))
* **forecasting:** fix saving of recurrence fields and improve soft suppression logic ([7613959](https://github.com/steierma/cogni-cash/commit/7613959b343b953ec53d1f0dc5c3f086c0bb0adc))
* forward client IP and User-Agent to Enable Banking for improved bank compatibility ([d5fd196](https://github.com/steierma/cogni-cash/commit/d5fd1963d5bba83675b44c13734ae11b5cda30dd))
* **frontend:** add missing RefreshCcw import in TransactionsPage ([65c5bab](https://github.com/steierma/cogni-cash/commit/65c5bab0d104a1320ac66fdcac8f6d2e1264c79b))
* **frontend:** implement robust local date formatting to fix timezone boundary issues in forecasting and reconciliation ([43f12f1](https://github.com/steierma/cogni-cash/commit/43f12f1d5e42dd3a1b9187182a67b8846f3e1ae7))
* **frontend:** import LucideIcon as type to fix SyntaxError ([abb0611](https://github.com/steierma/cogni-cash/commit/abb0611fb77edc6da4c1d6064e05fdb96a0b8c25))
* **frontend:** resolve all typescript and syntax errors (type-only imports and casting) ([be13785](https://github.com/steierma/cogni-cash/commit/be1378532be8efeedecc3b88a00c415dfa2da2ba))
* **frontend:** resolve build errors in SubscriptionsPage ([5f925d0](https://github.com/steierma/cogni-cash/commit/5f925d00c282ec210540ac2eb730b1ac99729f19))
* **i18n:** add missing translation keys for forecasting and settings ([eb54d1f](https://github.com/steierma/cogni-cash/commit/eb54d1ff378fd34d13755b2d4bac2e3e769ede18))
* **ingcsv:** correctly terminate metadata loop when header is reached ([3c03bde](https://github.com/steierma/cogni-cash/commit/3c03bde36e58011cd01a81693761bb9d88295cde))
* **invoices:** allow clearing all splits by sending an empty array instead of undefined ([cbb144f](https://github.com/steierma/cogni-cash/commit/cbb144f9c42ef283428c02a4898256fd7e8589b2))
* **invoices:** fix 400 Bad Request on invoice update by ensuring ISO date format and filtering splits ([65195f1](https://github.com/steierma/cogni-cash/commit/65195f173dca1da389011ff76071f4d9389b43c8))
* **mobile:** add missing Category import in forecast_view ([728ebe9](https://github.com/steierma/cogni-cash/commit/728ebe9aa40f365fa8cf8b1326975ebe3825f330))
* **multi-currency:** provide drop-down for currency ([a57d57e](https://github.com/steierma/cogni-cash/commit/a57d57e880bf2ad3ba642986fc57eb9e3ec17fa3))
* **reconcile:** align JSON payload keys with backend and improve error feedback ([e23d7ed](https://github.com/steierma/cogni-cash/commit/e23d7edef19e2704cc964b6bc683c862a4398946))
* **service:** resolve build error in forecasting and align tests with discovery tolerances ([99f275f](https://github.com/steierma/cogni-cash/commit/99f275f5bfedf164ce00809f90dd314e569f9d85))
* **subscription-management:** log entries for historical activity ([a4e6387](https://github.com/steierma/cogni-cash/commit/a4e63877765756a4e252ea7f7b57df85d6d6803e))
* **subscription-management:** smaller fixes of approving subscriptions ([6f85e35](https://github.com/steierma/cogni-cash/commit/6f85e3598854314e4a74d29c9acdd3fd9ceeeff5))
* **subscription:** correct sign for manually created subscriptions ([c412dc9](https://github.com/steierma/cogni-cash/commit/c412dc9814796e60a1555cf3ad9611a3fc6adc4d))
* **subscription:** correct spelling to 'cancelled' and improve manual creation reliability ([8e1913c](https://github.com/steierma/cogni-cash/commit/8e1913c2975e56be8ba6ef88d5b33cb7c815eb18))
* **subscriptions:** prevent duplicate suggestions after approval via hash-based filtering and improved backfill ([1d705bb](https://github.com/steierma/cogni-cash/commit/1d705bb851f38f1509a29feb142c45d3b2fa1250))
* **ui:** ensure subscription suggestion tooltips are visible and opaque ([4bbdfdf](https://github.com/steierma/cogni-cash/commit/4bbdfdf8810150aae24579544556aec2ae680d4a))
* **vw-parser:** return FormatMismatch on non-PDF files to allow parser chain to continue ([a337fdf](https://github.com/steierma/cogni-cash/commit/a337fdfccbe09b2c455d88e1db836e6f812762c6))


### Styles

* align manual subscription creation icon with main menu (RefreshCcw) ([bcc5b9c](https://github.com/steierma/cogni-cash/commit/bcc5b9c3a4af0f50abe93d851652b9e5077fd777))


### Maintenance

* **backen:** add tests for coverage ([542b87c](https://github.com/steierma/cogni-cash/commit/542b87cbdfe7902a1d02f2b39423011d47702616))
* **backen:** add tests for coverage ([8a93036](https://github.com/steierma/cogni-cash/commit/8a93036af116c0d5bd8075478c3f5b2c7bb4a0e2))
* **ci:** refactor to separate public-release workflow ([a325384](https://github.com/steierma/cogni-cash/commit/a3253840eaf0efa6d2387fdf463facc7285680b7))
* **docs:** update MEMORY.md for public release v2.3.0 ([5de0538](https://github.com/steierma/cogni-cash/commit/5de05382dc72dd212903ad022206cd23d1cd704b))
* **makefile:** add test-e2e target ([83e6651](https://github.com/steierma/cogni-cash/commit/83e665192e70a6352d002bd6d3b1f1509248ba10))
* **migrations:** squash migrations ([133a871](https://github.com/steierma/cogni-cash/commit/133a871ee1df4b37e14735971b09fb9d97da6c89))
* **mobile:** finalize mobile app for sharing information ([a0b9ea8](https://github.com/steierma/cogni-cash/commit/a0b9ea8cafeed5aa85eb25d24a2726859eb44e6f))
* **mobile:** start to fix mobile app for sharing ([ff722af](https://github.com/steierma/cogni-cash/commit/ff722af07a0104ece1b29a609f6f194d46d1eb99))
* **release:** 2.1.0 ([9d32fc6](https://github.com/steierma/cogni-cash/commit/9d32fc676a771bef00d9e1ca3da218fe4e6e2578))
* **release:** 2.2.0 ([22bb85c](https://github.com/steierma/cogni-cash/commit/22bb85c5b671909da76f4f9f439b2fb0b8b9fe8e))
* **release:** 2.3.0 ([9c038db](https://github.com/steierma/cogni-cash/commit/9c038dbbbf836e112ea1646754f35e55b98786cd))
* **release:** bump version to 2.0.1 ([a600726](https://github.com/steierma/cogni-cash/commit/a600726878dcee57169aedfe81b4dc37723a6677))
* **security:** add penetration-test agent ([fdfeb16](https://github.com/steierma/cogni-cash/commit/fdfeb16fe78cb93fb87f53401b21bb5acb1e9123))
* **skills:** add public-release-snapshotter skill ([2c6fef1](https://github.com/steierma/cogni-cash/commit/2c6fef155d955da98210b672893e0d813cc5c148))


### Tests

* anonymize discovery service test data ([e43948c](https://github.com/steierma/cogni-cash/commit/e43948cbc9bcdc16ef41251b31908b0ef362ddfc))
* **backend:** significantly improve test coverage and consolidate redundant tests ([881f0d7](https://github.com/steierma/cogni-cash/commit/881f0d740138533bc0913e1dc5c8c32870f7e0f8))
* fix mockDiscoveryUseCase to implement LinkTransactions ([4f72f97](https://github.com/steierma/cogni-cash/commit/4f72f97b8d4f5fba127e553e5e5f45a375a529dd))
* **settings:** fix tests and improve coverage for hardened settings service ([b490cba](https://github.com/steierma/cogni-cash/commit/b490cba1e42adb63d14e994816649e098d1b6fd7))


### Code Refactoring

* **frontend:** fix 66 lint errors (types, cascading renders, unused vars) ([4cfbb45](https://github.com/steierma/cogni-cash/commit/4cfbb4530bcbfa32e6fa2c9d4fa655925a3d9884))
* **parser:** split VW bank parser into PDF and CSV variants ([2d74b8e](https://github.com/steierma/cogni-cash/commit/2d74b8e9d13af966bf08c8d98050fa27720f4c8a))
* unify docker image names and fix backend build network timeout ([3a29025](https://github.com/steierma/cogni-cash/commit/3a29025e4b211bd038cf53216d931041a6ee8334))


### Documentation

* add concept for shared bank account forecasting and virtual accounts ([01e30b4](https://github.com/steierma/cogni-cash/commit/01e30b4fc7c7e9163c057fd13eb6b725f51064cd))
* **core:** reorganize documentation and update architecture navigation ([1c42cfa](https://github.com/steierma/cogni-cash/commit/1c42cfac44c6cfa38a108624dab7e11b17909cd4))
* finalize windows installation guide title ([15a3e16](https://github.com/steierma/cogni-cash/commit/15a3e16e4ba4725d88955400cef8de730f63b63b))
* **stories:** add implementation roadmap for Subscription Management ([c5454fc](https://github.com/steierma/cogni-cash/commit/c5454fc7f4ce200f6fb2d4a94ac2687180f4294f))
* synchronize documentation for collaborative forecasting and virtual accounts ([2ade1c7](https://github.com/steierma/cogni-cash/commit/2ade1c7fecca9a589aadb1ba5a6f6d7590a03719))
* synchronize memory, readme, and schema documentation ([2b46101](https://github.com/steierma/cogni-cash/commit/2b46101b4d8b6a41fa2fc2aa3f014827afcdf5f2))
* update MEMORY.md with asynchronous approval info ([12f35f9](https://github.com/steierma/cogni-cash/commit/12f35f923ef949e07530190378e1f28430010692))
* update MEMORY.md with coverage improvements ([73e3397](https://github.com/steierma/cogni-cash/commit/73e3397eaef08ae560420aed2e734bf64b6be986))
* update MEMORY.md with graceful shutdown improvements ([78882d7](https://github.com/steierma/cogni-cash/commit/78882d79ced2cf571b6f0a254bead9b09f45cec0))
* update MEMORY.md with invoice split logic improvements ([0b619d6](https://github.com/steierma/cogni-cash/commit/0b619d601050e441713e533821f681cacb297f2d))
* update MEMORY.md with invoice update bug fix info ([184083b](https://github.com/steierma/cogni-cash/commit/184083b2d1d51d447e77ac10de5cc735ea132150))
* update MEMORY.md with split clearing fix info ([3365c5e](https://github.com/steierma/cogni-cash/commit/3365c5efed984e16305383528abfc85a76d6d1cf))
* update MEMORY.md with subscription discovery loading state ([edfc6d7](https://github.com/steierma/cogni-cash/commit/edfc6d730576fc2ed74bf1523f421be1111f7344))
* update MEMORY.md with XLS download fix ([f578b63](https://github.com/steierma/cogni-cash/commit/f578b639b45ddd9b89bb565661f648b8cb6bdfe3))

## [2.3.0](https://github.com/steierma/cogni-cash/compare/v2.0.0...v2.3.0) (2026-04-21)


### Features

* add dynamic log level control and enhance bank connection debugging ([59a74c9](https://github.com/steierma/cogni-cash/commit/59a74c9f1f11ecdebe4ac7c0a30d838999caae1a))
* **backend:** implement robust graceful shutdown and background worker tracking ([b4159c3](https://github.com/steierma/cogni-cash/commit/b4159c3d14808c2fcb97128b5e909988ea09e04d))
* **backend:** support AES-encrypted PDFs for corporate payslips and bank statements ([d27804c](https://github.com/steierma/cogni-cash/commit/d27804c35c53cafc8006ab98f1b01251c0320232))
* **bank-statements:** exhaustive parsing and localized high-signal error handling ([f1059fc](https://github.com/steierma/cogni-cash/commit/f1059fc9eea770a6b2486c3d60a40da2f86e345e))
* **bank:** enhance bank connection UX and fix multi-tenancy account isolation ([bf330bd](https://github.com/steierma/cogni-cash/commit/bf330bd946f4f7e5c7d06232e6cbf5c2c17f98b3))
* **bank:** implement counterparty IBAN and transaction code mapping for Enable Banking ([8137052](https://github.com/steierma/cogni-cash/commit/813705257c248288d2ffdc06b605a96ae07281de))
* **categories:** make it possible to change the forecast based on the historical data by changing the time period for the past ([965bcef](https://github.com/steierma/cogni-cash/commit/965bcefe09fcddae541ac997863bb4037d4586ef))
* **cogni-cash:** multi currency support ([020d637](https://github.com/steierma/cogni-cash/commit/020d637200e52d5531e5968e73026ad7d1d8a477))
* **discovery:** consolidate subscription discovery engine and implement strict name-based deduplication ([9003880](https://github.com/steierma/cogni-cash/commit/9003880f28b362f2433a3ef7cad767c49bf351b0))
* **discovery:** implement user-configurable discovery tolerances and UI settings ([6d6e21c](https://github.com/steierma/cogni-cash/commit/6d6e21c2459def8751706f64af67139658e0e496))
* **docs:** add Subscription Management & One-Click Cancellation concept ([f5144c7](https://github.com/steierma/cogni-cash/commit/f5144c7e3833cd282bbea3f891b5ca23414f091b))
* **documents:** implement document vault and refactor frontend api into modular services ([b66d438](https://github.com/steierma/cogni-cash/commit/b66d438f62f927dbcb755e054012c2394d8f4a27))
* **forecasting:** implement recurring manual transactions and smart auto-suppression ([4c15e9d](https://github.com/steierma/cogni-cash/commit/4c15e9d0cf00235cda4fca987b66ff89a958194d))
* **frontend:** add loading state for subscription discovery ([3c2e187](https://github.com/steierma/cogni-cash/commit/3c2e18706ce1e6f11924da58d62c1806bfe1779a))
* **frontend:** refactor SubscriptionsPage layout and split active/canceled tables ([977f9c8](https://github.com/steierma/cogni-cash/commit/977f9c8244b48e01982e548e1984db098065cff8))
* implement editable billing intervals and enhance bank connection documentation ([cdbe0bb](https://github.com/steierma/cogni-cash/commit/cdbe0bb05ecb9b022bc94f57b331199a30d0077f))
* **ingcsv:** enhance parser with explicit statement typing and joint account support ([d88204d](https://github.com/steierma/cogni-cash/commit/d88204d191f0eabbfb8791b116ccc0a189c6727d))
* **invoices:** attribute remaining unsplit amount to main category in breakdown ([90dd5b2](https://github.com/steierma/cogni-cash/commit/90dd5b2ee7eb30141417e5610c8bf656653098bf))
* **invoices:** implement multi-category split tracking and update documentation ([623a873](https://github.com/steierma/cogni-cash/commit/623a8733be6c22c9b93f871e0c5e5ca542690279))
* make subscription approval asynchronous with background AI enrichment ([3431a9d](https://github.com/steierma/cogni-cash/commit/3431a9dd5c917e7e6412f4d91db3b7597a8935e7))
* **mobile:** align menu and features with web frontend ([171fab0](https://github.com/steierma/cogni-cash/commit/171fab044519ee76585b4c8e6ac8e69202ca9837))
* **mobile:** implement subscription parity and fix soft-delete category leak in analytics ([41e57bb](https://github.com/steierma/cogni-cash/commit/41e57bb80073dbc99c1c8126b3245516bea24b0c))
* **security:** enforce role-based access for sensitive settings and bank integration ([f76b63a](https://github.com/steierma/cogni-cash/commit/f76b63aec01e327c8da068c4186058e33be4ff4f))
* **settings:** harden settings service with encryption and masking ([09a84d9](https://github.com/steierma/cogni-cash/commit/09a84d9c8343ae93af5663948bbca489db63d1de))
* **sharing:** add backend, frontend and first mobile draft ([a4937e7](https://github.com/steierma/cogni-cash/commit/a4937e71bbcfe6fbafad1aa78a238a3207fcb419))
* **subscription-management:** broad backfill, activity log fix, cancellation UI fix, and extended field editing ([87312b0](https://github.com/steierma/cogni-cash/commit/87312b0f5212b202535762a3ac8a063d367d3f96))
* **subscription-management:** enhance discovery stability, AI clarity, and manual deactivation ([db27507](https://github.com/steierma/cogni-cash/commit/db275076130750631441b555d6a186cfd668b299))
* **subscription-management:** implement full subscription lifecycle including discovery, AI enrichment, and one-click cancellation ([87cebcd](https://github.com/steierma/cogni-cash/commit/87cebcd71bb8b35380aea9ff0c006a476e1c8dc6))
* **subscriptions:** implement ability to permanently decline subscription suggestions ([d949cd0](https://github.com/steierma/cogni-cash/commit/d949cd0d70a2affae7046a06c69b2ee07ec965cc))
* **subscriptions:** implement ability to undecline/restore ignored subscription suggestions ([7cded9f](https://github.com/steierma/cogni-cash/commit/7cded9f573f46e9f30b062821351c23062e38b87))
* **subscriptions:** implement manual subscription creation from transactions ([6dd324c](https://github.com/steierma/cogni-cash/commit/6dd324cbec463ee7dfe08e7e67a9960150028238))
* **subscriptions:** make discovery lookback period configurable ([2ad257c](https://github.com/steierma/cogni-cash/commit/2ad257ccfd0c25d3d11f7e74340a54ce8c2a0668))
* **subscriptions:** mandate-first discovery consolidation and automatic enrichment ([3bb7b65](https://github.com/steierma/cogni-cash/commit/3bb7b654228f106c096cd97bbebb6e0c9f27cb21))
* **types:** add subscription_id to Transaction interface and document manual creation concept ([7400d13](https://github.com/steierma/cogni-cash/commit/7400d1343aef7e5c82f092c415fa35b5340a424f))


### Bug Fixes

* **backend:** correct NewInvoiceService call in main.go to match new signature ([c2078bc](https://github.com/steierma/cogni-cash/commit/c2078bc0733ef9c1e096f8ae6d7fa10bb070caa2))
* **backend:** handle parser correctly to return specific error if vw parser fails ([342b20c](https://github.com/steierma/cogni-cash/commit/342b20c5519fe0e739fd3ca64688bf902e785ab8))
* **backend:** import strings ([0581978](https://github.com/steierma/cogni-cash/commit/05819788f18ff3f96323f3b50625b514c3c8466b))
* **backend:** improve subscription discovery filtering to check both description and counterparty ([eab0af0](https://github.com/steierma/cogni-cash/commit/eab0af075c634f85cd5052080c310c78ddc87515))
* **bank_statement:** improve logging for parsing and validation failures ([4eca776](https://github.com/steierma/cogni-cash/commit/4eca7768032a528299490c669ae8dd4928fe5ff3))
* **ci:** restore rsync-based deployment and enable local builds ([f47e182](https://github.com/steierma/cogni-cash/commit/f47e182b5f18efcc623e53b92873908a8fd7b928))
* **discovery:** improve subscription discovery robustness and fix LLM adapter panic ([39c882d](https://github.com/steierma/cogni-cash/commit/39c882d96036a133224a61ba0869b6125ff92ea5))
* **forecasting:** align api paths for planned transactions and improve document sorting ([0541744](https://github.com/steierma/cogni-cash/commit/05417447749c02bfe60c4e04f57984a6f962b2ee))
* **forecasting:** fix saving of recurrence fields and improve soft suppression logic ([7613959](https://github.com/steierma/cogni-cash/commit/7613959b343b953ec53d1f0dc5c3f086c0bb0adc))
* forward client IP and User-Agent to Enable Banking for improved bank compatibility ([d5fd196](https://github.com/steierma/cogni-cash/commit/d5fd1963d5bba83675b44c13734ae11b5cda30dd))
* **frontend:** add missing RefreshCcw import in TransactionsPage ([65c5bab](https://github.com/steierma/cogni-cash/commit/65c5bab0d104a1320ac66fdcac8f6d2e1264c79b))
* **frontend:** resolve build errors in SubscriptionsPage ([5f925d0](https://github.com/steierma/cogni-cash/commit/5f925d00c282ec210540ac2eb730b1ac99729f19))
* **ingcsv:** correctly terminate metadata loop when header is reached ([3c03bde](https://github.com/steierma/cogni-cash/commit/3c03bde36e58011cd01a81693761bb9d88295cde))
* **invoices:** allow clearing all splits by sending an empty array instead of undefined ([cbb144f](https://github.com/steierma/cogni-cash/commit/cbb144f9c42ef283428c02a4898256fd7e8589b2))
* **invoices:** fix 400 Bad Request on invoice update by ensuring ISO date format and filtering splits ([65195f1](https://github.com/steierma/cogni-cash/commit/65195f173dca1da389011ff76071f4d9389b43c8))
* **mobile:** add missing Category import in forecast_view ([728ebe9](https://github.com/steierma/cogni-cash/commit/728ebe9aa40f365fa8cf8b1326975ebe3825f330))
* **multi-currency:** provide drop-down for currency ([a57d57e](https://github.com/steierma/cogni-cash/commit/a57d57e880bf2ad3ba642986fc57eb9e3ec17fa3))
* **service:** resolve build error in forecasting and align tests with discovery tolerances ([99f275f](https://github.com/steierma/cogni-cash/commit/99f275f5bfedf164ce00809f90dd314e569f9d85))
* **subscription-management:** log entries for historical activity ([a4e6387](https://github.com/steierma/cogni-cash/commit/a4e63877765756a4e252ea7f7b57df85d6d6803e))
* **subscription-management:** smaller fixes of approving subscriptions ([6f85e35](https://github.com/steierma/cogni-cash/commit/6f85e3598854314e4a74d29c9acdd3fd9ceeeff5))
* **subscription:** correct spelling to 'cancelled' and improve manual creation reliability ([8e1913c](https://github.com/steierma/cogni-cash/commit/8e1913c2975e56be8ba6ef88d5b33cb7c815eb18))
* **subscriptions:** prevent duplicate suggestions after approval via hash-based filtering and improved backfill ([1d705bb](https://github.com/steierma/cogni-cash/commit/1d705bb851f38f1509a29feb142c45d3b2fa1250))
* **ui:** ensure subscription suggestion tooltips are visible and opaque ([4bbdfdf](https://github.com/steierma/cogni-cash/commit/4bbdfdf8810150aae24579544556aec2ae680d4a))
* **vw-parser:** return FormatMismatch on non-PDF files to allow parser chain to continue ([a337fdf](https://github.com/steierma/cogni-cash/commit/a337fdfccbe09b2c455d88e1db836e6f812762c6))


### Styles

* align manual subscription creation icon with main menu (RefreshCcw) ([bcc5b9c](https://github.com/steierma/cogni-cash/commit/bcc5b9c3a4af0f50abe93d851652b9e5077fd777))


### Maintenance

* **backen:** add tests for coverage ([542b87c](https://github.com/steierma/cogni-cash/commit/542b87cbdfe7902a1d02f2b39423011d47702616))
* **backen:** add tests for coverage ([8a93036](https://github.com/steierma/cogni-cash/commit/8a93036af116c0d5bd8075478c3f5b2c7bb4a0e2))
* **ci:** refactor to separate public-release workflow ([a325384](https://github.com/steierma/cogni-cash/commit/a3253840eaf0efa6d2387fdf463facc7285680b7))
* **migrations:** squash migrations ([133a871](https://github.com/steierma/cogni-cash/commit/133a871ee1df4b37e14735971b09fb9d97da6c89))
* **mobile:** finalize mobile app for sharing information ([a0b9ea8](https://github.com/steierma/cogni-cash/commit/a0b9ea8cafeed5aa85eb25d24a2726859eb44e6f))
* **mobile:** start to fix mobile app for sharing ([ff722af](https://github.com/steierma/cogni-cash/commit/ff722af07a0104ece1b29a609f6f194d46d1eb99))
* **release:** 2.1.0 ([9d32fc6](https://github.com/steierma/cogni-cash/commit/9d32fc676a771bef00d9e1ca3da218fe4e6e2578))
* **release:** 2.2.0 ([22bb85c](https://github.com/steierma/cogni-cash/commit/22bb85c5b671909da76f4f9f439b2fb0b8b9fe8e))
* **release:** bump version to 2.0.1 ([a600726](https://github.com/steierma/cogni-cash/commit/a600726878dcee57169aedfe81b4dc37723a6677))
* **security:** add penetration-test agent ([fdfeb16](https://github.com/steierma/cogni-cash/commit/fdfeb16fe78cb93fb87f53401b21bb5acb1e9123))
* **skills:** add public-release-snapshotter skill ([2c6fef1](https://github.com/steierma/cogni-cash/commit/2c6fef155d955da98210b672893e0d813cc5c148))


### Tests

* anonymize discovery service test data ([e43948c](https://github.com/steierma/cogni-cash/commit/e43948cbc9bcdc16ef41251b31908b0ef362ddfc))
* **backend:** significantly improve test coverage and consolidate redundant tests ([881f0d7](https://github.com/steierma/cogni-cash/commit/881f0d740138533bc0913e1dc5c8c32870f7e0f8))
* **settings:** fix tests and improve coverage for hardened settings service ([b490cba](https://github.com/steierma/cogni-cash/commit/b490cba1e42adb63d14e994816649e098d1b6fd7))


### Documentation

* **core:** reorganize documentation and update architecture navigation ([1c42cfa](https://github.com/steierma/cogni-cash/commit/1c42cfac44c6cfa38a108624dab7e11b17909cd4))
* finalize windows installation guide title ([15a3e16](https://github.com/steierma/cogni-cash/commit/15a3e16e4ba4725d88955400cef8de730f63b63b))
* **stories:** add implementation roadmap for Subscription Management ([c5454fc](https://github.com/steierma/cogni-cash/commit/c5454fc7f4ce200f6fb2d4a94ac2687180f4294f))
* synchronize memory, readme, and schema documentation ([2b46101](https://github.com/steierma/cogni-cash/commit/2b46101b4d8b6a41fa2fc2aa3f014827afcdf5f2))
* update MEMORY.md with asynchronous approval info ([12f35f9](https://github.com/steierma/cogni-cash/commit/12f35f923ef949e07530190378e1f28430010692))
* update MEMORY.md with coverage improvements ([73e3397](https://github.com/steierma/cogni-cash/commit/73e3397eaef08ae560420aed2e734bf64b6be986))
* update MEMORY.md with graceful shutdown improvements ([78882d7](https://github.com/steierma/cogni-cash/commit/78882d79ced2cf571b6f0a254bead9b09f45cec0))
* update MEMORY.md with invoice split logic improvements ([0b619d6](https://github.com/steierma/cogni-cash/commit/0b619d601050e441713e533821f681cacb297f2d))
* update MEMORY.md with invoice update bug fix info ([184083b](https://github.com/steierma/cogni-cash/commit/184083b2d1d51d447e77ac10de5cc735ea132150))
* update MEMORY.md with split clearing fix info ([3365c5e](https://github.com/steierma/cogni-cash/commit/3365c5efed984e16305383528abfc85a76d6d1cf))
* update MEMORY.md with subscription discovery loading state ([edfc6d7](https://github.com/steierma/cogni-cash/commit/edfc6d730576fc2ed74bf1523f421be1111f7344))

## [2.2.0](https://github.com/steierma/cogni-cash/compare/v2.0.0...v2.2.0) (2026-04-19)


### Features

* **backend:** support AES-encrypted PDFs for corporate payslips and bank statements ([d27804c](https://github.com/steierma/cogni-cash/commit/d27804c35c53cafc8006ab98f1b01251c0320232))
* **bank:** enhance bank connection UX and fix multi-tenancy account isolation ([bf330bd](https://github.com/steierma/cogni-cash/commit/bf330bd946f4f7e5c7d06232e6cbf5c2c17f98b3))
* **bank:** implement counterparty IBAN and transaction code mapping for Enable Banking ([8137052](https://github.com/steierma/cogni-cash/commit/813705257c248288d2ffdc06b605a96ae07281de))
* **categories:** make it possible to change the forecast based on the historical data by changing the time period for the past ([965bcef](https://github.com/steierma/cogni-cash/commit/965bcefe09fcddae541ac997863bb4037d4586ef))
* **discovery:** consolidate subscription discovery engine and implement strict name-based deduplication ([9003880](https://github.com/steierma/cogni-cash/commit/9003880f28b362f2433a3ef7cad767c49bf351b0))
* **discovery:** implement user-configurable discovery tolerances and UI settings ([6d6e21c](https://github.com/steierma/cogni-cash/commit/6d6e21c2459def8751706f64af67139658e0e496))
* **docs:** add Subscription Management & One-Click Cancellation concept ([f5144c7](https://github.com/steierma/cogni-cash/commit/f5144c7e3833cd282bbea3f891b5ca23414f091b))
* **documents:** implement document vault and refactor frontend api into modular services ([b66d438](https://github.com/steierma/cogni-cash/commit/b66d438f62f927dbcb755e054012c2394d8f4a27))
* **forecasting:** implement recurring manual transactions and smart auto-suppression ([4c15e9d](https://github.com/steierma/cogni-cash/commit/4c15e9d0cf00235cda4fca987b66ff89a958194d))
* **mobile:** implement subscription parity and fix soft-delete category leak in analytics ([41e57bb](https://github.com/steierma/cogni-cash/commit/41e57bb80073dbc99c1c8126b3245516bea24b0c))
* **security:** enforce role-based access for sensitive settings and bank integration ([f76b63a](https://github.com/steierma/cogni-cash/commit/f76b63aec01e327c8da068c4186058e33be4ff4f))
* **settings:** harden settings service with encryption and masking ([09a84d9](https://github.com/steierma/cogni-cash/commit/09a84d9c8343ae93af5663948bbca489db63d1de))
* **sharing:** add backend, frontend and first mobile draft ([a4937e7](https://github.com/steierma/cogni-cash/commit/a4937e71bbcfe6fbafad1aa78a238a3207fcb419))
* **subscription-management:** broad backfill, activity log fix, cancellation UI fix, and extended field editing ([87312b0](https://github.com/steierma/cogni-cash/commit/87312b0f5212b202535762a3ac8a063d367d3f96))
* **subscription-management:** enhance discovery stability, AI clarity, and manual deactivation ([db27507](https://github.com/steierma/cogni-cash/commit/db275076130750631441b555d6a186cfd668b299))
* **subscription-management:** implement full subscription lifecycle including discovery, AI enrichment, and one-click cancellation ([87cebcd](https://github.com/steierma/cogni-cash/commit/87cebcd71bb8b35380aea9ff0c006a476e1c8dc6))
* **subscriptions:** implement ability to permanently decline subscription suggestions ([d949cd0](https://github.com/steierma/cogni-cash/commit/d949cd0d70a2affae7046a06c69b2ee07ec965cc))
* **subscriptions:** implement ability to undecline/restore ignored subscription suggestions ([7cded9f](https://github.com/steierma/cogni-cash/commit/7cded9f573f46e9f30b062821351c23062e38b87))
* **subscriptions:** implement manual subscription creation from transactions ([6dd324c](https://github.com/steierma/cogni-cash/commit/6dd324cbec463ee7dfe08e7e67a9960150028238))
* **subscriptions:** make discovery lookback period configurable ([2ad257c](https://github.com/steierma/cogni-cash/commit/2ad257ccfd0c25d3d11f7e74340a54ce8c2a0668))
* **types:** add subscription_id to Transaction interface and document manual creation concept ([7400d13](https://github.com/steierma/cogni-cash/commit/7400d1343aef7e5c82f092c415fa35b5340a424f))


### Bug Fixes

* **backend:** improve subscription discovery filtering to check both description and counterparty ([eab0af0](https://github.com/steierma/cogni-cash/commit/eab0af075c634f85cd5052080c310c78ddc87515))
* **ci:** restore rsync-based deployment and enable local builds ([f47e182](https://github.com/steierma/cogni-cash/commit/f47e182b5f18efcc623e53b92873908a8fd7b928))
* **discovery:** improve subscription discovery robustness and fix LLM adapter panic ([39c882d](https://github.com/steierma/cogni-cash/commit/39c882d96036a133224a61ba0869b6125ff92ea5))
* **forecasting:** align api paths for planned transactions and improve document sorting ([0541744](https://github.com/steierma/cogni-cash/commit/05417447749c02bfe60c4e04f57984a6f962b2ee))
* **forecasting:** fix saving of recurrence fields and improve soft suppression logic ([7613959](https://github.com/steierma/cogni-cash/commit/7613959b343b953ec53d1f0dc5c3f086c0bb0adc))
* **frontend:** add missing RefreshCcw import in TransactionsPage ([65c5bab](https://github.com/steierma/cogni-cash/commit/65c5bab0d104a1320ac66fdcac8f6d2e1264c79b))
* **frontend:** resolve build errors in SubscriptionsPage ([5f925d0](https://github.com/steierma/cogni-cash/commit/5f925d00c282ec210540ac2eb730b1ac99729f19))
* **mobile:** add missing Category import in forecast_view ([728ebe9](https://github.com/steierma/cogni-cash/commit/728ebe9aa40f365fa8cf8b1326975ebe3825f330))
* **service:** resolve build error in forecasting and align tests with discovery tolerances ([99f275f](https://github.com/steierma/cogni-cash/commit/99f275f5bfedf164ce00809f90dd314e569f9d85))
* **subscription-management:** log entries for historical activity ([a4e6387](https://github.com/steierma/cogni-cash/commit/a4e63877765756a4e252ea7f7b57df85d6d6803e))
* **subscription-management:** smaller fixes of approving subscriptions ([6f85e35](https://github.com/steierma/cogni-cash/commit/6f85e3598854314e4a74d29c9acdd3fd9ceeeff5))
* **ui:** ensure subscription suggestion tooltips are visible and opaque ([4bbdfdf](https://github.com/steierma/cogni-cash/commit/4bbdfdf8810150aae24579544556aec2ae680d4a))


### Tests

* **settings:** fix tests and improve coverage for hardened settings service ([b490cba](https://github.com/steierma/cogni-cash/commit/b490cba1e42adb63d14e994816649e098d1b6fd7))


### Documentation

* **core:** reorganize documentation and update architecture navigation ([1c42cfa](https://github.com/steierma/cogni-cash/commit/1c42cfac44c6cfa38a108624dab7e11b17909cd4))
* finalize windows installation guide title ([15a3e16](https://github.com/steierma/cogni-cash/commit/15a3e16e4ba4725d88955400cef8de730f63b63b))
* **stories:** add implementation roadmap for Subscription Management ([c5454fc](https://github.com/steierma/cogni-cash/commit/c5454fc7f4ce200f6fb2d4a94ac2687180f4294f))
* synchronize memory, readme, and schema documentation ([2b46101](https://github.com/steierma/cogni-cash/commit/2b46101b4d8b6a41fa2fc2aa3f014827afcdf5f2))


### Styles

* align manual subscription creation icon with main menu (RefreshCcw) ([bcc5b9c](https://github.com/steierma/cogni-cash/commit/bcc5b9c3a4af0f50abe93d851652b9e5077fd777))


### Maintenance

* **backen:** add tests for coverage ([542b87c](https://github.com/steierma/cogni-cash/commit/542b87cbdfe7902a1d02f2b39423011d47702616))
* **backen:** add tests for coverage ([8a93036](https://github.com/steierma/cogni-cash/commit/8a93036af116c0d5bd8075478c3f5b2c7bb4a0e2))
* **ci:** refactor to separate public-release workflow ([a325384](https://github.com/steierma/cogni-cash/commit/a3253840eaf0efa6d2387fdf463facc7285680b7))
* **mobile:** finalize mobile app for sharing information ([a0b9ea8](https://github.com/steierma/cogni-cash/commit/a0b9ea8cafeed5aa85eb25d24a2726859eb44e6f))
* **mobile:** start to fix mobile app for sharing ([ff722af](https://github.com/steierma/cogni-cash/commit/ff722af07a0104ece1b29a609f6f194d46d1eb99))
* **release:** 2.1.0 ([9d32fc6](https://github.com/steierma/cogni-cash/commit/9d32fc676a771bef00d9e1ca3da218fe4e6e2578))
* **release:** bump version to 2.0.1 ([a600726](https://github.com/steierma/cogni-cash/commit/a600726878dcee57169aedfe81b4dc37723a6677))
* **security:** add penetration-test agent ([fdfeb16](https://github.com/steierma/cogni-cash/commit/fdfeb16fe78cb93fb87f53401b21bb5acb1e9123))
* **skills:** add public-release-snapshotter skill ([2c6fef1](https://github.com/steierma/cogni-cash/commit/2c6fef155d955da98210b672893e0d813cc5c148))

## [2.1.0](https://github.com/steierma/cogni-cash/compare/v2.0.0...v2.1.0) (2026-04-19)


### Features

* **backend:** support AES-encrypted PDFs for corporate payslips and bank statements ([d27804c](https://github.com/steierma/cogni-cash/commit/d27804c35c53cafc8006ab98f1b01251c0320232))
* **bank:** enhance bank connection UX and fix multi-tenancy account isolation ([bf330bd](https://github.com/steierma/cogni-cash/commit/bf330bd946f4f7e5c7d06232e6cbf5c2c17f98b3))
* **bank:** implement counterparty IBAN and transaction code mapping for Enable Banking ([8137052](https://github.com/steierma/cogni-cash/commit/813705257c248288d2ffdc06b605a96ae07281de))
* **categories:** make it possible to change the forecast based on the historical data by changing the time period for the past ([965bcef](https://github.com/steierma/cogni-cash/commit/965bcefe09fcddae541ac997863bb4037d4586ef))
* **discovery:** consolidate subscription discovery engine and implement strict name-based deduplication ([9003880](https://github.com/steierma/cogni-cash/commit/9003880f28b362f2433a3ef7cad767c49bf351b0))
* **discovery:** implement user-configurable discovery tolerances and UI settings ([6d6e21c](https://github.com/steierma/cogni-cash/commit/6d6e21c2459def8751706f64af67139658e0e496))
* **docs:** add Subscription Management & One-Click Cancellation concept ([f5144c7](https://github.com/steierma/cogni-cash/commit/f5144c7e3833cd282bbea3f891b5ca23414f091b))
* **documents:** implement document vault and refactor frontend api into modular services ([b66d438](https://github.com/steierma/cogni-cash/commit/b66d438f62f927dbcb755e054012c2394d8f4a27))
* **forecasting:** implement recurring manual transactions and smart auto-suppression ([4c15e9d](https://github.com/steierma/cogni-cash/commit/4c15e9d0cf00235cda4fca987b66ff89a958194d))
* **mobile:** implement subscription parity and fix soft-delete category leak in analytics ([41e57bb](https://github.com/steierma/cogni-cash/commit/41e57bb80073dbc99c1c8126b3245516bea24b0c))
* **security:** enforce role-based access for sensitive settings and bank integration ([f76b63a](https://github.com/steierma/cogni-cash/commit/f76b63aec01e327c8da068c4186058e33be4ff4f))
* **settings:** harden settings service with encryption and masking ([09a84d9](https://github.com/steierma/cogni-cash/commit/09a84d9c8343ae93af5663948bbca489db63d1de))
* **sharing:** add backend, frontend and first mobile draft ([a4937e7](https://github.com/steierma/cogni-cash/commit/a4937e71bbcfe6fbafad1aa78a238a3207fcb419))
* **subscription-management:** broad backfill, activity log fix, cancellation UI fix, and extended field editing ([87312b0](https://github.com/steierma/cogni-cash/commit/87312b0f5212b202535762a3ac8a063d367d3f96))
* **subscription-management:** enhance discovery stability, AI clarity, and manual deactivation ([db27507](https://github.com/steierma/cogni-cash/commit/db275076130750631441b555d6a186cfd668b299))
* **subscription-management:** implement full subscription lifecycle including discovery, AI enrichment, and one-click cancellation ([87cebcd](https://github.com/steierma/cogni-cash/commit/87cebcd71bb8b35380aea9ff0c006a476e1c8dc6))
* **subscriptions:** implement ability to permanently decline subscription suggestions ([d949cd0](https://github.com/steierma/cogni-cash/commit/d949cd0d70a2affae7046a06c69b2ee07ec965cc))
* **subscriptions:** implement ability to undecline/restore ignored subscription suggestions ([7cded9f](https://github.com/steierma/cogni-cash/commit/7cded9f573f46e9f30b062821351c23062e38b87))
* **subscriptions:** implement manual subscription creation from transactions ([6dd324c](https://github.com/steierma/cogni-cash/commit/6dd324cbec463ee7dfe08e7e67a9960150028238))
* **subscriptions:** make discovery lookback period configurable ([2ad257c](https://github.com/steierma/cogni-cash/commit/2ad257ccfd0c25d3d11f7e74340a54ce8c2a0668))
* **types:** add subscription_id to Transaction interface and document manual creation concept ([7400d13](https://github.com/steierma/cogni-cash/commit/7400d1343aef7e5c82f092c415fa35b5340a424f))


### Bug Fixes

* **ci:** restore rsync-based deployment and enable local builds ([f47e182](https://github.com/steierma/cogni-cash/commit/f47e182b5f18efcc623e53b92873908a8fd7b928))
* **discovery:** improve subscription discovery robustness and fix LLM adapter panic ([39c882d](https://github.com/steierma/cogni-cash/commit/39c882d96036a133224a61ba0869b6125ff92ea5))
* **forecasting:** align api paths for planned transactions and improve document sorting ([0541744](https://github.com/steierma/cogni-cash/commit/05417447749c02bfe60c4e04f57984a6f962b2ee))
* **forecasting:** fix saving of recurrence fields and improve soft suppression logic ([7613959](https://github.com/steierma/cogni-cash/commit/7613959b343b953ec53d1f0dc5c3f086c0bb0adc))
* **frontend:** add missing RefreshCcw import in TransactionsPage ([65c5bab](https://github.com/steierma/cogni-cash/commit/65c5bab0d104a1320ac66fdcac8f6d2e1264c79b))
* **frontend:** resolve build errors in SubscriptionsPage ([5f925d0](https://github.com/steierma/cogni-cash/commit/5f925d00c282ec210540ac2eb730b1ac99729f19))
* **mobile:** add missing Category import in forecast_view ([728ebe9](https://github.com/steierma/cogni-cash/commit/728ebe9aa40f365fa8cf8b1326975ebe3825f330))
* **service:** resolve build error in forecasting and align tests with discovery tolerances ([99f275f](https://github.com/steierma/cogni-cash/commit/99f275f5bfedf164ce00809f90dd314e569f9d85))
* **subscription-management:** log entries for historical activity ([a4e6387](https://github.com/steierma/cogni-cash/commit/a4e63877765756a4e252ea7f7b57df85d6d6803e))
* **subscription-management:** smaller fixes of approving subscriptions ([6f85e35](https://github.com/steierma/cogni-cash/commit/6f85e3598854314e4a74d29c9acdd3fd9ceeeff5))
* **ui:** ensure subscription suggestion tooltips are visible and opaque ([4bbdfdf](https://github.com/steierma/cogni-cash/commit/4bbdfdf8810150aae24579544556aec2ae680d4a))


### Tests

* **settings:** fix tests and improve coverage for hardened settings service ([b490cba](https://github.com/steierma/cogni-cash/commit/b490cba1e42adb63d14e994816649e098d1b6fd7))


### Maintenance

* **backen:** add tests for coverage ([542b87c](https://github.com/steierma/cogni-cash/commit/542b87cbdfe7902a1d02f2b39423011d47702616))
* **backen:** add tests for coverage ([8a93036](https://github.com/steierma/cogni-cash/commit/8a93036af116c0d5bd8075478c3f5b2c7bb4a0e2))
* **ci:** refactor to separate public-release workflow ([a325384](https://github.com/steierma/cogni-cash/commit/a3253840eaf0efa6d2387fdf463facc7285680b7))
* **mobile:** finalize mobile app for sharing information ([a0b9ea8](https://github.com/steierma/cogni-cash/commit/a0b9ea8cafeed5aa85eb25d24a2726859eb44e6f))
* **mobile:** start to fix mobile app for sharing ([ff722af](https://github.com/steierma/cogni-cash/commit/ff722af07a0104ece1b29a609f6f194d46d1eb99))
* **release:** bump version to 2.0.1 ([a600726](https://github.com/steierma/cogni-cash/commit/a600726878dcee57169aedfe81b4dc37723a6677))
* **security:** add penetration-test agent ([fdfeb16](https://github.com/steierma/cogni-cash/commit/fdfeb16fe78cb93fb87f53401b21bb5acb1e9123))
* **skills:** add public-release-snapshotter skill ([2c6fef1](https://github.com/steierma/cogni-cash/commit/2c6fef155d955da98210b672893e0d813cc5c148))


### Documentation

* **core:** reorganize documentation and update architecture navigation ([1c42cfa](https://github.com/steierma/cogni-cash/commit/1c42cfac44c6cfa38a108624dab7e11b17909cd4))
* finalize windows installation guide title ([15a3e16](https://github.com/steierma/cogni-cash/commit/15a3e16e4ba4725d88955400cef8de730f63b63b))
* **stories:** add implementation roadmap for Subscription Management ([c5454fc](https://github.com/steierma/cogni-cash/commit/c5454fc7f4ce200f6fb2d4a94ac2687180f4294f))
* synchronize memory, readme, and schema documentation ([2b46101](https://github.com/steierma/cogni-cash/commit/2b46101b4d8b6a41fa2fc2aa3f014827afcdf5f2))


### Styles

* align manual subscription creation icon with main menu (RefreshCcw) ([bcc5b9c](https://github.com/steierma/cogni-cash/commit/bcc5b9c3a4af0f50abe93d851652b9e5077fd777))
