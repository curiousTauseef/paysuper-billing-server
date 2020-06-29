# Changelog
All notable changes to this project will be documented in this file.

## [1.3.0] - 2020-06-23

### Added
- New onboarding flow. ([7570382](https://github.com/paysuper/paysuper-billing-server/commit/75703824c00062f9a1614351e62201e469b9d4b9))
- Add a command to recalculate payouts sums. ([3763968](https://github.com/paysuper/paysuper-billing-server/commit/376396836cd5268aa064c0bf4dc9a25495afce0f))
- A new method to calculate amounts in a payout. ([b12379c](https://github.com/paysuper/paysuper-billing-server/commit/b12379cace9415a005442f46c926e93268e4c376))
- A method to get a file content type changes. ([2905027](https://github.com/paysuper/paysuper-billing-server/commit/2905027ec56d314838c03e4623eef24ded5cab3d))
- A project status changed to production to check if a merchant hasn't completed onboarding. ([a749602](https://github.com/paysuper/paysuper-billing-server/commit/a74960294650a47ac16d9c640606a8d2e4da0f99))
- A new method to send financial reports to accountants. ([01ae855](https://github.com/paysuper/paysuper-billing-server/tree/01ae855b2499742fd276a506bdec6e768d025222))
- A new field in order_view_private with a payment method terminal ID. ([63112bf](https://github.com/paysuper/paysuper-billing-server/tree/63112bff1e93abfa59b1ebca018acf0be2f3abe5))
- Generate VAT reports for all countries. ([8741abe](https://github.com/paysuper/paysuper-billing-server/tree/8741abe58d5354c45a7f8eb6da4fad5a5245a932))
- Refund creation available for system admins. ([4a468ff](https://github.com/paysuper/paysuper-billing-server/tree/4a468ffbe301949a568c4f41719c1e736d501de5))
- Generate accounting entries only for orders with a real cash flow. ([bcf28db](https://github.com/paysuper/paysuper-billing-server/tree/bcf28db3b69532d04e5e69899e1facf88808ab65))
- A new grpc method to get an admin user by user ID. ([053beb7](https://github.com/paysuper/paysuper-billing-server/tree/053beb7ae14a4b174c390231d1c98299d4d9191c))
- Add indexes for orders filters. ([d1a5c59](https://github.com/paysuper/paysuper-billing-server/tree/d1a5c59ae3c800ad4368204676fa71a6e9f72361))
- Create an order metadata processing. ([188a2a5](https://github.com/paysuper/paysuper-billing-server/tree/188a2a509bb7a068a2fe152ac716f1ad4cf3cd48))
- A project metadata in an order create request. ([3c8fad5](https://github.com/paysuper/paysuper-billing-server/tree/3c8fad5b2f7a1a13503b0cc631f53f5fb308b86e))
- Added exchange directions support to the currency service mock. ([5e0d5d0](https://github.com/paysuper/paysuper-billing-server/tree/5e0d5d0f49e7137f0f0aa58b8d7d0ee4b215acca))
- Information about the minimal payout amounts. ([7744de5](https://github.com/paysuper/paysuper-billing-server/tree/7744de51aa37d48991bd05fd3333be01b9c82a49))
- Save webhook testing results. ([1df962e](https://github.com/paysuper/paysuper-billing-server/tree/1df962e09c8fe5579d1665443d440415fb571f64))

### Changed
- Royalty reports corrections changes. ([635610b](https://github.com/paysuper/paysuper-billing-server/commit/635610b0dad30521783188408313007172cf8c12)) ([bcdc4af](https://github.com/paysuper/paysuper-billing-server/commit/bcdc4af73619e2326f2dd785929600d3e41d4601))
- Correct rounding amounts of the order view public. ([63ada2e](https://github.com/paysuper/paysuper-billing-server/commit/63ada2e5f495218657ffc4f976ec31eb21d6b776))
- Rebuild a payout calculation. ([288d316](https://github.com/paysuper/paysuper-billing-server/commit/288d3167ede8387e1b2e8bf6d5ec9ca619d5ff13)) ([3fcb6ee](https://github.com/paysuper/paysuper-billing-server/commit/3fcb6ee800324d6c7c302c90861c4faa1cee37f6)) ([1084ace](https://github.com/paysuper/paysuper-billing-server/commit/1084acea8d9778b717071597dae24f0c2af157a6)) ([a02c4ed](https://github.com/paysuper/paysuper-billing-server/commit/a02c4ed150339c3bd059427f60451a5d4190e342))
- Balance calculation. ([e2fdc5a](https://github.com/paysuper/paysuper-billing-server/commit/e2fdc5a8e46182d693c89b0337e9c13c2645c16b))
- Update a merchant balance after a payout creation. ([bfebd1c](https://github.com/paysuper/paysuper-billing-server/commit/bfebd1c39c7fad40b2123e76fad595ea23c97cf3))
- Update dependencies. ([5c5b6df](https://github.com/paysuper/paysuper-billing-server/commit/5c5b6dfbeef854319c25bdc7877406c9d29fd108))
- Onboarding refactoring. ([87e35a7](https://github.com/paysuper/paysuper-billing-server/commit/87e35a7beb5f9cd783ef293e62c3688ef7d60e14))
- The royalty and payout reports refactoring. ([0bcacc4](https://github.com/paysuper/paysuper-billing-server/tree/0bcacc49358d6863fdec72bd0990ee936307a25d))
- The date filters refactoring. ([0e69aef](https://github.com/paysuper/paysuper-billing-server/tree/0e69aef8bdd3da9ca51f9a79c4105d9e5a152eb6))
- Add a sync mutex to a webhook. ([9e70d77](https://github.com/paysuper/paysuper-billing-server/tree/9e70d773c0471428c689b452606b8c4f85961e41))
- A merchant default processing currency. ([c6e9147](https://github.com/paysuper/paysuper-billing-server/tree/c6e914751457c74104302261168688a05029d7de))
- The URL for the royalty reports in an email. ([52c0991](https://github.com/paysuper/paysuper-billing-server/tree/52c0991819940321231680fd998e80cb3768ef33))
- The default mode for the project redirect settings. ([0ecc415](https://github.com/paysuper/paysuper-billing-server/tree/0ecc415641a52c956fab70ded4de77a515ab804c))

### Fixed
- An empty company name. ([225ab98](https://github.com/paysuper/paysuper-billing-server/commit/225ab9875501c44568925b5476d9bd79ab1cab7d))
- Sending am email with a receipt. ([eb6a350](https://github.com/paysuper/paysuper-billing-server/commit/eb6a350b90a57caab706d9ced842c571f116d1e6)) ([20fbd5d](https://github.com/paysuper/paysuper-billing-server/commit/20fbd5d6346e6edc4a0dce93afc362b10e9f7ee1))
- A merchant status. ([4f3a370](https://github.com/paysuper/paysuper-billing-server/commit/4f3a370abe920347d110907811bfc7fc993bf394))
- Sending an email about a payout invoice to a financier. ([2c5a20a](https://github.com/paysuper/paysuper-billing-server/commit/2c5a20a73dd2d21eeab84c5dec89ca1a80543fa4))
- Not saved tariffs in a merchant object. ([8e98b74](https://github.com/paysuper/paysuper-billing-server/commit/8e98b74ac95d937417aa210533668c52f8323f1b))
- The transaction log search by a customer's account. ([62dcd2b](https://github.com/paysuper/paysuper-billing-server/tree/62dcd2b248ef46268daef377cceda5c768b5d43a))
- Notification statuses. ([fe9be52](https://github.com/paysuper/paysuper-billing-server/tree/fe9be528a15db4f254f146ae0867b6561cb6f228))
- A payment link building. ([9b543c4](https://github.com/paysuper/paysuper-billing-server/tree/9b543c491cc4eae076cd9b89f39cabb04f993c44))
- Fees total local. ([d59e59b](https://github.com/paysuper/paysuper-billing-server/tree/d59e59b8ff296c36a98457b7245e383024a55869))
- Round a balance. ([acd73b8](https://github.com/paysuper/paysuper-billing-server/tree/acd73b8debc2a9186e28a83fca9b67501e25f643))

### Removed
- Remove a royalty ID from a select orders query. ([076fbcd](https://github.com/paysuper/paysuper-billing-server/commit/076fbcda6c21103ce36a185c30c2eaee9ee824f6))
- Migration to remove an order.project_account field. ([885e136](https://github.com/paysuper/paysuper-billing-server/commit/885e1366114b5551ddad1ffa6ac058f5d340e673))
- Remove a project account field in an order object. ([b71bc7a](https://github.com/paysuper/paysuper-billing-server/commit/b71bc7ae534cbd78e63c12b204b606e1a44d749b))

***

## [1.2.0] - 2020-02-04

### Added
- Checking connection with the Redis server, the Redis cluster and database using the health check request.
- The new parameter to filter the transactions log by the production or test mode.
- The new parameter to display the customer's email address in the payment receipt.
- The new project's settings group to redirect user at the end of the payment process.
- The new settings option to update the text on the redirect button displaying at the end of the payment process when creating the payment token.
- Added checking to disable the partial refunds.

### Changed
- Changed the context lifetime created with the health check request.
- Some entities moved to the external repository (Price group, User profile, ZIP code, Country, User role).
- Removed unused code.
- Updated GO and Alpine Linux versions in the dockerfile
- Update project's dependencies.

### Fixed
- Fixed the automatically payouts method.
- Fixed the order's structure tags.
- Edited the descriptions of the structures.

***

## [1.1.0] - 2019-12-24

### Added
- The centrifugo has been split into multiple instances for sending notification to the Dashboard and the payment form.
- Added an API method for VAT calculating in a payment process.

### Changed
- Changed the Project settings for VAT calculation. Added some options: to disable VAT for a customer in a payment process, to include VAT in a total payment amount.
- Changed the card number checking for the China UnionPay card validation.
- Update project's dependencies.

### Removed
- Removed the file `.gitlab-ci.yml`.

### Fixed
- Fix for a customer's country detection if it had not been determined by a user's IP address.

***

## [1.0.0] - 2019-12-19

### Added
- Limiting payments by country depending on the country issuing of the customer's bank card.
- The logic of the rounding method for a payment amount for various currencies considering the presence or absence of a currency's fractional part.

### Changed
- Added new response parameters when changing a language on a payment form.
- Added a project ID to payment form events' responses for sending data to web analytics services.
- Added a VAT parameter to a response for a rendering of a payment form.
- Added a merchant's legal name for onboarding process mails.
- Updated README.

### Fixed
- An order with products will be paid in a product's fallback currency if a customer's selected currency does not exist in a project for this product.
- Corrected minimum payments amounts for various currencies.
- Fix an order initialization for products with outdated project's settings.
- The purchase receipt letter sends only for a completed payment.
- Webhooks notifications send for all payments statuses including CANCEL and DECLINE.
- Fix for a payment form language selection via a user's locale in a token parameter.
- Edited the rounding method for a payment amount.

### Removed
- Deleted the unused source code.