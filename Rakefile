# frozen_string_literal: true

require 'English'

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
task :has_golangci_linter do
  Rake::Task['command_exists'].invoke('golangci-lint')
end

desc 'default task, runs server'
task default: ['run:server']

namespace :run do
  desc 'run server'
  task :server do
    run = %{ go run -race cmd/server/main.go }
    pid = Process.spawn(run)
    Process.wait(pid)
    $CHILD_STATUS&.exitstatus || 1
  rescue Interrupt
    Process.getpgid(pid)
    Process.kill('KILL', pid)
    0
  end

  namespace :kafka do
    namespace :github do
      desc 'run kafka github consumer'
      task :consumer do
        run = %{ go run -race cmd/githubconsumer/main.go }
        pid = Process.spawn(run)
        Process.wait(pid)
        $CHILD_STATUS&.exitstatus || 1
      rescue Interrupt
        Process.getpgid(pid)
        Process.kill('KILL', pid)
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

    desc 'run migrator'
    task :migrator do
      system %{
        docker run \
          --env DATABASE_URL=${DATABASE_URL_DOCKER_TO_HOST} \
          devchain-migrator:latest
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

    desc 'build migrator'
    task :migrator do
      system %{ docker build -f Dockerfile.migrator -t devchain-migrator:latest . }
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

DATABASE_NAME = ENV['DATABASE_NAME'] || nil
DATABASE_URL = ENV['DATABASE_URL'] || nil
DATABASE_URL_MIGRATION = ENV['DATABASE_URL_MIGRATION'] || nil

task :pg_running do
  Rake::Task['command_exists'].invoke('pg_isready')
  abort 'DATABASE_NAME is not set' if DATABASE_NAME.nil?
  abort 'DATABASE_URL is not set' if DATABASE_URL.nil?
  abort 'DATABASE_URL_MIGRATION is not set' if DATABASE_URL_MIGRATION.nil?
  abort 'postgresql is not running' if `$(command -v pg_isready) > /dev/null 2>&1 && echo $?`.chomp.empty?
end


namespace :db do
  desc 'runs rake db:migrate up (shortcut)'
  task migrate: 'migrate:up'

  desc 'connect local db with psql'
  task psql: [:pg_running] do
    abort 'DATABASE_URL is not set' if DATABASE_URL.nil?
    system %{ PGOPTIONS="--search_path=cauldron" psql #{DATABASE_NAME} }
    $CHILD_STATUS&.exitstatus || 1
  rescue Interrupt
    0
  end

  desc 'reset database (drop and create)'
  task reset: %i[pg_running confirm] do
    system %{
      dropdb --if-exists "#{DATABASE_NAME}" &&
      createdb "#{DATABASE_NAME}" -O "${USER}" &&
      echo "#{DATABASE_NAME} was created"
    }
    $CHILD_STATUS&.exitstatus || 1
  rescue Interrupt
    0
  end

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

  namespace :migrate do
    desc 'run migrate up'
    task up: %i[pg_running has_go_migrate] do
      system %{ migrate -database "#{DATABASE_URL_MIGRATION}" -path "migrations" up }
      $CHILD_STATUS&.exitstatus || 1
    rescue Interrupt
      0
    end

    desc 'run migrate down'
    task down: %i[pg_running has_go_migrate] do
      system %{ migrate -database "#{DATABASE_URL_MIGRATION}" -path "migrations" down }
      $CHILD_STATUS&.exitstatus || 1
    rescue Interrupt
      0
    end

    desc 'go to migration'
    task :goto, [:index] => %i[pg_running has_go_migrate] do |_, args|
      abort 'DATABASE_URL_MIGRATION is not set' if DATABASE_URL_MIGRATION.nil?
      args.with_defaults(index: 0)

      index = begin
        Integer(args.index)
      rescue ArgumentError
        0
      end

      abort 'zero (0) is not a valid index' if index.zero?

      system %{ migrate -database "#{DATABASE_URL_MIGRATION}" -path "migrations" goto #{index} }
      $CHILD_STATUS&.exitstatus || 1
    rescue Interrupt
      0
    end
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

desc 'run golang-ci linter'
task lint: [:has_golangci_linter] do
  system %{ golangci-lint run }
  $CHILD_STATUS&.exitstatus || 1
rescue Interrupt
  0
end

namespace :test do
  task :test_all do
    system %{ go test -failfast -v -coverprofile=coverage.out ./... }
    $CHILD_STATUS&.exitstatus || 1
  rescue Interrupt
    0
  end

  desc 'run tests and show coverage'
  task :coverage do
    system %{ go test -v -coverprofile=coverage.out ./... && go tool cover -html=coverage.out }
    $CHILD_STATUS&.exitstatus || 1
  rescue Interrupt
    0
  end
end

desc 'runs tests (shortcut)'
task test: 'test:test_all'
