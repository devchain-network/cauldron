# frozen_string_literal: true

task :command_exists, [:command] do |_, args|
  abort "#{args.command} doesn't exists" if `command -v #{args.command} > /dev/null 2>&1 && echo $?`.chomp.empty?
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


task :pg_running do
  Rake::Task['command_exists'].invoke('pg_isready')

  abort 'DATABASE_NAME is not set' if DATABASE_NAME.nil?
  abort 'DATABASE_URL is not set' if DATABASE_URL.nil?
  abort 'DATABASE_URL_MIGRATION is not set' if DATABASE_URL_MIGRATION.nil?
  abort 'postgresql is not running' if `$(command -v pg_isready) > /dev/null 2>&1 && echo $?`.chomp.empty?
end

task :confirm do
  puts 'are you sure? [y,N]'
  input = $stdin.gets.chomp
  abort unless input.downcase == 'y'
end
