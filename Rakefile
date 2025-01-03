# frozen_string_literal: true

require 'English'

desc 'default task, runs server'
task default: ['run:server']

namespace :run do
  desc 'run server'
  task :server do
    system %{ go run -race cmd/server/main.go }
    $CHILD_STATUS&.exitstatus || 1
  rescue Interrupt
    0
  end
end
