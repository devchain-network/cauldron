# frozen_string_literal: true

require 'English'

namespace :test do
  task :test_all do
    system %{ go test -failfast -v -coverprofile=coverage.out ./... }
    exit($CHILD_STATUS&.exitstatus || 1) unless ENV['RAKE_CONTINUE']
  rescue Interrupt
    exit(0)
  end

  desc 'run tests and show coverage'
  task :coverage do
    system %{ go test -v -coverprofile=coverage.out ./... && go tool cover -html=coverage.out }
    exit($CHILD_STATUS&.exitstatus || 1) unless ENV['RAKE_CONTINUE']
  rescue Interrupt
    exit(0)
  end
end

desc 'runs tests (shortcut)'
task test: 'test:test_all'
