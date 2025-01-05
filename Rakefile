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

  namespace :kafka do
    namespace :github do
      desc 'run kafka github consumer'
      task :consumer do
        system %{ go run -race cmd/githubconsumer/main.go }
        $CHILD_STATUS&.exitstatus || 1
      rescue Interrupt
        0
      end
    end
  end

  namespace :compose do
    desc 'run docker compose up'
    task :up do
      system %{ docker compose -f docker-compose.local.yml up }
      $CHILD_STATUS&.exitstatus || 1
    rescue Interrupt
      0
    end

    desc 'run docker compose down'
    task :down do
      system %{ docker compose -f docker-compose.local.yml down --remove-orphans }
      $CHILD_STATUS&.exitstatus || 1
    rescue Interrupt
      0
    end
  end
end
