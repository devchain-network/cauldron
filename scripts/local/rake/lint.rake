# frozen_string_literal: true

require 'English'

namespace :lint do

  desc 'run golang-ci linter'
  task go: [:has_golangci_linter] do
    system %{ golangci-lint run }
    exit($CHILD_STATUS&.exitstatus || 1) unless ENV['RAKE_CONTINUE']
  rescue Interrupt
    exit(0)
  end

  desc 'lint ruby'
  task ruby: [:has_rubocop] do
    system %{ rubocop -F }
    exit($CHILD_STATUS&.exitstatus || 1) unless ENV['RAKE_CONTINUE']
  rescue Interrupt
    exit(0)
  end

  desc 'lint ruby and autofix'
  task 'ruby:autofix' => [:has_rubocop] do
    system %{ rubocop -A }
    exit($CHILD_STATUS&.exitstatus || 1) unless ENV['RAKE_CONTINUE']
  rescue Interrupt
    exit(0)
  end

end
