# frozen_string_literal: true

require 'English'

$LOAD_PATH << './scripts/local/rake'

Dir.glob('scripts/local/rake/**/*.rake').each { |r| import r }

DATABASE_NAME = ENV['DATABASE_NAME'] || nil
DATABASE_URL = ENV['DATABASE_URL'] || nil
DATABASE_URL_MIGRATION = ENV['DATABASE_URL_MIGRATION'] || nil
INFRA_POSTGRES_PASSWORD = ENV['POSTGRES_PASSWORD'] || nil

desc 'default task, runs webhookserver'
task default: ['run:webhookserver']
