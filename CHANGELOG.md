# Changelog

## [0.0.11](https://github.com/spectrocloud-labs/validator-plugin-vsphere/compare/v0.0.10...v0.0.11) (2023-10-20)


### Bug Fixes

* ct lints ([739b4a8](https://github.com/spectrocloud-labs/validator-plugin-vsphere/commit/739b4a80e8b0ba0723213e09b94ba1b7fd97ea2f))
* **deps:** update module sigs.k8s.io/controller-runtime to v0.16.3 ([#82](https://github.com/spectrocloud-labs/validator-plugin-vsphere/issues/82)) ([f031533](https://github.com/spectrocloud-labs/validator-plugin-vsphere/commit/f03153304f8c9a331242b255178749aef8b1fe48))


### Other

* **deps:** bump golang.org/x/net from 0.16.0 to 0.17.0 ([#74](https://github.com/spectrocloud-labs/validator-plugin-vsphere/issues/74)) ([a27698d](https://github.com/spectrocloud-labs/validator-plugin-vsphere/commit/a27698d6a0df19c892880dac7f83512f91d24e0e))
* **deps:** update actions/checkout digest to b4ffde6 ([#80](https://github.com/spectrocloud-labs/validator-plugin-vsphere/issues/80)) ([0b5d05d](https://github.com/spectrocloud-labs/validator-plugin-vsphere/commit/0b5d05d24901f870e66ac30d20a70216c3f99722))
* **deps:** update actions/setup-python digest to 65d7f2d ([#78](https://github.com/spectrocloud-labs/validator-plugin-vsphere/issues/78)) ([5652f1d](https://github.com/spectrocloud-labs/validator-plugin-vsphere/commit/5652f1d641f3f53253858703f5be857b35ac9dc9))
* **deps:** update gcr.io/kubebuilder/kube-rbac-proxy docker tag to v0.14.4 ([#71](https://github.com/spectrocloud-labs/validator-plugin-vsphere/issues/71)) ([db4676e](https://github.com/spectrocloud-labs/validator-plugin-vsphere/commit/db4676e5d4621da7910c1cbefda5f1ac414b0d5f))
* **deps:** update google-github-actions/release-please-action digest to 4c5670f ([#79](https://github.com/spectrocloud-labs/validator-plugin-vsphere/issues/79)) ([93fb4a8](https://github.com/spectrocloud-labs/validator-plugin-vsphere/commit/93fb4a8659650db5a45353830b1985dd1baae286))
* enable renovate automerges ([2639694](https://github.com/spectrocloud-labs/validator-plugin-vsphere/commit/2639694fac061f5e284f1fccc48a675b0738fdef))
* release 0.0.11 ([dffec26](https://github.com/spectrocloud-labs/validator-plugin-vsphere/commit/dffec2605f8cf5db33522a6c7ff772fd48e37193))


### Refactoring

* validator -&gt; validator ([#83](https://github.com/spectrocloud-labs/validator-plugin-vsphere/issues/83)) ([acf1f53](https://github.com/spectrocloud-labs/validator-plugin-vsphere/commit/acf1f53d94f209fd22da31fec62f34b5afee6b53))

## [0.0.10](https://github.com/spectrocloud-labs/validator-plugin-vsphere/compare/v0.0.9...v0.0.10) (2023-10-16)


### Bug Fixes

* Fix dockerfile for refactor of vsphere from internal to pkg ([#70](https://github.com/spectrocloud-labs/validator-plugin-vsphere/issues/70)) ([a08e02e](https://github.com/spectrocloud-labs/validator-plugin-vsphere/commit/a08e02eddfc428e47cd271cdd4b06408b3aff73c))
* move vsphere libs to pkg/ so they can be used by other projects ([#69](https://github.com/spectrocloud-labs/validator-plugin-vsphere/issues/69)) ([1bd8012](https://github.com/spectrocloud-labs/validator-plugin-vsphere/commit/1bd801235804e5100291b69cb7807d1b4b066c84))


### Other

* **deps:** update golang:1.21 docker digest to 02d7116 ([#67](https://github.com/spectrocloud-labs/validator-plugin-vsphere/issues/67)) ([d0a82c6](https://github.com/spectrocloud-labs/validator-plugin-vsphere/commit/d0a82c6d55c56b8a6b2bf04e01a78fe4d04cbe38))
* **deps:** update golang:1.21 docker digest to 24a0937 ([#68](https://github.com/spectrocloud-labs/validator-plugin-vsphere/issues/68)) ([5bd2734](https://github.com/spectrocloud-labs/validator-plugin-vsphere/commit/5bd27348f44f56c6daaebcc9ed7b11aed5374708))
* **deps:** update golang:1.21 docker digest to 4d5cf6c ([#65](https://github.com/spectrocloud-labs/validator-plugin-vsphere/issues/65)) ([0dacc3b](https://github.com/spectrocloud-labs/validator-plugin-vsphere/commit/0dacc3bf964b3e408c6a4a3ed4a989ad7442555e))

## [0.0.9](https://github.com/spectrocloud-labs/validator-plugin-vsphere/compare/v0.0.8...v0.0.9) (2023-10-10)


### Bug Fixes

* **deps:** update module github.com/onsi/ginkgo/v2 to v2.13.0 ([#61](https://github.com/spectrocloud-labs/validator-plugin-vsphere/issues/61)) ([f927c76](https://github.com/spectrocloud-labs/validator-plugin-vsphere/commit/f927c7660e8238a51f48abacf845e63180c3c2cb))
* **deps:** update module github.com/spectrocloud-labs/validator to v0.0.9 ([#64](https://github.com/spectrocloud-labs/validator-plugin-vsphere/issues/64)) ([fdb3027](https://github.com/spectrocloud-labs/validator-plugin-vsphere/commit/fdb30270d936f4cfa29d4bfaecd2d41e1356910c))


### Other

* better log messages in controller ([#60](https://github.com/spectrocloud-labs/validator-plugin-vsphere/issues/60)) ([2e64ce3](https://github.com/spectrocloud-labs/validator-plugin-vsphere/commit/2e64ce3dd1e3fd74e6e89bfc127ad72741b5e4a3))
* **deps:** update golang:1.21 docker digest to e9ebfe9 ([#42](https://github.com/spectrocloud-labs/validator-plugin-vsphere/issues/42)) ([204b045](https://github.com/spectrocloud-labs/validator-plugin-vsphere/commit/204b045a018fb93359bbbf099b3df7acc26db785))

## [0.0.8](https://github.com/spectrocloud-labs/validator-plugin-vsphere/compare/v0.0.7...v0.0.8) (2023-10-09)


### Bug Fixes

* release please comments in chart.yaml and values.yaml ([#58](https://github.com/spectrocloud-labs/validator-plugin-vsphere/issues/58)) ([6b0ad05](https://github.com/spectrocloud-labs/validator-plugin-vsphere/commit/6b0ad0550b6753721030ddc6a13ce119fa9ed2c3))

## [0.0.7](https://github.com/spectrocloud-labs/validator-plugin-vsphere/compare/v0.0.6...v0.0.7) (2023-10-09)


### Bug Fixes

* update charts and add proper templating for auth secret ([#55](https://github.com/spectrocloud-labs/validator-plugin-vsphere/issues/55)) ([3ba9c4b](https://github.com/spectrocloud-labs/validator-plugin-vsphere/commit/3ba9c4b2e4b9a1a00e659c81da4185f837b814fc))

## [0.0.6](https://github.com/spectrocloud-labs/validator-plugin-vsphere/compare/v0.0.5...v0.0.6) (2023-10-06)


### Bug Fixes

* yaml tag for auth secretName ([#53](https://github.com/spectrocloud-labs/validator-plugin-vsphere/issues/53)) ([cec752f](https://github.com/spectrocloud-labs/validator-plugin-vsphere/commit/cec752fa55f23748c5943a32d065142a4f41fabf))

## [0.0.5](https://github.com/spectrocloud-labs/validator-plugin-vsphere/compare/v0.0.4...v0.0.5) (2023-10-06)


### Bug Fixes

* fix yaml tag for nodepool cpu ([e21d8dc](https://github.com/spectrocloud-labs/validator-plugin-vsphere/commit/e21d8dcf04098d428732b348a5bf22f27092330e))


### Other

* release 0.0.4 ([#50](https://github.com/spectrocloud-labs/validator-plugin-vsphere/issues/50)) ([c419c9d](https://github.com/spectrocloud-labs/validator-plugin-vsphere/commit/c419c9d4e9298ee8127ad884c2d70d00aa3b5b87))
* release 0.0.5 ([#51](https://github.com/spectrocloud-labs/validator-plugin-vsphere/issues/51)) ([04508a8](https://github.com/spectrocloud-labs/validator-plugin-vsphere/commit/04508a88d66c6ea42aaa9162fc9e5c939dba7cf2))

## [0.0.4](https://github.com/spectrocloud-labs/validator-plugin-vsphere/compare/v0.0.3...v0.0.4) (2023-10-06)


### Bug Fixes

* fix generated code and temporarily disable tests ([#46](https://github.com/spectrocloud-labs/validator-plugin-vsphere/issues/46)) ([56cf9a7](https://github.com/spectrocloud-labs/validator-plugin-vsphere/commit/56cf9a715086f30fd952d98c449cd8df31dae6c0))


### Other

* Disable roleprivilege tests temporarily ([#48](https://github.com/spectrocloud-labs/validator-plugin-vsphere/issues/48)) ([3d4d736](https://github.com/spectrocloud-labs/validator-plugin-vsphere/commit/3d4d73622a6c0ab46b7cd288ed76a2558ad21bf9))

## [0.0.3](https://github.com/spectrocloud-labs/validator-plugin-vsphere/compare/v0.0.2...v0.0.3) (2023-10-06)


### Features

* Add support to validate arbitrary user's role and entity privileges instead of the one specified under auth secret ([#41](https://github.com/spectrocloud-labs/validator-plugin-vsphere/issues/41)) ([033f665](https://github.com/spectrocloud-labs/validator-plugin-vsphere/commit/033f665794dfadbd4d1473c7fdaed1242d7d0669))


### Other

* Add yaml tags to api types ([#44](https://github.com/spectrocloud-labs/validator-plugin-vsphere/issues/44)) ([1578a1f](https://github.com/spectrocloud-labs/validator-plugin-vsphere/commit/1578a1f43992f7fa25ce0316431dc39c5e18d5ad))

## [0.0.2](https://github.com/spectrocloud-labs/validator-plugin-vsphere/compare/v0.0.1...v0.0.2) (2023-10-02)


### Features

* Add support to validate available resources on cluster, hosts and resourcepools ([#26](https://github.com/spectrocloud-labs/validator-plugin-vsphere/issues/26)) ([d15e5a4](https://github.com/spectrocloud-labs/validator-plugin-vsphere/commit/d15e5a4a3ce7fc1bbe898dacff6f53388a9356ae))


### Bug Fixes

* better logging and missing fields in cmd/main.go ([#10](https://github.com/spectrocloud-labs/validator-plugin-vsphere/issues/10)) ([bb39b0b](https://github.com/spectrocloud-labs/validator-plugin-vsphere/commit/bb39b0b0a4d12cc6554041f86442e9115ba93889))
* **deps:** update kubernetes packages to v0.28.2 ([#28](https://github.com/spectrocloud-labs/validator-plugin-vsphere/issues/28)) ([cd84314](https://github.com/spectrocloud-labs/validator-plugin-vsphere/commit/cd84314cec33ac51d2f7a9f75ca851edfa50359b))
* **deps:** update module github.com/onsi/ginkgo to v2 ([#39](https://github.com/spectrocloud-labs/validator-plugin-vsphere/issues/39)) ([0251709](https://github.com/spectrocloud-labs/validator-plugin-vsphere/commit/025170979179cd839cf967a71cce29ee00961a61))
* **deps:** update module github.com/onsi/ginkgo/v2 to v2.12.1 ([#33](https://github.com/spectrocloud-labs/validator-plugin-vsphere/issues/33)) ([4303aed](https://github.com/spectrocloud-labs/validator-plugin-vsphere/commit/4303aed9d4c53c6eb764b39262b464480ee51874))
* **deps:** update module github.com/onsi/gomega to v1.27.10 ([#1](https://github.com/spectrocloud-labs/validator-plugin-vsphere/issues/1)) ([f0579a8](https://github.com/spectrocloud-labs/validator-plugin-vsphere/commit/f0579a804a165d4b568cb95e997cb315b70cfab5))
* **deps:** update module github.com/onsi/gomega to v1.28.0 ([#36](https://github.com/spectrocloud-labs/validator-plugin-vsphere/issues/36)) ([14b3f34](https://github.com/spectrocloud-labs/validator-plugin-vsphere/commit/14b3f3477f59ddd1684f088b79dee8ab12602347))
* **deps:** update module github.com/sirupsen/logrus to v1.9.3 ([#34](https://github.com/spectrocloud-labs/validator-plugin-vsphere/issues/34)) ([588e237](https://github.com/spectrocloud-labs/validator-plugin-vsphere/commit/588e2370111567e3548c038d098bbe7bfebf8cbd))
* **deps:** update module github.com/spectrocloud-labs/validator to v0.0.6 ([#7](https://github.com/spectrocloud-labs/validator-plugin-vsphere/issues/7)) ([ff931ed](https://github.com/spectrocloud-labs/validator-plugin-vsphere/commit/ff931edd2782e664149a6c51c67e4d2364489ef3))
* **deps:** update module github.com/spectrocloud-labs/validator to v0.0.8 ([#20](https://github.com/spectrocloud-labs/validator-plugin-vsphere/issues/20)) ([9c54342](https://github.com/spectrocloud-labs/validator-plugin-vsphere/commit/9c54342788a302ea591c630d272fbd7e2471d02a))
* **deps:** update module github.com/vmware/govmomi to v0.31.0 ([#37](https://github.com/spectrocloud-labs/validator-plugin-vsphere/issues/37)) ([530cca0](https://github.com/spectrocloud-labs/validator-plugin-vsphere/commit/530cca01ba680dff1207b3629a390a42cb33937f))
* **deps:** update module github.com/vmware/govmomi to v0.32.0 ([#40](https://github.com/spectrocloud-labs/validator-plugin-vsphere/issues/40)) ([e4d2478](https://github.com/spectrocloud-labs/validator-plugin-vsphere/commit/e4d2478e5d3be3fc382b0e197b588dee54a66b56))
* **deps:** update module sigs.k8s.io/controller-runtime to v0.16.2 ([#35](https://github.com/spectrocloud-labs/validator-plugin-vsphere/issues/35)) ([8327ed6](https://github.com/spectrocloud-labs/validator-plugin-vsphere/commit/8327ed6ec6446ad5c73f8c1cd24485ec687ea498))
* issues with updating validationresult and chart fixes ([#9](https://github.com/spectrocloud-labs/validator-plugin-vsphere/issues/9)) ([6cfbc56](https://github.com/spectrocloud-labs/validator-plugin-vsphere/commit/6cfbc569ae551da357593b2bb74a6d8f06838c43))


### Other

* Add charts ([#8](https://github.com/spectrocloud-labs/validator-plugin-vsphere/issues/8)) ([a0584bd](https://github.com/spectrocloud-labs/validator-plugin-vsphere/commit/a0584bd7e59ca2fadf5f7fd8d706fecfe928d5a5))
* add github workflows ([#11](https://github.com/spectrocloud-labs/validator-plugin-vsphere/issues/11)) ([fdc6b8f](https://github.com/spectrocloud-labs/validator-plugin-vsphere/commit/fdc6b8fb3f2682f58b52bf23eb2cc6f68aee0c59))
* add pre-commit config ([0e3cfb3](https://github.com/spectrocloud-labs/validator-plugin-vsphere/commit/0e3cfb3ed8760e76bdf8d68419d062be0c2d4b9b))
* Add release-please-config ([#29](https://github.com/spectrocloud-labs/validator-plugin-vsphere/issues/29)) ([27ca573](https://github.com/spectrocloud-labs/validator-plugin-vsphere/commit/27ca573fd3d5e8d526b75dc469b44149192b1c02))
* configure renovate ([6de3e6b](https://github.com/spectrocloud-labs/validator-plugin-vsphere/commit/6de3e6b713ca065b47268fe9e9e0c24bec044c51))
* **deps:** update actions/checkout action to v4 ([#17](https://github.com/spectrocloud-labs/validator-plugin-vsphere/issues/17)) ([fab282f](https://github.com/spectrocloud-labs/validator-plugin-vsphere/commit/fab282fa3d32ad7d1b42ba5417809da001be61b8))
* **deps:** update actions/checkout digest to f43a0e5 ([#14](https://github.com/spectrocloud-labs/validator-plugin-vsphere/issues/14)) ([f12efa7](https://github.com/spectrocloud-labs/validator-plugin-vsphere/commit/f12efa7108e25e4e524546e12926df33ae55484f))
* **deps:** update actions/setup-go digest to 93397be ([#15](https://github.com/spectrocloud-labs/validator-plugin-vsphere/issues/15)) ([3df0d01](https://github.com/spectrocloud-labs/validator-plugin-vsphere/commit/3df0d0104b7c4de8cd63e325e2edbf478dc90b22))
* **deps:** update actions/upload-artifact digest to a8a3f3a ([#18](https://github.com/spectrocloud-labs/validator-plugin-vsphere/issues/18)) ([1245a73](https://github.com/spectrocloud-labs/validator-plugin-vsphere/commit/1245a738ecb1c72e2b95bb60109429be087eddd5))
* **deps:** update docker/build-push-action action to v5 ([#22](https://github.com/spectrocloud-labs/validator-plugin-vsphere/issues/22)) ([f258548](https://github.com/spectrocloud-labs/validator-plugin-vsphere/commit/f25854898b324b0ef9bef50bd8062494740f054c))
* **deps:** update docker/build-push-action digest to 0a97817 ([#21](https://github.com/spectrocloud-labs/validator-plugin-vsphere/issues/21)) ([da721a1](https://github.com/spectrocloud-labs/validator-plugin-vsphere/commit/da721a117ae351db07126d2d73f9bcd520d49cc6))
* **deps:** update docker/login-action action to v3 ([#23](https://github.com/spectrocloud-labs/validator-plugin-vsphere/issues/23)) ([191a50e](https://github.com/spectrocloud-labs/validator-plugin-vsphere/commit/191a50e2d7c53dbe592139f11ec460e060a06362))
* **deps:** update docker/setup-buildx-action action to v3 ([#38](https://github.com/spectrocloud-labs/validator-plugin-vsphere/issues/38)) ([97446de](https://github.com/spectrocloud-labs/validator-plugin-vsphere/commit/97446dea7e18d4ca43e127ad17353ca5bfa38867))
* **deps:** update docker/setup-buildx-action digest to 885d146 ([#16](https://github.com/spectrocloud-labs/validator-plugin-vsphere/issues/16)) ([0dbadcc](https://github.com/spectrocloud-labs/validator-plugin-vsphere/commit/0dbadccbd2e52226022109be8f7cfa28fa948548))
* **deps:** update golang docker tag to v1.21 ([#2](https://github.com/spectrocloud-labs/validator-plugin-vsphere/issues/2)) ([212821b](https://github.com/spectrocloud-labs/validator-plugin-vsphere/commit/212821bc6443d68d252d4b60e93d4bbeb48b16d0))
* **deps:** update golang:1.21 docker digest to 19600fd ([#19](https://github.com/spectrocloud-labs/validator-plugin-vsphere/issues/19)) ([fde96bb](https://github.com/spectrocloud-labs/validator-plugin-vsphere/commit/fde96bbe6b54eea6e39d67392c1bfdcb6eb7bf57))
* Minor enhancements to tags and controller ([#32](https://github.com/spectrocloud-labs/validator-plugin-vsphere/issues/32)) ([6a79097](https://github.com/spectrocloud-labs/validator-plugin-vsphere/commit/6a79097d7e7102b1601d5181aad3e6bbd15c502a))
* refactor and move out functions from vsphere into respective v… ([#12](https://github.com/spectrocloud-labs/validator-plugin-vsphere/issues/12)) ([aee8f3d](https://github.com/spectrocloud-labs/validator-plugin-vsphere/commit/aee8f3d14bcf53f5a3e818d2011e41a3a05acd5d))
* release 0.0.1 ([#30](https://github.com/spectrocloud-labs/validator-plugin-vsphere/issues/30)) ([1ec57dc](https://github.com/spectrocloud-labs/validator-plugin-vsphere/commit/1ec57dc549e22b4f6f1bd71eb2e1ea0b9e196588))
* Remove unused RuleEngine struct ([#13](https://github.com/spectrocloud-labs/validator-plugin-vsphere/issues/13)) ([c3a1f95](https://github.com/spectrocloud-labs/validator-plugin-vsphere/commit/c3a1f95e111ed67e22b45d5ee164dda15734c533))
