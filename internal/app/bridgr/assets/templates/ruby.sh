#!/bin/bash
mkdir -p /packages/gems
gem install builder
bundle package --all --no-install --cache-path=/packages/gems
# 'package' does not grab bundler, even when specified
gem fetch bundler
mv bundler*.gem /packages/gems/
gem generate_index -d /packages
