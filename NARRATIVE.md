# Bridgr

## Context

For many industries, there is a necessity to develop or deploy software systems on disconnected – or air-gapped – infrastructure. Banking, Energy, Government and Aviation are examples of industries that need such protection.  

The need usually comes from protecting data that resides on the air-gapped system, and thus the protection of the system itself is necessary. However, software and software development are constantly changing – so much so that keeping static air-gapped systems is a higher risk than the reward of newer software. We know then, custom and COTS software must be eventually updated and the recent rise of Agile and DevOps pushes those update cycles to an ever-faster iteration time.

In some cases, the needed artifacts are for updated Operating System packages, for others it’s shared code libraries for a particular language, still others need raw data files. The purpose varies as well: artifacts could be for ready-built software, or to facilitate custom development on the air-gapped network. Regardless of what type of data and for what purpose, a constant flow of artifacts is needed to sustain the air-gapped systems functionality for its users.

## Approaches

Without having tools that function better to support a DevOps pace of development and upgrades, most projects are forced to do the best with what they do have. On most projects I’ve participated in the implementation manifests as having a shared drive, web server or (worse?) squirreled away directories of artifacts from individual developers. For larger projects, there is likely a Change Control Board made of senior-level engineers and management who can’t reasonably be expected to understand nuances of dependent library versioning across the enterprise. The board then becomes a “rubber stamp” event for developers – or they simply circumvent the onerous process.

Now, with artifacts coming to the air-gapped network the challenge of managing them rears its head. Many projects use a “dumping ground” of artifacts on the shared drive with very little organization to it. Fewer use a breakdown of kinds of artifacts: JARs, Gems, ZIP files, etc. Fewer still have a dedicated artifact service available on those air-gapped networks. But even in this best-case scenario (an artifact management service) the crucial issue of knowledge and control of what is brought to the artifact service does not exist – at least there is no tool to simplify that activity. The antiquity of a control board also means antiquated forms of documentation: noting the artifact and version in a ticketing system, word document or the like means that there is a disconnect from the reality of what is fetched and transferred. Further, it puts more process on developers to reduce that disconnect and to do so in perpetuity – even when someone else requests a newer version of the same artifact. This lies in stark contrast to the pressures of Agile and DevOps to automate, develop faster, deliver increasing quality.

On smaller projects it is possible to focus on very specific needs to reduce the problem space and create a workable solution. For example, on a project that uses exclusively Python it is the most efficient use of developer or system administrator resources to simply clone the entire Pypi repository because “all of the items we could ever need are in there somewhere, so the problem is solved”. Unfortunately, that introduces immense security risk to the development environment and the air-gapped system, which is the very reason it was air-gapped to begin with! While the above approach may actually work for a single team, the reality is nearly any system of consequence uses many different kinds of artifacts. While the simplicity of cloning one repository may work for one team, it becomes untenable at system scale and for different kinds of artifacts.

The usage of COTS tools such as Artifactory and Nexus certainly help to solve this problem and present a useful interface for developers to consume the artifacts. These products do not specifically address all of the nuances for transferring and inputting artifacts of different types into an air-gapped instance. They may require further scripting to upload NPM packages compared to RPM packages, or they may not support producing only deltas from an internet-connected instance to be consumed by the air-gapped instance (thereby reducing transfer time). Additionally, the cost of these products is prohibitive for some projects when the only benefit is to serve artifacts on the air-gapped network. Most repository formats are known or documented, thus could reasonably be re-created on the air-gapped network. However, there appears to be a gap between these two markets: management doesn’t recognize the need to dedicate resources to custom-create and maintain repositories, and engineers don’t recognize value in purchasing those COTS products. What is needed is a simple, low-cost tool dedicated to performing one task well.

## Introducing Bridgr

The idea for Bridgr came out of the above-described situation. The Bridgr name comes from the idea that artifacts must find a way to Bridge the air-gap, and the tool is the activity of creating that “bridge”. Bridgr is an application written in Go (Golang) that allows governance of many different kinds of artifacts from internet-connected sources to air-gapped networks. It does this by fetching artifacts from specified sources and creating functional repositories – complete with metadata – that can be transferred to the air-gapped network and simply hosted on a static web server. The elegance of implementing with Go is that a single, statically-compiled binary can be easily produced for nearly any operating environment (Linux, Windows, Mac, ARM, etc) and consumed by users’ DevOps pipeline with minimal setup cost.

