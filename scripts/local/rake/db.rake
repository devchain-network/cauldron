# frozen_string_literal: true

require 'English'

namespace :db do
  desc 'runs rake db:migrate up (shortcut)'
  task migrate: 'migrate:up'

  desc 'connect local db with psql'
  task psql: [:pg_running] do
    abort 'DATABASE_URL is not set' if DATABASE_URL.nil?
    system %{ PGOPTIONS="--search_path=cauldron" psql #{DATABASE_NAME} }

    exit($CHILD_STATUS&.exitstatus || 1) unless ENV['RAKE_CONTINUE']
  rescue Interrupt
    exit(0)
  end

  desc 'reset database (drop and create)'
  task reset: %i[pg_running confirm] do
    system %{
      dropdb --if-exists "#{DATABASE_NAME}" &&
      createdb "#{DATABASE_NAME}" -O "${USER}" &&
      echo "#{DATABASE_NAME} was created"
    }

    exit($CHILD_STATUS&.exitstatus || 1) unless ENV['RAKE_CONTINUE']
  rescue Interrupt
    exit(0)
  end

  desc 'init database'
  task init: %i[pg_running confirm] do
    unless `psql -Xqtl | cut -d \\| -f1 | grep -qw #{DATABASE_NAME} > /dev/null 2>&1 && echo $?`.chomp.empty?
      abort "#{DATABASE_NAME} is already exists"
    end

    system %{ createdb "#{DATABASE_NAME}" -O "#{ENV.fetch('USER', nil)}" && echo "#{DATABASE_NAME} was created." }
    exit($CHILD_STATUS&.exitstatus || 1) unless ENV['RAKE_CONTINUE']
  rescue Interrupt
    exit(0)
  end

  namespace :migrate do

    desc 'run migrate up'
    task up: %i[pg_running has_go_migrate] do
      system %{ migrate -database "#{DATABASE_URL_MIGRATION}" -path "migrations" up }
      exit($CHILD_STATUS&.exitstatus || 1) unless ENV['RAKE_CONTINUE']
    rescue Interrupt
      exit(0)
    end

    desc 'run migrate down'
    task down: %i[pg_running has_go_migrate] do
      system %{ migrate -database "#{DATABASE_URL_MIGRATION}" -path "migrations" down }
      exit($CHILD_STATUS&.exitstatus || 1) unless ENV['RAKE_CONTINUE']
    rescue Interrupt
      exit(0)
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
      exit($CHILD_STATUS&.exitstatus || 1) unless ENV['RAKE_CONTINUE']
    rescue Interrupt
      exit(0)
    end
  end
end
