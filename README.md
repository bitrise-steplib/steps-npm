# Run npm command

[![Step changelog](https://shields.io/github/v/release/bitrise-steplib/steps-npm?include_prereleases&label=changelog&color=blueviolet)](https://github.com/bitrise-steplib/steps-npm/releases)

The Step runs npm with the command and arguments you provide, for example, to install missing packages or run a package's test.

<details>
<summary>Description</summary>

You can install missing JS dependencies with this Step if you insert it before any build step and provide the `install` command. 
You can also test certain packages with the `test` command. 
You can do both in one Workflow, however, this requires one **Run npm command** Step for installation followed by another **Run npm command** Step for testing purposes.

### Configuring the Step
1. Add the **Run npm command** Step to your Workflow preceding any build Step.
2. Set the **Working directory**.
3. Set the command you want npm to execute, for example `install` to run `npm install` in the **The npm command with arguments to run** input.
4. If you're looking for a particular npm version, you can set it in the **Version of npm to use** input.
5. You can cache the content of the node modules directory if you select `true` in the drop-down. 
By default this input is set to false.

### Troubleshooting
Make sure you insert the Step before any build Step so that every dependency is downloaded a build Step starts running.

### Useful links
- [Getting started Ionic/Cordova apps](https://devcenter.bitrise.io/getting-started/getting-started-with-ionic-cordova-apps/)
- [About npm](https://www.npmjs.com/)
</details>

## 🧩 Get started

Add this step directly to your workflow in the [Bitrise Workflow Editor](https://devcenter.bitrise.io/steps-and-workflows/steps-and-workflows-index/).

You can also run this step directly with [Bitrise CLI](https://github.com/bitrise-io/bitrise).

## ⚙️ Configuration

<details>
<summary>Inputs</summary>

| Key | Description | Flags | Default |
| --- | --- | --- | --- |
| `workdir` | Working directory of the step. You can leave it empty to not change it.  |  | `$BITRISE_SOURCE_DIR` |
| `command` | Specify the command with arguments to run with `npm`.  This input value will be append to the end of the `npm` command call.  For example:  - `install` -> `npm install` - `install -g cordova` -> `npm install -g cordova` | required |  |
| `npm_version` | Set this value to the version of npm that is required to run the command. Must be a valid semver string. |  |  |
| `cache_local_deps` | Select if the contents of node_modules directory should be cached.  `true`: Mark local dependencies to be cached.  `false`: Do not use cache.  | required | `false` |
</details>

<details>
<summary>Outputs</summary>
There are no outputs defined in this step
</details>

## 🙋 Contributing

We welcome [pull requests](https://github.com/bitrise-steplib/steps-npm/pulls) and [issues](https://github.com/bitrise-steplib/steps-npm/issues) against this repository.

For pull requests, work on your changes in a forked repository and use the Bitrise CLI to [run step tests locally](https://devcenter.bitrise.io/bitrise-cli/run-your-first-build/).

Learn more about developing steps:

- [Create your own step](https://devcenter.bitrise.io/contributors/create-your-own-step/)
- [Testing your Step](https://devcenter.bitrise.io/contributors/testing-and-versioning-your-steps/)
