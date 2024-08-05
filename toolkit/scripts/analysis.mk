# Copyright (c) Microsoft Corporation.
# Licensed under the MIT License.

# Contains:
#	- Generate list of built packages
#	- Run check for ABI changes of built packages.
#	- Run check for .so files version change of built packages.
#	- Validate package licenses

# Requires DNF on Azure Linux / yum and yum-utils on Ubuntu.

######## SODIFF and BUILD SUMMARY ########

# A folder with sodiff-related artifacts
SODIFF_OUTPUT_FOLDER=$(BUILD_DIR)/sodiff
RPM_BUILD_LOGS_DIR=$(LOGS_DIR)/pkggen/rpmbuilding
# A CSV file containing a list of "SRPM \t generated RPMs" entries
BUILD_SUMMARY_FILE=$(SODIFF_OUTPUT_FOLDER)/build-summary.csv
# A list of packages built during the current run
BUILT_PACKAGES_FILE=$(SODIFF_OUTPUT_FOLDER)/built-packages.txt
# Repositories that SODIFF runs the checks against
ifneq ($(build_arch),x86_64)
# Microsoft OSS repository only exists for x86_64 - skip that .repo file;
# otherwise package manager will signal an error due to being unable to make contact
SODIFF_REPO_SOURCES="azurelinux-official-base.repo"
else
SODIFF_REPO_SOURCES="azurelinux-official-base.repo azurelinux-ms-oss.repo"
endif

SODIFF_REPO_FILE=$(SCRIPTS_DIR)/sodiff/sodiff.repo
# An artifact containing a list of packages that need to be dash-rolled due to their dependency having a new .so version
SODIFF_SUMMARY_FILE=$(SODIFF_OUTPUT_FOLDER)/sodiff-summary.txt
# A script doing the sodiff work
SODIFF_SCRIPT=$(SCRIPTS_DIR)/sodiff/mariner-sodiff.sh

clean: clean-sodiff

clean-sodiff:
	rm -rf $(BUILD_SUMMARY_FILE)
	rm -rf $(BUILT_PACKAGES_FILE)
	rm -rf $(SODIFF_OUTPUT_FOLDER)
	rm -rf $(SODIFF_REPO_FILE)

.PHONY: built-packages-summary
built-packages-summary: $(BUILT_PACKAGES_FILE)

.PHONY: build-summary
build-summary: $(BUILD_SUMMARY_FILE)

# $(BUILT_PACKAGES_FILE): Generates a file containing a space-separated list of built RPM packages and subpackages.
$(BUILT_PACKAGES_FILE): $(BUILD_SUMMARY_FILE)
	cut -f2 --output-delimiter=" " $(BUILD_SUMMARY_FILE) > $(BUILT_PACKAGES_FILE)

