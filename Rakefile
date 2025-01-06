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

namespace :docker do
  namespace :run do
    desc 'run server'
    task :server do
      system %{
        docker run \
          --env GITHUB_HMAC_SECRET=${GITHUB_HMAC_SECRET} \
          -p 8000:8000 \
          devchain-server:latest
      }
      $CHILD_STATUS&.exitstatus || 1
    rescue Interrupt
      0
    end

    desc 'run github consumer'
    task :github_consumer do
      system %{
        docker run \
          --env KC_TOPIC=${KC_TOPIC} \
          devchain-gh-consumer:latest
      }
      $CHILD_STATUS&.exitstatus || 1
    rescue Interrupt
      0
    end
  end
  namespace :build do
    desc 'build server'
    task :server do
      system %{ docker build -f Dockerfile.server -t devchain-server:latest . }
      $CHILD_STATUS&.exitstatus || 1
    rescue Interrupt
      0
    end

    desc 'build github consumer'
    task :github_consumer do
      system %{ docker build -f Dockerfile.github-consumer -t devchain-gh-consumer:latest . }
      $CHILD_STATUS&.exitstatus || 1
    rescue Interrupt
      0
    end
  end
end
