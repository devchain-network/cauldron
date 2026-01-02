# frozen_string_literal: true

require 'English'

namespace :psql do

  desc 'connect to infra database with psql'
  task :infra do
    abort 'infra POSTGRES_PASSWORD environment variable is not set' if INFRA_POSTGRES_PASSWORD.nil?

    system %{
      PGPASSWORD="#{INFRA_POSTGRES_PASSWORD}" \
      PGOPTIONS="--search_path=cauldron,public" \
      psql -h localhost -p 5433 -U postgres -d #{DATABASE_NAME}
    }
    exit($CHILD_STATUS&.exitstatus || 1) unless ENV['RAKE_CONTINUE']
  rescue Interrupt
    exit(0)
  end
end
