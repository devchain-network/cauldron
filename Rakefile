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

  namespace :compose do

    namespace :kafka do
      desc 'run the kafka and kafka-ui only'
      task :up do
        system %{ docker compose -f docker-compose.kafka.yml up --build }
        $CHILD_STATUS&.exitstatus || 1
      rescue Interrupt
        0
      end

      desc 'stop the kafka and kafka-ui only'
      task :down do
        system %{ docker compose -f docker-compose.kafka.yml down --remove-orphans }
        $CHILD_STATUS&.exitstatus || 1
      rescue Interrupt
        0
      end
    end

    namespace :infra do
      desc 'run the infra with all components'
      task :up do
        system %{ docker compose -f docker-compose.infra.yml up --build }
        $CHILD_STATUS&.exitstatus || 1
      rescue Interrupt
        0
      end

      desc 'stop the infra with all components'
      task :down do
        system %{ docker compose -f docker-compose.infra.yml down --remove-orphans }
        $CHILD_STATUS&.exitstatus || 1
      rescue Interrupt
        0
      end
    end

  end
end

task :command_exists, [:command] do |_, args|
  if `command -v #{args.command} > /dev/null 2>&1 && echo $?`.chomp.empty?
    abort "'#{args.command}' command doesn't exists"
  end
end

task :has_rubocop do
  Rake::Task['command_exists'].invoke('rubocop')
end
task :has_go_migrate do
  Rake::Task['command_exists'].invoke('migrate')
end

DATABASE_NAME = ENV['DATABASE_NAME'] || nil
DATABASE_URL = ENV['DATABASE_URL'] || nil

task :pg_running do
  Rake::Task['command_exists'].invoke('pg_isready')
  abort 'DATABASE_NAME is not set' if DATABASE_NAME.nil?
  abort 'DATABASE_URL is not set' if DATABASE_URL.nil?
  abort 'postgresql is not running' if `$(command -v pg_isready) > /dev/null 2>&1 && echo $?`.chomp.empty?
end


namespace :db do
  desc 'init database'
  task init: %i[pg_running confirm] do
    unless `psql -Xqtl | cut -d \\| -f1 | grep -qw #{DATABASE_NAME} > /dev/null 2>&1 && echo $?`.chomp.empty?
      abort "#{DATABASE_NAME} is already exists"
    end
    system %{ createdb "#{DATABASE_NAME}" -O "#{ENV.fetch('USER', nil)}" && echo "#{DATABASE_NAME} was created." }
    $CHILD_STATUS&.exitstatus || 1
  rescue Interrupt
    0
  end

  desc 'run migrate up'
  task migrate: [:has_go_migrate] do
    system %{ migrate -database "#{DATABASE_URL}" -path "migrations" up }
    $CHILD_STATUS&.exitstatus || 1
  rescue Interrupt
    0
  end
end

namespace :rubocop do
  desc 'lint ruby'
  task lint: [:has_rubocop] do
    system %{ rubocop -F }
    $CHILD_STATUS&.exitstatus || 1
  rescue Interrupt
    0
  end

  desc 'lint ruby and autofix'
  task autofix: [:has_rubocop] do
    system %{ rubocop -A }
    $CHILD_STATUS&.exitstatus || 1
  rescue Interrupt
    0
  end
end

task :confirm do
  puts 'are you sure? [y,N]'
  input = $stdin.gets.chomp
  abort unless input.downcase == 'y'
end