Because Bridgr came from the past experiences of working in situations as described above, I have ensured that Bridgr addresses those problems while functioning in the most common DevOps environments. A single Go executable was the first step, but to take it further, I have made the execution of Bridgr very configurable with a single YAML formatted file. This is in line with how many other DevOps-centric tooling works and is easy for developers to modify. This configuration file also defines the source all of the artifacts that Bridgr should fetch when executed. The benefit of this is that the “manifest” of artifacts needed can be kept in a simple, plain-text form and version controlled right alongside other code (such as a Git repository).

The architecture of the application is such that the inputs and the outputs can be easily developed to change or add support for more artifact types later. A plugin system was chosen to **not** be used since that would negate a major benefit of the single executable, however the architecture makes this as easy as possible. The input part of the architecture is a Go struct that knows how to read its appropriate section of the configuration file and get all of the sources and values normalized and ready to be handed to a worker. The main engine of Bridgr orchestrates creating the config structs and handing them to the appropriate worker. The worker is a struct that knows how to fetch sources defined in the config and output them into a subdirectory. For most workers the intended output is a type of repository that is useful to the users, such as YUM, RubyGem or PyPi. The worker knows how to create the metadata for these repository types and creates a fully functional repository with the artifacts that have been fetched. Finally, the engine takes these artifacts and sends them through any number of output structs. The output struct knows how to work with artifacts to get them to a final output location. The simplest output is to a local file system, but others can be supported, and multiple outputs may be selected by the user. One special output is a checksum: this output creates a checksum of every file written by the worker, then checksums the checksums of all subdirectories to arrive at a final, single checksum. This allows quick validation that the entire set of files are what the user expects.

Bridgr plans to support the following input types through the configuration file: YUM, RubyGem, Python, NPM, Docker, Jenkins Plugins, Maven, Git, Vagrant Box and File (File is any standalone file source – usually Zip or Tar files). For each type, the configuration file allows specifying (or not) the version of artifacts to be fetched (for example, the RubyGem `rails`, version `~5.1.0` which has specific meaning in RubyGems). Planned support for output types are: Local Filesystem, S3, and DVD/CD (aka, ISO images).

To accomplish supporting such a vast list of input and worker types, Bridgr cannot possibly (re)implement all these formats reliably and maintain them over time. Therefore, the main function of most workers is to craft a script that is fed to a docker container (the File worker is different in that it just uses Golang HTTP libraries to fetch artifacts). Within the docker container Bridgr utilizes the native tooling for creating repositories of various types. For example, the Ruby worker runs a docker image with ruby, and uses specific Gems that know how to create a RubyGem repository. Bridgr provides the knowledge of how to use these native tools and orchestrates their execution. Therefore, Docker is a required service that must be available on a system to successfully run Bridgr.

Because Bridgr is to be a trusted tool in bringing data to potentially sensitive systems and networks, the trust and validity of Bridgr must be paramount. I decided to create Bridgr as an Open Source development project on GitHub so that transparency as to the quality and, ultimately trust of the community can be as high as possible – otherwise Bridgr is of no use. Potential community contribution was also a significant factor in going Open Source. A final factor was the availability of DevOps service integrations for Open Source projects on GitHub for no cost. These have deep integrations into GitHub services, and are the standard to which many open source developers now expect. For Bridgr, I have set up the following integrations:

