# Changelog

All notable changes to this project will be documented in this file. See [standard-version](https://github.com/conventional-changelog/standard-version) for commit guidelines.

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