# $(BUILD_SUMMARY_FILE): Generates a file containing 2 columns separated by a tab character:
# SRPM name and a space-separated list of RPM packages and subpackages generated by building the SRPM.
# Information is obtained from the build logs.
$(BUILD_SUMMARY_FILE): | $(RPM_BUILD_LOGS_DIR) $(SODIFF_OUTPUT_FOLDER)
	sed -nE -e 's#^.+level=info msg="Built \(([^\)]+)\) -> \[(.+)\].+#\1\t\2#p' $(RPM_BUILD_LOGS_DIR)/* > $(BUILD_SUMMARY_FILE)

$(RPM_BUILD_LOGS_DIR) $(SODIFF_OUTPUT_FOLDER):
	mkdir -p $@
	touch $@

# fake-built-packages-list: Generates a fake list of built packages by producing a file listing all present RPM files in the RPM directory.
.PHONY: fake-built-packages-list
fake-built-packages-list: | $(SODIFF_OUTPUT_FOLDER)
	touch $(RPM_BUILD_LOGS_DIR)
	touch $(BUILD_SUMMARY_FILE)
	find $(RPMS_DIR) -type f -name '*.rpm' -exec basename {} \; > $(BUILT_PACKAGES_FILE)

# sodiff-repo: Generate just the sodiff.repo file
.PHONY: sodiff-repo
sodiff-repo: $(SODIFF_REPO_FILE)

$(SODIFF_REPO_FILE):
	echo $(SODIFF_REPO_SOURCES) | sed -E 's:([^ ]+[.]repo):$(SPECS_DIR)/azurelinux-repos/\1:g' | xargs cat > $(SODIFF_REPO_FILE)

# sodiff-setup: populate gpg-keys from SPECS/azurelinux-repos for mariner official repos for ubuntu
.PHONY: sodiff-setup
sodiff-setup:
	mkdir -p /etc/pki/rpm-gpg
	cp $(SPECS_DIR)/azurelinux-repos/MICROSOFT-RPM-GPG-KEY /etc/pki/rpm-gpg/

# sodiff-check: runs check in a mariner container. Each failed package will be listed in $(SODIFF_OUTPUT_FOLDER).
.SILENT .PHONY: sodiff-check

sodiff-check: $(BUILT_PACKAGES_FILE) | $(SODIFF_REPO_FILE)
	<$(BUILT_PACKAGES_FILE) $(SODIFF_SCRIPT) $(RPMS_DIR)/ $(SODIFF_REPO_FILE) $(RELEASE_MAJOR_ID) $(SODIFF_OUTPUT_FOLDER)

package-toolkit: $(SODIFF_REPO_FILE)

######## LICENSE CHECK ########

license_check_build_dir   = $(BUILD_DIR)/license_check_tool
license_out_dir           = $(OUT_DIR)/license_check
license_results_file_pkg  = $(license_out_dir)/license_check_results_pkg.json
license_summary           = $(license_check_build_dir)/license_check_summary.txt

.PHONY: license-check license-check-pkg license-check-img clean-license-check

clean: clean-license-check
clean-license-check:
	@echo Verifying no mountpoints present in $(license_check_build_dir)
	$(SCRIPTS_DIR)/safeunmount.sh "$(license_check_build_dir)" && \
	rm -rf $(license_check_build_dir) && \
	rm -rf $(license_out_dir)

license_check_common_deps = $(go-licensecheck) $(chroot_worker) $(LICENSE_CHECK_EXCEPTION_FILE) $(LICENSE_CHECK_NAME_FILE) $(depend_LICENSE_CHECK_MODE)
# licensecheck-command: Helper function to run licensecheck with the given parameters.
# $(1): List of directories to check for licenses.
# $(2): (optional)Results .json file
# $(3): (optional)Results summary .txt file
# $(4): Log file

define licensecheck-command
	$(go-licensecheck) \
		$(foreach license_dir, $(1),--rpm-dirs="$(license_dir)" ) \
		--exception-file="$(LICENSE_CHECK_EXCEPTION_FILE)" \
		--name-file="$(LICENSE_CHECK_NAME_FILE)" \
		--worker-tar="$(chroot_worker)" \
		--build-dir="$(license_check_build_dir)" \
		--dist-tag=$(DIST_TAG) \
		--mode="$(LICENSE_CHECK_MODE)" \
		$(if $(2),--results-file="$(2)") \
		$(if $(3),--summary-file="$(3)") \
		--log-file=$(4) \
		--log-level=$(LOG_LEVEL)
endef

##help:target:license-check=Validate all packages in any of LICENSE_CHECK_DIRS for license compliance.
license-check: $(license_check_common_deps)
	$(if $(LICENSE_CHECK_DIRS),,$(error Must set LICENSE_CHECK_DIRS=))
	$(call licensecheck-command,$(LICENSE_CHECK_DIRS),$(license_results_file_pkg),$(license_summary),$(LOGS_DIR)/licensecheck/license-check-manual.log)

##help:target:license-check-pkg=Validate all packages in $(RPMS_DIR) for license compliance, building packages as needed.
license-check-pkg: $(license_check_common_deps) $(RPMS_DIR)
	$(call licensecheck-command,$(RPMS_DIR),$(license_results_file_pkg),$(license_summary),$(LOGS_DIR)/licensecheck/license-check-pkg.log)

##help:target:license-check-img=Validate all packages needed for an image for license compliance. Must set CONFIG_FILE=<path_to_config>.
license-check-img: $(license_results_file_img)
$(license_results_file_img): $(license_check_common_deps) $(image_package_cache_summary)
	$(call licensecheck-command,$(local_and_external_rpm_cache),$(license_results_file_img),$(license_summary),$(LOGS_DIR)/licensecheck/license-check-img.log)