- Travis CI: this is the FOSS standard for CI engines. Its usage is similar to GitLab CI in that it is configured with a YAML file in the root of the project. Travis is triggered by commits to master, feature branches, PRs and tags, and can use different execution paths based on how it was triggered. Additionally, Travis automatically picks up and builds commits to PRs so that developers can be sure their fixes haven’t caused a regression.
- Todo: Code that is committed with `TODO:` in the text will automatically create a new issue in GitHub referencing the line number and text following TODO:
- WIP: Prefixing a PR title with `WIP` causes the PR to not be mergeable until the developer removes WIP. This is a useful GitLab feature that I miss in stock GitHub.
- Dependabot: This service daily checks the code repository for libraries that the application uses for vulnerabilities and creates a new issue and PR to address and fix the vulnerability. In the case of Bridgr, it reads the `go.mod` file to identify libraries being used but can support other languages to varying degrees of integration (on a JavaScript project I have, it can even make a PR with the fixed library version in package.json - all I have to do is click “Merge”).
- Codecov: This is a code coverage analysis tool, using the results from unit tests. Badges are available to put in the README so that users/developers can see right off what level of quality code to expect. Additional integration comes from evaluating PRs on each commit, and disallowing merging if there is a drop or regression in the overall coverage OR in the code modified by the PR.
- CodeScene: This is a tool that has been developed from the book [*Your code as a crime scene*](https://pragprog.com/book/atcrime/code-as-a-crime-scene). It uses metadata from the git repository to identify and prioritize hidden risks, suggest improvements and make early warnings of delivery risk before issues manifest. It is predictive behavioral analysis of how an application is being developed using temporal history rather than static code analysis (which is a snapshot in time). This is a unique and otherwise paid service that is free for FOSS projects. This provides product owner guidance for where to utilize resources most effectively and for developers to have situational awareness - such as where (in the code) to prioritize code reviews. See <https://empear.com/docs/CodeSceneUseCasesAndRoles.pdf> for much more detail.

The above project management and tooling is intended create as much transparency, developer accessibility and trust as possible for a tool that relies on these traits for acceptance.

## Future

With a new tool (Bridgr) available to solve certain artifact management problems across air-gapped systems, a new workflow becomes possible for projects. The Bridgr configuration file `bridge.yaml` should be kept in a Git repository, but can be separate or co-located with other code as appropriate. Changes to the configuration are now tracked by version control, and all of the benefits that provides: history, accountability, release management, and conflict resolution for artifacts. Security teams can now be involved in the process of dependency management, because Pull Requests for the configuration file can be reviewed by the security team prior to merging to mainline. It is even possible to make reviews by security mandatory before allowing merges. Furthering this DevSecOps interaction, the security team can actually code changes themselves: when a vulnerability is found, the security engineer can change the version of an artifact already in the configuration file and get concurrence from developers before merging into the baseline. The ease of YAML formatted files allows these interactions to be possible.

Once configuration is committed to the version control system, it should trigger a CI pipeline to begin. This pipeline will simply checkout the configuration file and execute Bridgr against it recreating all output files from what is given by the configuration. It’s recommended to start each pipeline with a clean output directory so that the results are only what is defined in Bridgr configuration – not cruft left from previous pipelines. If local filesystem is the target output, it is recommended to use something like the rsync tool to minimize transfer of artifacts to a final location. S3 output will be synchronized by Bridgr.

The second part of the pipeline should use a tool such as dependency-check to evaluate the files that have been downloaded by Bridgr against the National Vulnerability Database. I would suggest failing or alerting from the pipeline if vulnerabilities are found. Depending on the size of your configuration and source network connection, you may run this pipeline daily if it hasn’t been triggered by a source change – this would let you know immediately when new vulnerabilities exist in currently used artifacts.

Once artifacts are transferred to the air-gapped system, the checksum output that Bridgr makes can be used to validate that the air-gapped system contains exactly what was intended from the Bridgr configuration file. A single checksum string is the rollup of worker checksum files; thus, a single checksum value can be used to verify the entirety of Bridgr’s fileset in seconds. When the project is ready to host the collected artifacts on the air-gapped system or network, all that is required is a static web server. For example, a docker container running [Nginx](https://hub.docker.com/_/nginx) or even an S3-compatible service (AWS or [minio](https://min.io)) is sufficient to host the artifacts for all users. As updated artifacts come in, a simple rsync with delete option will quickly get the air-gapped hosting files updated. This minimizes time needed by infrastructure teams to handle transfer and updating of artifacts and the associated metadata.

Bridgr does not perform the actual transfer of artifacts. This step is too individualized for the industry, system and project to allow Bridgr to perform this action well. The literal transferring is left to another tool. However, Bridgr aims to be as helpful and compatible with how transfer tooling works as possible.

The time savings of Bridgr can be enormous: the labor hours required to manually check all libraries, their versions and their transitive dependencies can be crippling to a project. And keeping up with continuous changes is so impossible most projects just don’t. Bridgr reduces this effort to minutes of human attention. Further, Bridgr automates the action of downloading which occurs as fast as the computer can process.

Bridgr enables a much more Agile, modern and DevSecOps way of managing and transferring artifacts for air-gapped systems. In the spirit of automation, Bridgr provides a tool that automates tedious tasks that are not well suited for humans. Visibility into the libraries and artifacts being used by software systems can be attained and limited to strictly the artifacts truly needed by the system being built. Multiple Bridgr configuration files can be used when large projects desire to split the responsibilities of artifact management to the respective teams, but Bridgr also offers the flexibility of a single source of truth for artifacts.